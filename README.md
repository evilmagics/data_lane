# PDF Generator Service

A robust background service for generating PDFs from transaction data stored in Microsoft Access databases. Features include a job queue, scheduling, and JSON API.

## Features

- **HTTP API**: Built with Fiber v3 for high performance.
- **Job Queue**: Durable SQLite-backed queue using `goqite` (Turso compatible).
- **Scheduling**: Cron-based scheduling for recurring PDF generation.
- **Database**: 
  - **Turso (LibSQL)**: Stores tasks, schedules, and job queue.
  - **Microsoft Access via ODBC**: Source of transaction data.
- **Real-time Progress**: SSE endpoint for tracking task generation progress.

## Prerequisites

- Go 1.25+
- ODBC Driver for Microsoft Access
- Turso Database (or local SQLite file)

## Configuration

Copy `.env.example` to `.env` and configure:

```ini
TURSO_DATABASE_URL=libsql://...
TURSO_AUTH_TOKEN=...
ACCESS_DB_ROOT_FOLDER=D:/Path/To/Root
OUTPUT_PATH=./output
```

## Running

```bash
# Install dependencies
go mod tidy

# Run server
go run cmd/main.go
```

## API Usage

See [API.md](API.md) for full documentation.

### Quick Start

**Enqueue a Task:**

```bash
curl -X POST http://localhost:3000/api/tasks/enqueue \
  -H "Content-Type: application/json" \
  -d '{
    "root_folder": "D:/Root",
    "station": "AMPLAS",
    "date_type": "single",
    "single_date": "2025-12-01"
  }'
```

**Stream Progress:**

```bash
curl -N http://localhost:3000/api/tasks/{task_id}/progress
```
