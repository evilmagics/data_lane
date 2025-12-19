# PDF Generator Architecture

## Design Pattern
**Modular Monolith** with **Clean Architecture** principles.

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Fiber)                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐   │
│  │  Handlers   │ │ Middleware  │ │    SSE Hub          │   │
│  └──────┬──────┘ └──────┬──────┘ └──────────┬──────────┘   │
└─────────┼───────────────┼───────────────────┼───────────────┘
          │               │                   │
          ▼               ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │  Auth    │ │  Queue   │ │ Schedule │ │ PDF Generator│   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────┬───────┘   │
└───────┼────────────┼────────────┼──────────────┼────────────┘
        │            │            │              │
        ▼            ▼            ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Session  │ │   Task   │ │ Schedule │ │   Settings   │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────┬───────┘   │
└───────┼────────────┼────────────┼──────────────┼────────────┘
        │            │            │              │
        ▼            ▼            ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Database (SQLite + GORM)                  │
└─────────────────────────────────────────────────────────────┘
```

## Folder Structure
```
pdf_generator/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── docs/
│   ├── api_spec.md              # API specification
│   └── architecture.md          # This file
├── internal/
│   ├── core/
│   │   ├── domain/              # Domain models (entities)
│   │   │   ├── task.go
│   │   │   ├── schedule.go
│   │   │   ├── settings.go
│   │   │   ├── session.go
│   │   │   ├── apikey.go
│   │   │   └── log.go
│   │   ├── ports/               # Interfaces (Repository, Service)
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── services/            # Business logic
│   │       ├── auth_service.go
│   │       ├── queue_service.go
│   │       ├── schedule_service.go
│   │       ├── settings_service.go
│   │       └── pdf_service.go
│   ├── adapters/
│   │   ├── handlers/            # HTTP handlers
│   │   │   ├── auth_handler.go
│   │   │   ├── task_handler.go
│   │   │   ├── schedule_handler.go
│   │   │   ├── settings_handler.go
│   │   │   ├── apikey_handler.go
│   │   │   └── sse_handler.go
│   │   ├── middleware/          # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── hmac.go
│   │   │   └── logger.go
│   │   ├── repository/          # Database implementations
│   │   │   ├── task_repo.go
│   │   │   ├── schedule_repo.go
│   │   │   ├── settings_repo.go
│   │   │   ├── session_repo.go
│   │   │   ├── apikey_repo.go
│   │   │   └── log_repo.go
│   │   └── scheduler/           # Gocron implementation
│   │       └── scheduler.go
│   ├── config/                  # Configuration loading
│   │   └── config.go
│   └── server/                  # HTTP server setup
│       ├── server.go
│       └── routes.go
├── pkg/
│   ├── database/                # Database initialization
│   │   └── sqlite.go
│   ├── logger/                  # Zerolog wrapper
│   │   └── logger.go
│   ├── queue/                   # Backlite wrapper
│   │   └── queue.go
│   ├── generator/               # PDF generation
│   │   └── generator.go
│   ├── datasource/              # ADODB Access connector
│   │   └── access.go
│   ├── utils/                   # Helpers
│   │   ├── crypto.go            # AES encryption, HMAC
│   │   ├── jwt.go               # JWT utilities
│   │   └── response.go          # Standard API response
│   └── api/                     # API types
│       ├── request.go
│       └── response.go
├── ui/                          # Frontend assets (embedded)
│   ├── index.html
│   ├── css/
│   └── js/
├── output/                      # Generated PDFs (gitignored)
├── logs/                        # Application logs (gitignored)
├── deprecated/                  # Old implementation (reference only)
├── .env
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## Database Schema

### Tables

#### `settings`
| Column        | Type | Constraints | Description        |
| ------------- | ---- | ----------- | ------------------ |
| `key`         | TEXT | PRIMARY KEY | Setting identifier |
| `value`       | TEXT | NOT NULL    | Setting value      |
| `description` | TEXT |             | Human description  |

#### `sessions`
| Column       | Type     | Constraints  | Description   |
| ------------ | -------- | ------------ | ------------- |
| `id`         | TEXT     | PRIMARY KEY  | UUID          |
| `user_id`    | TEXT     | NOT NULL     | "admin"       |
| `token_hash` | TEXT     | UNIQUE INDEX | SHA256 of JWT |
| `ip_address` | TEXT     |              | Client IP     |
| `user_agent` | TEXT     |              | Browser info  |
| `expires_at` | DATETIME | INDEX        | Expiration    |
| `created_at` | DATETIME |              | Login time    |

#### `tasks`
| Column             | Type     | Constraints | Description                                     |
| ------------------ | -------- | ----------- | ----------------------------------------------- |
| `id`               | TEXT     | PRIMARY KEY | UUID                                            |
| `schedule_id`      | TEXT     | NULLABLE FK | Link to schedule                                |
| `status`           | TEXT     | INDEX       | queued/pending/running/completed/failed/removed |
| `metadata`         | TEXT     |             | JSON payload                                    |
| `output_file_path` | TEXT     |             | Relative path                                   |
| `output_file_size` | INTEGER  |             | Bytes                                           |
| `created_at`       | DATETIME |             |                                                 |
| `updated_at`       | DATETIME |             |                                                 |

#### `schedules`
| Column         | Type     | Constraints  | Description     |
| -------------- | -------- | ------------ | --------------- |
| `id`           | TEXT     | PRIMARY KEY  | UUID            |
| `cron`         | TEXT     | NOT NULL     | Cron expression |
| `task_payload` | TEXT     | NOT NULL     | JSON template   |
| `active`       | BOOLEAN  | DEFAULT true |                 |
| `last_run`     | DATETIME |              |                 |
| `next_run`     | DATETIME |              |                 |
| `created_at`   | DATETIME |              |                 |
| `updated_at`   | DATETIME |              |                 |

#### `api_keys`
| Column          | Type     | Constraints  | Description                |
| --------------- | -------- | ------------ | -------------------------- |
| `id`            | TEXT     | PRIMARY KEY  | UUID                       |
| `name`          | TEXT     | NOT NULL     | Display name               |
| `key_hash`      | TEXT     | UNIQUE       | Bcrypt hash                |
| `encrypted_key` | TEXT     |              | AES encrypted (for reveal) |
| `active`        | BOOLEAN  | DEFAULT true |                            |
| `created_at`    | DATETIME |              |                            |

#### `logs`
| Column       | Type     | Constraints               | Description     |
| ------------ | -------- | ------------------------- | --------------- |
| `id`         | INTEGER  | PRIMARY KEY AUTOINCREMENT |                 |
| `level`      | TEXT     | INDEX                     | info/warn/error |
| `message`    | TEXT     |                           | Log message     |
| `context`    | TEXT     |                           | JSON context    |
| `created_at` | DATETIME | INDEX                     |                 |

## Key Design Decisions

1. **SQLite**: Simple, file-based, perfect for single-server deployment.
2. **Backlite**: SQLite-backed queue for persistence across restarts.
3. **JWT + Session**: Hybrid approach - JWT for stateless auth, Session table for revocation/limits.
4. **AES Encryption**: API keys encrypted (not just hashed) to allow admin to reveal them.
5. **SSE over WebSocket**: Simpler for unidirectional server-to-client events.
