# DataLane - PDF Verification & Generation Service

DataLane is a robust, API-first system designed to automate the generation of verification PDF reports from transaction data stored in legacy Microsoft Access databases. It features a modern web interface, a durable job queue, distributed service architecture, and comprehensive scheduling capabilities.

## 🚀 Features

*   **Distributed Architecture**: Separation of concerns with a central API server (`datalane`) and a dedicated background worker service (`pdf-generator`).
*   **Legacy Integration**: Seamlessly reads data from Microsoft Access (`.mdb`) files using ODBC.
*   **Robust Job Queue**: SQLite-backed persistent queue (using `backlite`) ensures no tasks are lost, even during restarts.
*   **High Performance**: Built on Go and Fiber v3 for low-latency API handling.
*   **PDF Generation**: dynamic PDF creation using `maroto`, capable of handling large datasets and custom templates.
*   **Modern UI**: Embedded Next.js dashboard for managing tasks, schedules, stations, and settings.
*   **Scheduling**: Cron-based scheduler for automated recurring reports.
*   **Security**: Role-based access control, API Key support, and HMAC signature verification for critical operations.
*   **Cross-Platform**: Runs on Windows (primary target for Access/ODBC support), Linux, and macOS.

## 🛠️ Architecture Overview

The system consists of two main executables:

1.  **DataLane API (`datalane.exe`)**
    *   Host the HTTP REST API.
    *   Serves the embedded Next.js Frontend.
    *   Manages the SQLite database and Queue (Producer).
    *   Controls the Worker Service (Start/Stop/Status).
    *   Handles Authentication and Scheduling.

2.  **PDF Worker (`pdf-generator.exe`)**
    *   Runs as a background system service (Daemon).
    *   Consumes tasks from the Queue.
    *   Performs the heavy lifting of PDF generation.
    *   Connects directly to the Access Database.

## 📦 Modules & Structure

*   `cmd/`
    *   `main.go`: Entry point for the main API server.
    *   `pdf-generator/`: Entry point for the background worker service.
*   `internal/`
    *   `core/`: Domain logic, ports, and services (Hexagonal Architecture).
    *   `adapters/`: Implementation of ports (HTTP Handlers, Repositories, Middleware).
    *   `server/`: Fiber app configuration and route setup.
    *   `assets/`: Embedded UI assets.
*   `pkg/`
    *   `queue/`: Durable queue implementation wrapping `backlite`.
    *   `generator/`: PDF generation logic using `maroto`.
    *   `datasource/`: ODBC connection handling for Access databases.
    *   `database/`: SQLite connection and GORM setup.
    *   `config/`: Configuration loading.
*   `ui/`: Next.js frontend application.

## ⚙️ Installation & Setup

### Prerequisites
*   **Go** 1.25 or higher
*   **Node.js** 18+ (for building UI)
*   **GCC** (required for `go-sqlite3` / CGO)
*   **Microsoft Access ODBC Driver** (if running on Windows and connecting to real MDB files)

### 1. Build the Project

```bash
# 1. Build the UI
cd ui
npm install
npm run build
cd ..

# 2. Build the API Server (embeds the UI)
go build -o output/datalane.exe ./cmd/main.go

# 3. Build the Worker Service
go build -o output/pdf-generator.exe ./cmd/pdf-generator/main.go
```

### 2. Configuration

Create a `.env` file in the `output/` directory (or wherever you run the binaries):

```ini
# Server Configuration
SERVER_PORT=3000
SERVER_HOST=0.0.0.0
SERVER_URL=http://localhost:3000

# Database
DB_PATH=data/datalane.db

# Access Database Source
ACCESS_DB_ROOT_FOLDER=D:/Projects/Data

# Session
SESSION_EXPIRY=12h
SESSION_MAX_CONCURRENT=5

# Logging configuration (managed via UI/DB mostly)
```

### 3. Running the System

**Start the Main API:**

```bash
./output/datalane.exe
```
Access the dashboard at `http://localhost:3000`.

**Install & Start the Worker Service:**

You can manage the worker service via the API/Dashboard or manually:

```bash
# Manual installation (Administrator terminal)
./output/pdf-generator.exe -service install
./output/pdf-generator.exe -service start
```

## 📚 API Documentation

Full API specification is available in `docs/api_spec.md`.

### Key Endpoints

*   `POST /api/auth/login`: Authenticate and get session.
*   `POST /api/queue`: Enqueue a new PDF generation task.
*   `GET /api/tasks`: List tasks and their status.
*   `GET /api/system/services/pdf-generator/status`: Check worker status.
*   `GET /sse/events`: Real-time updates via Server-Sent Events.

## 🧪 Testing

```bash
go test ./...
```

## 📜 License

Internal Project. All rights reserved.
