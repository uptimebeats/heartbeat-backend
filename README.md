# Heartbeat Monitoring API

## Implementation Details

### Endpoint
- **URL**: `GET /b/:uuid` (also supports POST/HEAD)
- **Rate Limit**: 3 requests per minute per unique ID (In-memory).

### Response Codes
| Status Code | Condition |
| :--- | :--- |
| `200 OK` | Heartbeat received successfully. Site marked UP. |
| `404 Not Found` | Invalid Unique ID (UUID). |
| `429 Too Many Requests` | Limit exceeded (>3 req/min). |

### Background Logic (Cron)
- **Frequency**: Runs every 1 minute.
- **Downtime Detection**:
    - Checks sites where `Last Heartbeat + CheckFrequency + Tolerance < Now`.
    - **Ignored**: Sites that have *never* received a heartbeat (new records).
    - **Action**: Marks site `DOWN`, creates an incident.
- **Recovery**:
    - When a valid heartbeat is received for a `DOWN` site, it is immediately marked `UP` and a "Resolved" incident is created.
