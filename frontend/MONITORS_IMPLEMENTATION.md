# Monitors Management Implementation

This document describes the implementation of the Monitors Management feature for V-Insight.

## Overview

The Monitors Management feature provides a comprehensive UI for managing website/service monitors with full CRUD operations, sortable tables, and detailed monitor views with visual charts.

## Components Created

### 1. MonitorTable Component (`src/lib/components/MonitorTable.svelte`)

A reusable table component for displaying monitors with the following features:

- **Search/Filter**: Real-time search across monitor names and URLs
- **Sortable Columns**: Click column headers to sort by:
  - Name
  - URL
  - Last Check time
- **Status Indicators**: Color-coded status dots showing monitor health
- **Actions**: Edit and Delete buttons for each monitor
- **Row Click**: Click any row to navigate to monitor details
- **Responsive Design**: Mobile-friendly layout with horizontal scrolling

**Events:**
- `view` - Emitted when a monitor row is clicked
- `edit` - Emitted when the edit button is clicked
- `delete` - Emitted when the delete button is clicked

### 2. MonitorModal Component (`src/lib/components/MonitorModal.svelte`)

A modal dialog for creating and editing monitors with comprehensive validation:

**Form Fields:**
- `name` (required) - Monitor display name
- `url` (required) - Website/service URL to monitor
- `check_interval` - How often to check (seconds, min: 60)
- `timeout` - Request timeout (seconds, 5-120)
- `enabled` - Whether monitoring is active
- `check_ssl` - Enable SSL certificate monitoring
- `ssl_alert_days` - Days before expiry to alert (min: 1)

**Features:**
- Client-side validation with error messages
- URL format validation
- Positive integer validation for intervals/timeouts
- Conditional SSL settings display
- Loading states during submission
- Error handling with user feedback

**Props:**
- `isOpen` - Boolean to control modal visibility
- `monitor` - Monitor object for editing (null for create)

**Events:**
- `save` - Emitted when monitor is successfully saved
- `close` - Emitted when modal is closed

### 3. Monitors List Page (`src/routes/domains/+page.svelte`)

The main page for managing monitors:

**Features:**
- Displays all monitors in a sortable, filterable table
- "Add Monitor" button to create new monitors
- Integration with MonitorTable and MonitorModal components
- Loading and error states
- Automatic refresh after create/update/delete operations
- Confirmation dialog before deletion

**API Integration:**
- `GET /api/v1/monitors` - Load all monitors
- `POST /api/v1/monitors` - Create new monitor
- `PUT /api/v1/monitors/:id` - Update existing monitor
- `DELETE /api/v1/monitors/:id` - Delete monitor

### 4. Monitor Detail Page (`src/routes/domains/[id]/+page.svelte`)

Comprehensive view of a single monitor with statistics and visualizations:

**Features:**
- **Statistics Cards:**
  - Current status (Enabled/Disabled)
  - Uptime percentage (last 24 hours)
  - Average response time
  - Check interval

- **Uptime History Chart:**
  - Visual bar chart showing success/failure over last 24 hours
  - Green bars for successful checks
  - Red bars for failed checks
  - Hover tooltips with timestamp and status

- **Response Time Chart:**
  - Bar chart showing response times
  - Hover tooltips with exact millisecond values
  - Dynamic scaling based on max response time

- **SSL Certificate Info** (for HTTPS monitors):
  - Validity status
  - Expiration date
  - Issuer information
  - Error messages if applicable

- **Monitor Settings:**
  - Timeout configuration
  - SSL check status
  - SSL alert threshold
  - Last check timestamp

- **Recent Checks Table:**
  - Last 10 checks with details
  - Timestamp, status, HTTP code, response time
  - Error messages if failures occurred

**API Integration:**
- `GET /api/v1/monitors/:id` - Load monitor details
- `GET /api/v1/monitors/:id/checks` - Load check history
- `GET /api/v1/monitors/:id/ssl-status` - Load SSL information

## Installation

### Quick Setup

Create the detail page directory (one-time only):

```bash
cd frontend/src/routes/domains
mkdir '[id]'
```

Then copy the monitor detail page content from `MONITOR_DETAIL_PAGE.md` into `src/routes/domains/[id]/+page.svelte`.

**Important:** This is a ONE-TIME setup for the SvelteKit dynamic route. Once created, it handles ALL monitors automatically.

