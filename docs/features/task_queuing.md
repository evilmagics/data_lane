# Task Queuing & Processing

## Overview
The PDF Generator uses a persistent task queue to handle PDF generation jobs asynchronously. This ensures reliability, allows for retries, and manages server load by limiting concurrent executions.

## Architecture
1.  **Request**: Client sends a POST request to `/api/queue` with task metadata.
2.  **Persistence**: The task is immediately stored in the primary `sqlite` database with status `queued`. Key fields are extracted for efficient querying:
    *   `root_folder`: Source data root path
    *   `gate_id`: Target gate ID
    *   `station_id`: Target station ID
    *   `filter_json`: JSON serialized filter configuration
3.  **Enqueue**: The task ID and metadata are pushed to the `backlite` queue (backed by `sqlite`).
4.  **Worker**: A background worker (running in the same process) picks up the task.
5.  **Processing**:
    *   Status updated to `running`.
    *   **Progress tracking**: The task reports progress with stage info (`loading_data`, `generating`, `saving`) and counts (current/total transactions).
    *   PDF is generated using `maroto` with a detailed layout (including transaction images, custom fonts, and verified Access DB data).
    *   Output file is saved to `output/` directory as an **absolute path**.
6.  **Completion**:
    *   On success: Status updated to `completed`, output details saved.
    *   On failure: Status updated to `failed`, **error message stored in `error_message` field**. Retries may occur based on configuration.

## Progress Tracking
*   **Real-time Updates**: Progress is pushed via SSE endpoint `/sse/tasks/:id` every **1 second**.
*   **Progress Data**:
    ```json
    {
      "status": "running",
      "progress_stage": "generating",
      "progress_current": 150,
      "progress_total": 1000,
      "error_message": ""
    }
    ```
*   **Stages** (human-readable descriptions):
    *   `Initializing settings`: Task starting, loading configuration
    *   `Connecting to database`: Opening Access database connection
    *   `Loading fonts`: Loading custom PDF fonts
    *   `Building PDF header`: Creating document header section
    *   `Appending transaction X of Y`: Processing each transaction row (progress_current increments)
    *   `All transactions appended`: All transaction rows have been added to the document
    *   `Rendering PDF document`: Generating the final PDF
    *   `Writing file to disk`: Saving the PDF to output directory
    *   `Completed`: Done

## Queue Configuration
*   **Concurrency**: Controlled by the `queue_concurrency` setting (default: 1).
*   **Persistence**: Tasks are stored in the application database (`d:\Projects\intracs\pdf_generator\app.db` by default). Queue tables are automatically created/updated on startup.
*   **Retries**: Configured to retry 3 times with exponential backoff.
*   **Database Locking**: SQLite WAL mode is enabled with a 5-second busy timeout to handle concurrent access between the API and background workers.

## Error Handling
*   When a task fails, the `error_message` field contains the detailed error description.
*   This is visible via the task detail API (`GET /api/tasks/:id`) and SSE stream.

## Monitoring
*   **API**: `GET /api/queue/stats` (if implemented) or via `GET /api/tasks`.
*   **SSE**: Real-time updates are pushed to the `/sse/tasks/:id` stream when task status or progress changes.

