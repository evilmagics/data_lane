# Task Queuing & Processing

## Overview
The PDF Generator uses a persistent task queue to handle PDF generation jobs asynchronously. This ensures reliability, allows for retries, and manages server load by limiting concurrent executions.

## Architecture
1.  **Request**: Client sends a POST request to `/api/queue` with task metadata.
2.  **Persistence**: The task is immediately stored in the primary `sqlite` database with status `queued`.
3.  **Enqueue**: The task ID and metadata are pushed to the `backlite` queue (backed by `sqlite`).
4.  **Worker**: A background worker (running in the same process) picks up the task.
5.  **Processing**:
    *   Status updated to `running`.
    *   PDF is generated using `maroto` and Access DB data.
    *   Output file is saved to `output/` directory.
6.  **Completion**:
    *   On success: Status updated to `completed`, output details saved.
    *   On failure: Status updated to `failed`, error logged. Retries may occur based on configuration.

## Queue Configuration
*   **Concurrency**: Controlled by the `queue_concurrency` setting (default: 1).
*   **Persistence**: Tasks are stored in the application database (`d:\Projects\intracs\pdf_generator\app.db` by default).
*   **Retries**: Configured to retry 3 times with exponential backoff.
*   **Database Locking**: SQLite WAL mode is enabled with a 5-second busy timeout to handle concurrent access between the API and background workers.

## Monitoring
*   **API**: `GET /api/queue/stats` (if implemented) or via `GET /api/tasks`.
*   **SSE**: Real-time updates are pushed to the `/sse/events` stream when task status changes.
