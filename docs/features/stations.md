# Stations Management

## Overview
The Stations Management module allows users to configure and manage station entities within the DataLane system. Stations are fundamental units for data collection and processing.

## Features

### Station Management
- **List View**: View all registered stations with ID and Name.
- **Selection**: Select multiple stations for batch actions.
- **Search/Filter**: (Planned) Filter stations by ID or Name.

### Creation & Editing
- **Single Create**: Add a new station manually by specifying ID and Name.
- **Batch Insert**: Add multiple stations simultaneously using JSON format.
  - Supports large JSON payloads with a scrollable input view.
  - JSON Format: `[{"id": "001", "name": "Station A"}, ...]`
- **Edit**: Modify station details (Name).
- **Delete**: Remove a station.

### Batch Actions
- **Batch Edit**: Update multiple stations via JSON.
- **Batch Delete**: Delete multiple selected stations.

## UI Components
- **Modal**: Creation and Editing occur in a modal dialog.
- **Tabs**: Switch between Single and Batch modes.
- **Scrollable Input**: The batch JSON input text area has a maximum height of 400px and supports scrolling to handle large datasets effectively.
