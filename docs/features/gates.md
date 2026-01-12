# Gate Management

## Overview
Allows management of gates (toll gates) in the system.

## Data Model
- **ID**: integer (unique)
- **Name**: string
- **Created/Updated**: timestamp

## API Endpoints

### List Gates
`GET /api/gates`

### Create Gate
`POST /api/gates`
Body:
```json
{ "id": 1, "name": "Gate A" }
```

### Batch Create
`POST /api/gates`
Body: `[{ "id": 1, "name": "Gate A" }, ...]`

### Update Gate
`PUT /api/gates/:id`
Body: `{ "name": "New Name" }`

### Delete Gate
`DELETE /api/gates/:id` by ID or Name.

### Creation & Editing
- **Single Create**: Add a new gate manually by specifying ID and Name.
- **Batch Insert**: Add multiple gates simultaneously using JSON format.
  - Supports large JSON payloads with a scrollable input view.
  - JSON Format: `[{"id": "001", "name": "Gate A"}, ...]`
- **Edit**: Modify gate details (Name).
- **Delete**: Remove a gate.

### Batch Actions
- **Batch Edit**: Update multiple gates via JSON.
- **Batch Delete**: Delete multiple selected gates.

## UI Components
- **Modal**: Creation and Editing occur in a modal dialog.
- **Tabs**: Switch between Single and Batch modes.
- **Scrollable Input**: The batch JSON input text area has a maximum height of 400px and supports scrolling to handle large datasets effectively.