### Dependencies

No additional npm dependencies are required. The implementation uses:
- Native SvelteKit features
- Existing components (MonitorStatus, existing from dashboard)
- Tailwind CSS for styling
- Pure CSS for chart visualizations (no Chart.js needed)

## Usage

### Creating a Monitor

1. Navigate to `/domains`
2. Click "Add Monitor" button
3. Fill in the form:
   - Enter a descriptive name
   - Enter the full URL (including https://)
   - Adjust check interval if needed (default: 300s)
   - Set timeout (default: 30s)
   - Configure SSL settings for HTTPS URLs
4. Click "Create Monitor"

### Editing a Monitor

1. Navigate to `/domains`
2. Find the monitor in the table
3. Click "Edit" button on the right
4. Update fields as needed
5. Click "Update Monitor"

### Deleting a Monitor

1. Navigate to `/domains`
2. Find the monitor in the table
3. Click "Delete" button on the right
4. Confirm deletion in the dialog

### Viewing Monitor Details

1. Navigate to `/domains`
2. Click on any monitor row
3. View detailed statistics, charts, and history

## Technical Details

### State Management

- Uses Svelte's reactive declarations (`$:`) for computed values
- Local component state with `let` declarations
- Event-driven communication between components

### API Communication

- Uses `fetchAPI` helper from `$lib/api/client`
- Automatic authentication token handling
- Proxy architecture (no CORS issues)
- Proper error handling with user feedback

### Responsive Design

- Mobile-first approach with Tailwind CSS
- Grid layouts that adapt to screen size
- Horizontal scrolling for tables on mobile
- Touch-friendly button sizes

### Charts and Visualizations

- Pure CSS bar charts (no external libraries)
- Responsive sizing
- Interactive hover effects
- Tooltips with detailed information
- Simple, performant implementation

## API Endpoints Used

All endpoints are prefixed with `/api/v1/` and require authentication:

- `GET /monitors` - List all monitors for tenant
- `POST /monitors` - Create new monitor
- `GET /monitors/:id` - Get monitor by ID
- `PUT /monitors/:id` - Update monitor
- `DELETE /monitors/:id` - Delete monitor
- `GET /monitors/:id/checks` - Get check history
- `GET /monitors/:id/ssl-status` - Get SSL certificate status

## Future Enhancements

Potential improvements for future iterations:

1. **Advanced Charts:**
   - Use Chart.js for more sophisticated visualizations
   - Interactive time range selection
   - Zooming and panning capabilities

2. **Real-time Updates:**
   - WebSocket integration for live status updates
   - Auto-refresh of monitor details page

3. **Batch Operations:**
   - Select multiple monitors
   - Bulk enable/disable
   - Bulk delete

4. **Filtering:**
   - Filter by status (up/down/disabled)
   - Filter by URL pattern
   - Filter by last check time

5. **Export:**
   - Export monitor list to CSV
   - Export check history
   - Generate PDF reports

6. **Notifications:**
   - Toast notifications for successful operations
   - Alert banners for failed operations

## Testing

### Manual Testing Checklist

- [ ] Create a new monitor with valid data
- [ ] Create a monitor with invalid data (test validation)
- [ ] Edit an existing monitor
- [ ] Delete a monitor
- [ ] Search/filter monitors
- [ ] Sort monitors by different columns
- [ ] Click monitor row to view details
- [ ] View monitor with check history
- [ ] View HTTPS monitor with SSL info
- [ ] Test on mobile devices
- [ ] Test with many monitors (pagination)

### TypeScript Validation

Run TypeScript check:
```bash
cd frontend
npm run check
```

## Troubleshooting

### Monitor detail page not found

If you get a 404 when navigating to `/domains/{id}`, ensure:
1. The `[id]` directory was created correctly
2. The `+page.svelte` file is in the right location
3. Run `npm run dev` to restart the dev server

### Charts not displaying

If charts don't appear:
1. Ensure check history data exists
2. Check browser console for JavaScript errors
3. Verify API responses in Network tab

### Modal not opening

If the modal doesn't open:
1. Check browser console for errors
2. Verify component import in +page.svelte
3. Ensure `isModalOpen` state is properly bound

## Support

For issues or questions, refer to:
- V-Insight Documentation
- SvelteKit Documentation
- Tailwind CSS Documentation
