# Schedules Management UI

The Schedules Page allows managing cron-based task schedules.

## Features
- **Scrollable Create Schedule Modal**: The "Create Schedule" dialog utilizes a `ScrollArea` component.
- **Table Filters**:
    - **Gate ID**: Filter schedules by Target Gate ID.
    - **Station ID**: Filter schedules by Target Station ID.
    - **Date Filter**: Text search within the "Date Filter" column (filters by the description of the task's date logic).
    - **Status**: Filter by Active/Paused.
- **Date Filter Column**: Displays the task's date filter logic (e.g. "Yesterday").
