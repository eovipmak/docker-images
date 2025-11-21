# Monitors Management - Visual Component Guide

## Page Flow

```
/domains (Monitors List)
    |
    |-- Click "Add Monitor" --> Modal Opens --> Create New Monitor
    |
    |-- Click "Edit" Button --> Modal Opens --> Update Monitor
    |
    |-- Click "Delete" Button --> Confirmation --> Delete Monitor
    |
    |-- Click Monitor Row --> /domains/{id} (Monitor Details)
                                  |
                                  |-- View Statistics
                                  |-- View Charts
                                  |-- View SSL Info
                                  |-- View Recent Checks
                                  |
                                  |-- Click "Back" --> /domains
```

## Component Hierarchy

```
/domains (+page.svelte)
├── MonitorTable
│   ├── Search Input
│   ├── Table Header (sortable)
│   ├── Table Rows
│   │   ├── MonitorStatus Component
│   │   ├── Edit Button (triggers edit event)
│   │   └── Delete Button (triggers delete event)
│   └── Empty State
└── MonitorModal
    ├── Form Fields
    │   ├── Name Input
    │   ├── URL Input
    │   ├── Check Interval Input
    │   ├── Timeout Input
    │   ├── SSL Check Toggle
    │   ├── SSL Alert Days Input
    │   └── Enabled Toggle
    ├── Validation Logic
    └── Submit Button

/domains/[id] (+page.svelte)
├── Back Button
├── Header
│   ├── Monitor Name
│   ├── Monitor URL
│   └── MonitorStatus Component
├── Stats Cards Grid
│   ├── Status Card
│   ├── Uptime Card
│   ├── Avg Response Time Card
│   └── Check Interval Card
├── Charts Section
│   ├── Uptime History Chart (CSS bars)
│   └── Response Time Chart (CSS bars)
├── SSL Certificate Section (conditional)
│   └── SSL Information Grid
├── Monitor Settings Section
│   └── Settings Grid
└── Recent Checks Section
    └── Checks Table
```

## Data Flow

### Create Monitor Flow
```
User clicks "Add Monitor"
    ↓
MonitorModal opens (monitor = null)
    ↓
User fills form
    ↓
User clicks "Create Monitor"
    ↓
Validation runs
    ↓
POST /api/v1/monitors
    ↓
Response received
    ↓
Modal emits 'save' event
    ↓
Modal closes
    ↓
Monitors list refreshes
    ↓
New monitor appears in table
```

### Edit Monitor Flow
```
User clicks "Edit" on table row
    ↓
MonitorModal opens (monitor = selected monitor data)
    ↓
Form pre-populated with current values
    ↓
User modifies fields
    ↓
User clicks "Update Monitor"
    ↓
Validation runs
    ↓
PUT /api/v1/monitors/{id}
    ↓
Response received
    ↓
Modal emits 'save' event
    ↓
Modal closes
    ↓
Monitors list refreshes
    ↓
Updated values appear in table
```

### Delete Monitor Flow
```
User clicks "Delete" on table row
    ↓
Confirmation dialog appears
    ↓
User confirms
    ↓
DELETE /api/v1/monitors/{id}
    ↓
Response received
    ↓
Monitors list refreshes
    ↓
Monitor removed from table
```

### View Monitor Details Flow
```
User clicks on monitor row
    ↓
Navigate to /domains/{id}
    ↓
Load monitor data:
  - GET /api/v1/monitors/{id}
  - GET /api/v1/monitors/{id}/checks
  - GET /api/v1/monitors/{id}/ssl-status (if HTTPS)
    ↓
Render page with:
  - Statistics calculated from data
  - Charts rendered with check history
  - SSL info displayed (if available)
  - Recent checks in table
```

## Component Props & Events

### MonitorTable Component

**Props:**
- `monitors: any[]` - Array of monitor objects

**Events:**
- `view` - Emitted when row is clicked, detail = monitor object
- `edit` - Emitted when edit button clicked, detail = monitor object
- `delete` - Emitted when delete button clicked, detail = monitor object

**Usage:**
```svelte
<MonitorTable
  {monitors}
  on:view={handleViewMonitor}
  on:edit={handleEditMonitor}
  on:delete={handleDeleteMonitor}
/>
```

