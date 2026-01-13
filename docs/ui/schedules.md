# Schedules Management UI

The Schedules Page allows managing cron-based task schedules.

## Features
- **Scrollable Create Schedule Modal**: The "Create Schedule" dialog utilizes a `ScrollArea` component.
- **Advanced Filters**: "Create Schedule" supports advanced filtering by `Origin Gate ID`, `Transaction Status`, `Limit`, and `Day Start Time` via an accordion.
- **Table Filters**:
    - **Gate ID**: Filter schedules by Target Gate ID.
    - **Station ID**: Filter schedules by Target Station ID.
    - **Transaction Date**: Text search within the "Transaction Date" column (filters by date, range, or "yesterday").
    - **Status**: Filter by Active/Paused.
- **Transaction Date Column**: Displays the task's date filter logic (e.g. "Yesterday", "{date}", or "{start} - {end}").
