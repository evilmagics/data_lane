# PDF Generator API Specification

**Version**: 1.0  
**Base URL**: `/api`

### UI Configuration
The embedded UI requires `/config.js` to be loaded to provision runtime configuration:
* `API_BASE_URL`: Defines the API endpoint location.

---

## Global Standards

### Authentication Headers
| Header          | Description                                                         | Required                     |
| --------------- | ------------------------------------------------------------------- | ---------------------------- |
| `Authorization` | `Bearer <JWT>` for Admin access                                     | For Admin endpoints          |
| `X-API-Key`     | API Key for External access                                         | For Shared endpoints         |
| `X-Signature`   | HMAC-SHA256 signature of request body (Signed using API Key or JWT) | For POST/PUT/PATCH with body |

### Standard Response Format
```json
{
  "code": 0,
  "message": "OK",
  "data": {}
}
```

| Field     | Type                  | Description                      |
| --------- | --------------------- | -------------------------------- |
| `code`    | `int`                 | `0` = Success, `>0` = Error code |
| `message` | `string`              | Human-readable message           |
| `data`    | `object\|array\|null` | Response payload                 |

### Error Codes
| Code | Description                 |
| ---- | --------------------------- |
| 0    | Success                     |
| 1001 | Invalid request body        |
| 1002 | Validation error            |
| 2001 | Unauthorized                |
| 2002 | Invalid credentials         |
| 2003 | Session limit reached       |
| 2004 | Token expired               |
| 2005 | Invalid API Key             |
| 2006 | HMAC signature mismatch     |
| 3001 | Resource not found          |
| 3002 | Task not ready for download |
| 5001 | Internal server error       |

---

## Endpoints

### A. Authentication

#### 1. Login
**POST** `/auth/login`  
**Access**: Public

**Request Body**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response** (`data`):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-16T12:00:00Z",
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

#### 2. Logout
**POST** `/auth/logout`  
**Access**: Admin  
**Headers**: `Authorization: Bearer <token>`

**Response** (`data`): `null`

---

### B. Task Management

#### 1. Enqueue Task
**POST** `/queue`  
**Access**: Shared (Admin + API Key)  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
{
  "root_folder": "C:\\Data\\AccessDB",
  "branch_id": 1,
  "gate_id": 1,
  "station_id": 1,
  "filter": {
    "date_mode": "daily",
    "date": "2025-12-15",
    "range_start": "2025-12-01",
    "range_end": "2025-12-31",
    "transaction_status": "periodic"
  },
  "settings": {
    "branch_name": "BALMERA",
    "management_company": "PT Jasa Marga",
    "page_size": "A4",
    "output_filename_format": "{branch_id}_{date}_{gate_id}"
  }
}
```

**Response** (`data`):
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "queued",
  "queue_position": 3,
  "queue_size": 5,
  "created_at": "2025-12-15T10:00:00Z"
}
```

---

#### 2. List Tasks
**GET** `/tasks`  
**Access**: Shared

**Query Parameters**:
| Param    | Type   | Default | Description              |
| -------- | ------ | ------- | ------------------------ |
| `page`   | int    | 1       | Page number              |
| `limit`  | int    | 10      | Items per page (max 100) |
| `status` | string | -       | Filter by status         |
| `from`   | date   | -       | Filter from date         |
| `to`     | date   | -       | Filter to date           |

**Response** (`data`):
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "status": "completed",
      "output_file_size": 102400,
      "created_at": "2025-12-15T10:00:00Z",
      "updated_at": "2025-12-15T10:05:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "page": 1,
    "limit": 10,
    "total_pages": 10
  }
}
```

---

#### 3. Get Task Detail
**GET** `/tasks/:id`  
**Access**: Shared

**Response** (`data`):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "schedule_id": null,
  "status": "completed",
  "metadata": {
    "branch_id": 1,
    "gate_id": 1
  },
  "output_file_path": "output/001_20251215_A1.pdf",
  "output_file_size": 102400,
  "created_at": "2025-12-15T10:00:00Z",
  "updated_at": "2025-12-15T10:05:00Z"
}
```

---

#### 4. Cancel Task
**DELETE** `/tasks/:id`  
**Access**: Shared

**Response** (`data`):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "cancelled"
}
```

---

#### 5. Download PDF
**GET** `/tasks/:id/download`  
**Access**: Shared

**Response**: Binary PDF file stream  
**Headers**: `Content-Type: application/pdf`

**Error**: Returns JSON error if task is not `completed`.

---

### C. Scheduler

#### 1. Create Schedule
**POST** `/schedules`  
**Access**: Shared  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
{
  "cron": "0 0 * * *",
  "task_payload": {
    "root_folder": "C:\\Data\\AccessDB",
    "branch_id": 1,
    "gate_id": 1,
    "station_id": 1,
    "filter": {
      "date_mode": "yesterday"
    }
  }
}
```

**Response** (`data`):
```json
{
  "schedule_id": "550e8400-e29b-41d4-a716-446655440002",
  "cron": "0 0 * * *",
  "next_run": "2025-12-16T00:00:00Z",
  "active": true
}
```

---

#### 2. List Schedules
**GET** `/schedules`  
**Access**: Shared