### MonitorModal Component

**Props:**
- `isOpen: boolean` - Controls modal visibility
- `monitor: any | null` - Monitor to edit, or null for create

**Events:**
- `save` - Emitted when monitor is saved, detail = saved monitor
- `close` - Emitted when modal is closed

**Usage:**
```svelte
<MonitorModal
  bind:isOpen={isModalOpen}
  monitor={selectedMonitor}
  on:save={handleModalSave}
  on:close={handleModalClose}
/>
```

## State Management

### Monitors List Page State
```typescript
let monitors: any[] = [];           // All monitors from API
let isLoading = true;                // Loading state
let error = '';                      // Error message
let isModalOpen = false;             // Modal visibility
let selectedMonitor: any = null;     // Monitor being edited
```

### Monitor Detail Page State
```typescript
let monitorId: string;               // From route params
let monitor: any = null;             // Monitor details
let checks: any[] = [];              // Check history
let sslStatus: any = null;           // SSL certificate info
let isLoading = true;                // Loading state
let error = '';                      // Error message
```

### MonitorModal State
```typescript
let formData: FormData = {           // Form fields
  name: '',
  url: '',
  check_interval: 300,
  timeout: 30,
  enabled: true,
  check_ssl: true,
  ssl_alert_days: 30
};
let errors: Record<string, string> = {};  // Validation errors
let isSubmitting = false;            // Submit state
```

## Styling Patterns

### Colors Used
- **Blue** (`bg-blue-600`) - Primary actions (Add, Save buttons)
- **Green** (`bg-green-500`) - Success, Up status
- **Red** (`bg-red-500`) - Errors, Down status
- **Gray** (`bg-gray-50`) - Backgrounds, disabled state
- **Yellow** - Warning states

### Responsive Breakpoints
- **Mobile**: Default (< 768px)
- **Tablet**: `md:` (≥ 768px)
- **Desktop**: `lg:` (≥ 1024px)

### Common Classes
- **Cards**: `bg-white rounded-lg shadow-md p-6`
- **Buttons**: `px-4 py-2 rounded-md transition-colors`
- **Tables**: `min-w-full divide-y divide-gray-200`
- **Inputs**: `w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500`

## Chart Implementation

### Uptime Chart (CSS Bars)
```svelte
<div class="flex items-end gap-1 h-48">
  {#each checks as check}
    <div
      class="flex-1 rounded-t"
      class:bg-green-500={check.success}
      class:bg-red-500={!check.success}
      style="height: {check.success ? '100%' : '20%'}"
    ></div>
  {/each}
</div>
```

### Response Time Chart (CSS Bars)
```svelte
{@const maxTime = Math.max(...responseTimes)}
<div class="flex items-end gap-1 h-48">
  {#each responseTimes as time}
    <div
      class="flex-1 bg-blue-500 rounded-t"
      style="height: {(time / maxTime) * 100}%"
    ></div>
  {/each}
</div>
```

## Validation Rules

### Monitor Form Validation

1. **Name**
   - Required
   - Must not be empty after trim

2. **URL**
   - Required
   - Must be valid URL format
   - Validated using `new URL()`

3. **Check Interval**
   - Must be ≥ 60 seconds
   - Integer value

4. **Timeout**
   - Must be between 5-120 seconds
   - Integer value

5. **SSL Alert Days**
   - Must be ≥ 1 day
   - Integer value

## API Response Types

### Monitor Object
```typescript
{
  id: string;
  tenant_id: number;
  name: string;
  url: string;
  check_interval: number;
  timeout: number;
  enabled: boolean;
  check_ssl: boolean;
  ssl_alert_days: number;
  last_checked_at?: string;
  created_at: string;
  updated_at: string;
}
```

### Monitor Check Object
```typescript
{
  id: string;
  monitor_id: string;
  checked_at: string;
  status_code?: number;
  response_time_ms?: number;
  ssl_valid?: boolean;
  ssl_expires_at?: string;
  error_message?: string;
  success: boolean;
}
```

### SSL Status Object
```typescript
{
  ssl_valid: boolean;
  ssl_expires_at?: string;
  ssl_issuer?: string;
  error_message?: string;
}
```

This guide provides a complete visual overview of how all components work together in the Monitors Management feature.
