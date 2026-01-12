# Queue Management UI

The Queue Page displays the list of tasks (queued, running, completed, failed).

## Features
- **Scrollable Add Task Modal**: The "Enqueue New Task" dialog utilizes a `ScrollArea` component to handle long forms gracefully without manual scrollbars.
- **Date Filter Column**: Displays the formatted date filter criteria used for the task (e.g., "Yesterday", "Range: YYYY-MM-DD - YYYY-MM-DD").
- **Table Filters**: Filter tasks by:
    - Status
    - Gate ID
    - Station ID
    - Date Range (Created At)
