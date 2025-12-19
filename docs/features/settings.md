# Settings Feature

## Overview
The Settings feature allows dynamic configuration of the PDF Generator application. Settings are stored in the database but cached in memory for high performance.

## Caching Behavior
- **Initialization**: All settings are loaded into a global in-memory cache upon application startup.
- **Read Operations**: All `Get` operations read directly from the cache, eliminating database queries for frequent access (e.g., middleware, config injection).
- **Write Operations**: `Set` or `Update` operations write to the database first. If successful, the cache is updated immediately to ensure consistency.
- **GetAll**: Retrieves full settings objects from the cache.

## Configuration Keys
Key settings include:
- `enable_hmac`: Toggles HMAC signature verification for API requests (Global).
- `branch_id`, `branch_name`: Identifies the station/branch.
- `queue_concurrency`: Controls parallel processing of tasks.

## API
Settings are managed via the `/api/settings` endpoints (Admin only).
See `api_spec.md` for full API details.
