# Data Source Resolution

## Overview
The system dynamically resolves the path to the Microsoft Access (.mdb) data source based on the station and transaction date provided in the task metadata. This ensures that the correct monthly folder and daily database file are accessed for PDF generation.

## Logic

The data source path is constructed using the following components:
- **Root Folder**: The base directory configured for the task.
- **Date**: Extracted from the task filter.
- **Station ID**: The station identifier.

### Path Format
The file path follows this structure:
```
{root}/{MMYYYY}/{StationID}/{DDMMYYYY}.mdb
```

### Components
- `MMYYYY`: Month (2 digits) and Year (4 digits) of the transaction date.
- `StationID`: 2-digit station ID (padded with '0' if single digit).
- `DDMMYYYY`: Day (2 digits), Month (2 digits), and Year (4 digits) of the transaction date.

### Example
Given:
- **Root Folder**: `D:\TollData`
- **Date**: `2024-12-20`
- **Station ID**: `1`

The resolved path will be:
```
D:\TollData\122024\01\20122024.mdb
```

## Fallback Behavior
- If the date is not specified in the filter, the system attempts to use the `RangeStart` date.
- If no date parsing is successful, the current system date is used as a fallback.
