# Dashboard Overview Feature

## Overview
The dashboard provides a real-time overview of the monitoring system status, including statistics, recent checks, and open incidents.

## Backend Implementation

### Endpoints

#### GET /api/v1/dashboard
Returns complete dashboard data including:
- Statistics (total monitors, up count, down count, open incidents)
- Recent checks (last 10 monitor checks with details)
- Open incidents (all currently open incidents with monitor details)

**Authentication:** Required (Bearer token)

**Response:**
```json
{
  "stats": {
    "total_monitors": 5,
    "up_count": 4,
    "down_count": 1,
    "open_incidents": 2
  },
  "recent_checks": [
    {
      "check": {
        "id": "uuid",
        "monitor_id": "uuid",
        "checked_at": "2024-01-01T12:00:00Z",
        "status_code": 200,
        "response_time_ms": 150,
        "success": true
      },
      "monitor": {
        "id": "uuid",
        "name": "My Website",
        "url": "https://example.com",
        ...
      }
    }
  ],
  "open_incidents": [
    {
      "incident": {
        "id": "uuid",
        "monitor_id": "uuid",
        "alert_rule_id": "uuid",
        "started_at": "2024-01-01T10:00:00Z",
        "status": "open",
        ...
      },
      "monitor": {
        "id": "uuid",
        "name": "My Website",
        "url": "https://example.com",
        ...
      }
    }
  ]
}
```

#### GET /api/v1/dashboard/stats
Returns just the statistics (lighter endpoint).

**Authentication:** Required (Bearer token)

**Response:**
```json
{
  "total_monitors": 5,
  "up_count": 4,
  "down_count": 1,
  "open_incidents": 2
}
```

### Implementation Details

**Location:** `backend/internal/api/handlers/dashboard_handler.go`

**Key Features:**
- Multi-tenant support (filters by tenant_id from JWT)
- Counts only enabled monitors for up/down statistics
- Handles monitors without checks gracefully
- Efficient database queries (one per monitor for checks)
- Error handling for database failures

**Logic:**
- `total_monitors`: Total count of all monitors for the tenant
- `up_count`: Enabled monitors with latest check success = true
- `down_count`: Enabled monitors with latest check success = false
- `open_incidents`: Count of incidents with status = "open"
- `recent_checks`: Last check from each monitor (up to 10)

## Frontend Implementation

### Components

#### StatCard.svelte
Reusable component for displaying statistics.

**Props:**
- `title: string` - Card title
- `value: string | number` - Value to display
- `valueColor: string` - Tailwind color class for value (default: 'text-gray-900')
- `bgColor: string` - Tailwind background class (default: 'bg-white')

**Usage:**
```svelte
<StatCard title="Total Monitors" value={stats.total_monitors} valueColor="text-gray-900" />
```

#### MonitorStatus.svelte
Displays monitor status with animated indicator.

**Props:**
- `status: 'up' | 'down' | 'unknown'` - Monitor status
- `showText: boolean` - Show text label (default: true)

**Features:**
- Animated pulsing dot
- Color-coded: green (up), red (down), gray (unknown)
- Optional text label

**Usage:**
```svelte
<MonitorStatus status={check.success ? 'up' : 'down'} />
```

#### IncidentBadge.svelte
Displays incident status badge.

**Props:**
- `status: 'open' | 'resolved'` - Incident status
- `severity: 'critical' | 'warning' | 'info'` - Severity level (default: 'warning')

**Features:**
- Color-coded badges
- Open incidents: red (critical), yellow (warning), blue (info)
- Resolved incidents: green

**Usage:**
```svelte
<IncidentBadge status={incident.status} severity="warning" />
```

### Dashboard Page

**Location:** `frontend/src/routes/dashboard/+page.svelte`

**Features:**
- Client-side data loading using fetchAPI
- Loading state with spinner
- Error handling with user-friendly messages
- Responsive grid layout (1-4 columns based on screen size)
- Relative time formatting ("2 minutes ago")
- Real-time status indicators

**Layout:**
1. **Stats Cards (4 columns on desktop, 2 on tablet, 1 on mobile):**
   - Total Monitors (gray)
   - Monitors Up (green)
   - Monitors Down (red)
   - Open Incidents (yellow)

2. **Two-column layout (1 column on mobile):**
   - **Recent Checks:** Last 10 monitor checks with:
     - Status indicator (up/down)
     - Monitor name and URL
     - Response time
     - Error message (if any)
     - Relative timestamp
     - HTTP status code

   - **Open Incidents:** All open incidents with:
     - Incident badge (status/severity)
     - Monitor name and URL
     - Trigger value
     - Relative timestamp (when started)

## Usage

### Backend
The dashboard endpoints are automatically registered in `backend/cmd/api/main.go` and protected by authentication and tenant middleware.

### Frontend
The dashboard page is accessible at `/dashboard` and automatically loads data on mount.

## Styling

All components use Tailwind CSS for styling with consistent design patterns:
- Cards: white background, rounded corners, shadow
- Grid layouts: responsive with gap-6
- Colors: green (success), red (error), yellow (warning), gray (neutral)
- Typography: consistent font sizes and weights
- Spacing: consistent padding and margins

## Performance Considerations

- Backend queries one check per monitor (limit 1)
- Recent checks limited to 10 items
- No pagination (suitable for small to medium monitor counts)
- For large deployments (100+ monitors), consider adding pagination

## Future Enhancements

- Real-time updates using WebSocket
- Pagination for recent checks and incidents
- Filtering and sorting options
- Customizable dashboard widgets
- Export functionality
- Historical trend charts
