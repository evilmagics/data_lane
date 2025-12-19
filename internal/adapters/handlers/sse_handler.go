package handlers

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"

	"pdf_generator/internal/core/ports"
)

// SSEHandler handles Server-Sent Events
type SSEHandler struct {
	taskRepo ports.TaskRepository
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(taskRepo ports.TaskRepository) *SSEHandler {
	return &SSEHandler{taskRepo: taskRepo}
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

// TaskEvents handles GET /sse/tasks/:id (Shared)
func (h *SSEHandler) TaskEvents(c fiber.Ctx) error {
	taskID := c.Params("id")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.SendStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastStatus string
		
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

				// Only send if status changed
				if string(task.Status) != lastStatus {
					data := fmt.Sprintf(`{"status":"%s","output_size":%d}`, task.Status, task.OutputFileSize)
					fmt.Fprintf(w, "event: status\n")
					fmt.Fprintf(w, "data: %s\n\n", data)
					if err := w.Flush(); err != nil {
						return
					}
					
					lastStatus = string(task.Status)

					// Stop streaming if task is terminal
					if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
						return
					}
				}
			}
		}
	}))

	return nil
}