**Response** (`data`):
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "cron": "0 0 * * *",
      "active": true,
      "last_run": "2025-12-15T00:00:00Z",
      "next_run": "2025-12-16T00:00:00Z"
    }
  ]
}
```

---

#### 3. Delete Schedule
**DELETE** `/schedules/:id`  
**Access**: Shared

**Response** (`data`): `null`

---

### D. Settings (Admin Only)

#### 1. Get All Settings
**GET** `/settings`  
**Access**: Admin

**Response** (`data`):
```json
{
  "settings": [
    { "key": "branch_id", "value": "001", "description": "Default Branch ID" },
    { "key": "branch_name", "value": "BALMERA", "description": "Branch Name" },
    { "key": "management_company", "value": "PT Jasa Marga", "description": "Company Name" },
    { "key": "page_size", "value": "A4", "description": "PDF Page Size" },
    { "key": "output_filename_format", "value": "{branch_id}_{date}", "description": "Output Filename Template" },
    { "key": "time_overlap", "value": "0", "description": "Time overlap in minutes" },
    { "key": "max_output_age_days", "value": "7", "description": "Auto-delete files older than N days" },
    { "key": "max_concurrent_sessions", "value": "5", "description": "Max admin sessions" },
    { "key": "queue_concurrency", "value": "1", "description": "Parallel queue workers" },
    { "key": "enable_hmac", "value": "true", "description": "Enable/Disable HMAC signature globally" }
  ]
}
```

---

#### 2. Update Setting
**PUT** `/settings`  
**Access**: Admin  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
{
  "key": "max_output_age_days",
  "value": "14"
}
```

**Response** (`data`):
```json
{
  "key": "max_output_age_days",
  "value": "14"
}
```

---

### E. API Keys (Admin Only)

#### 1. List API Keys
**GET** `/api-keys`  
**Access**: Admin

**Response** (`data`):
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "External Client A",
      "active": true,
      "created_at": "2025-12-01T00:00:00Z"
    }
  ]
}
```

---

#### 2. Create API Key
**POST** `/api-keys`  
**Access**: Admin  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
{
  "name": "External Client B"
}
```

**Response** (`data`):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "name": "External Client B",
  "api_key": "pk_live_abc123xyz789",
  "active": true
}
```

> **Note**: `api_key` is shown only once upon creation.

---

#### 3. Show API Key
**GET** `/api-keys/:id/show`  
**Access**: Admin

**Response** (`data`):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "api_key": "pk_live_abc123xyz789"
}
```

---

#### 4. Toggle API Key Status
**PUT** `/api-keys/:id/toggle`  
**Access**: Admin

**Response** (`data`):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "active": false
}
```

---

#### 5. Delete API Key
**DELETE** `/api-keys/:id`  
**Access**: Admin

**Response** (`data`): `null`

---

### F. Sessions (Admin Only)

#### 1. List Active Sessions
**GET** `/sessions`  
**Access**: Admin

**Response** (`data`):
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2025-12-15T08:00:00Z",
      "expires_at": "2025-12-15T20:00:00Z"
    }
  ]
}
```

---

#### 2. Revoke Session
**DELETE** `/sessions/:id`  
**Access**: Admin

**Response** (`data`): `null`

---

### G. Realtime (SSE)

#### 1. Global Event Stream
**GET** `/sse/events`  
**Access**: Admin

**Event Types**:
- `global_stats`: `{ "queue_size": 5, "running_tasks": 2, "active_sessions": 3 }`
- `task_update`: `{ "id": "...", "status": "running" }`
- `log`: `{ "level": "info", "message": "...", "timestamp": "..." }`

---

#### 2. Task Status Stream
**GET** `/sse/tasks/:id`  
**Access**: Shared

**Event Types**:
- `status`: `{ "status": "running", "progress": 50 }`
- `completed`: `{ "status": "completed", "output_size": 102400 }`
- `error`: `{ "status": "failed", "error": "..." }`

---

### H. Stations

#### 1. List Stations
**GET** `/stations`  
**Access**: Protected

**Response** (`data`):
```json
{
  "stations": [
    {
      "id": 1,
      "name": "Station A",
      "created_at": "2025-12-15T00:00:00Z",
      "updated_at": "2025-12-15T00:00:00Z"
    }
  ]
}
```

---

#### 2. Create Station (Single or Batch)
**POST** `/stations`  
**Access**: Shared  
**Headers**: `X-Signature` (Required)

**Request Body (Single)**:
```json
{
  "id": 2,
  "name": "Station B"
}
```

**Request Body (Batch)**:
```json
[
  { "id": 3, "name": "Station C" },
  { "id": 4, "name": "Station D" }
]
```

**Response** (`data`):
- Single: Station object
- Batch: `{ "count": 2 }`

---

#### 3. Update Stations (Batch)
**PUT** `/stations`  
**Access**: Shared  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
[
  { "id": 1, "name": "Updated Station A" }
]
```

**Response** (`data`): `{ "count": 1 }`

---

#### 4. Update Station (Single)
**PUT** `/stations/:id`  
**Access**: Shared  
**Headers**: `X-Signature` (Required)

**Request Body**:
```json
{
  "name": "Updated Station Name"
}
```

**Response** (`data`): `{ "id": 5, "name": "Updated Station Name" }`

---

#### 5. Delete Stations
**DELETE** `/stations` (Batch) or `/stations/:id` (Single)  
**Access**: Shared

**Request Body (Batch)**:
```json
[1, 2, 3]
```

**Response** (`data`):
- Single: `{ "deleted": 1 }`
- Batch: `{ "count": 3 }`

---


## Postman Collection
A Postman collection is available for this API.

**Collection ID**: `bfdcfbe0-0ec2-4661-b332-25795feecb38`  
**Local Backup**: `docs/postman/DataLane_Collection.json`

The collection is kept in sync with this specification via the MCP Postman Server.
The collection configuration includes a **Pre-request Script** that automatically calculates the `X-Signature` header for POST/PUT/PATCH requests using the currently active `api_key` or `token`.
