# Stations Management

The Stations Management page allows users to manage toll stations.

## Features

- **List Stations**: View all stations with ID and Name.
- **Create Station**:
    - Single: Add a station manually (ID and Name).
    - Batch: Add multiple stations via JSON.
    - **Confirmation**: A dialog prompts for confirmation before creation.
    - **Success/Failure**: A bottom-left alert indicates the result.
- **Update Station**: Edit name of existing stations.
- **Delete Station**: Remove stations.

## JSON Format for Batch

```json
[
  { "id": 1, "name": "Station A" },
  { "id": 2, "name": "Station B" }
]
```

## UI Components
- Uses Shadcn `AlertDialog` for confirmation.
- Uses Shadcn `Alert` for notifications.
