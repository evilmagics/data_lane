package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"

	"pdf_generator/internal/core/ports"
)

// SSEHandler handles Server-Sent Events
type SSEHandler struct {
	taskRepo ports.TaskRepository
	queue    ports.QueueService
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(taskRepo ports.TaskRepository, queue ports.QueueService) *SSEHandler {
	return &SSEHandler{
		taskRepo: taskRepo,
		queue:    queue,
	}
}

// GlobalEvents handles GET /sse/events (Admin only)
func (h *SSEHandler) GlobalEvents(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.SendStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			// Check if client disconnected by writing a ping or just checking error on Flush?
			// fasthttp loop ends when this function returns.
			// We block here.

			select {
			case <-ticker.C:
				// Use Background context because c.Context() is invalid here
				ctx := context.Background()
				
				// Send global stats
				// Error handling ignored for brevity in stream
				queueSize, _ := h.taskRepo.CountByStatus(ctx, "queued")
				runningCount, _ := h.taskRepo.CountByStatus(ctx, "running")

				data := fmt.Sprintf(`{"queue_size":%d,"running_tasks":%d}`, queueSize, runningCount)
				
				_, err := fmt.Fprintf(w, "event: global_stats\n")
				if err != nil {
					return
				}
				_, err = fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					return
				}
				
				if err := w.Flush(); err != nil {
					return // Client disconnected
				}
			}
		}
	}))

	return nil
}

// TaskProgressEvent represents the SSE event data for task progress
type TaskProgressEvent struct {
	Status          string `json:"status"`
	OutputSize      int64  `json:"output_size"`
	ErrorMessage    string `json:"error_message,omitempty"`
	ProgressStage   string `json:"progress_stage,omitempty"`
	ProgressCurrent int    `json:"progress_current"`
	ProgressTotal   int    `json:"progress_total"`
}

// TaskEvents handles GET /sse/tasks/:id (Shared)
// Pushes progress updates every 1 second
func (h *SSEHandler) TaskEvents(c fiber.Ctx) error {
	taskID := c.Params("id")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.SendStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second) // Push every 1 second
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				task, err := h.taskRepo.GetByID(ctx, taskID)
				if err != nil {
					fmt.Fprintf(w, "event: error\n")
					fmt.Fprintf(w, "data: {\"error\":\"Task not found\"}\n\n")
					w.Flush()
					return
				}

				// Get real-time progress from queue if available
				var progressStage string
				var progressCurrent, progressTotal int
				
				if h.queue != nil {
					progress := h.queue.GetProgress(taskID)
					if progress != nil {
						progressStage = progress.Stage
						progressCurrent = progress.Current
						progressTotal = progress.Total
					} else {
						// Fall back to DB values
						progressStage = task.ProgressStage
						progressCurrent = task.ProgressCurrent
						progressTotal = task.ProgressTotal
					}
				} else {
					progressStage = task.ProgressStage
					progressCurrent = task.ProgressCurrent
					progressTotal = task.ProgressTotal
				}

				// Build event data
				eventData := TaskProgressEvent{
					Status:          string(task.Status),
					OutputSize:      task.OutputFileSize,
					ErrorMessage:    task.ErrorMessage,
					ProgressStage:   progressStage,
					ProgressCurrent: progressCurrent,
					ProgressTotal:   progressTotal,
				}

				eventJSON, _ := json.Marshal(eventData)

				// Always send progress every second
				fmt.Fprintf(w, "event: progress\n")
				fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
				if err := w.Flush(); err != nil {
					return
				}

				// Stop streaming if task is terminal
				if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
					return
				}
			}
		}
	}))

	return nil
}
