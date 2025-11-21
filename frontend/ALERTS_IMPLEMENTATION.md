# Alerts Management Page Implementation

## Overview
This implementation provides a complete UI for managing alert rules and notification channels in V-Insight.

## Components Created

### 1. AlertRuleModal (`/frontend/src/lib/components/AlertRuleModal.svelte`)
A modal component for creating and editing alert rules with the following features:

**Form Fields:**
- **Name** (required): Name of the alert rule
- **Monitor**: Dropdown to select a specific monitor or "All monitors"
- **Trigger Type** (required): Down, Slow Response, or SSL Expiry
- **Threshold Value** (required): Dynamic based on trigger type
  - Down: Number of consecutive failed checks
  - Slow Response: Response time in milliseconds
  - SSL Expiry: Days before certificate expiration
- **Alert Channels**: Multi-select checkboxes for notification channels
- **Enabled**: Toggle to enable/disable the rule

**Features:**
- Form validation with error messages
- Dynamic threshold labels and help text based on trigger type
- Multi-select for alert channels with visual feedback
- Loads monitors and channels from API on open
- Supports both create and edit modes
- API Integration: POST `/api/v1/alert-rules` (create) and PUT `/api/v1/alert-rules/:id` (edit)

### 2. AlertChannelModal (`/frontend/src/lib/components/AlertChannelModal.svelte`)
A modal component for creating and editing alert channels with dynamic configuration forms.

**Form Fields:**
- **Name** (required): Name of the alert channel
- **Type** (required): Webhook, Discord, or Email (locked after creation)
- **Configuration** (dynamic): Changes based on selected type
  - Webhook: URL input with validation
  - Discord: Discord Webhook URL with domain validation
  - Email: Email address with format validation
- **Enabled**: Toggle to enable/disable the channel

**Features:**
- Dynamic configuration forms based on channel type
- Type-specific validation (URL format, Discord domain, email format)
- Type field is disabled in edit mode (cannot change type after creation)
- API Integration: POST `/api/v1/alert-channels` (create) and PUT `/api/v1/alert-channels/:id` (edit)

### 3. Alerts Page (`/frontend/src/routes/alerts/+page.svelte`)
Main page with tabbed interface for managing alert rules and channels.

**Features:**

#### Alert Rules Tab:
- Displays all alert rules in a table
- Columns: Name, Monitor, Trigger Type, Threshold, Channels, Status
- "Create Rule" button to open creation modal
- Click on status badge to quickly toggle enabled/disabled
- Edit and Delete actions for each rule
- Shows monitor name or "All monitors"
- Displays number of channels attached to each rule

#### Alert Channels Tab:
- Displays all alert channels in a table
- Columns: Name, Type, Configuration, Status
- "Create Channel" button to open creation modal
- Click on status badge to quickly toggle enabled/disabled
- Edit and Delete actions for each channel
- Shows truncated configuration details (URL, email, etc.)

**Common Features:**
- Loading spinner while fetching data
- Error handling with user-friendly messages
- Responsive design with proper spacing and layout
- Consistent styling with the rest of the application
- Tab navigation between Rules and Channels
- Real-time data reload after any CRUD operation
- Confirmation dialog before deletion

## API Endpoints Used

### Alert Rules
- `GET /api/v1/alert-rules` - List all rules
- `POST /api/v1/alert-rules` - Create new rule
- `GET /api/v1/alert-rules/:id` - Get rule details
- `PUT /api/v1/alert-rules/:id` - Update rule
- `DELETE /api/v1/alert-rules/:id` - Delete rule

### Alert Channels
- `GET /api/v1/alert-channels` - List all channels
- `POST /api/v1/alert-channels` - Create new channel
- `GET /api/v1/alert-channels/:id` - Get channel details
- `PUT /api/v1/alert-channels/:id` - Update channel
- `DELETE /api/v1/alert-channels/:id` - Delete channel

### Supporting APIs
- `GET /api/v1/monitors` - Get list of monitors for dropdown

## Validation Rules

### Alert Rules
- Name: Required, non-empty
- Trigger Type: Required, must be 'down', 'slow_response', or 'ssl_expiry'
- Threshold Value:
  - Down: Minimum 1 check
  - Slow Response: Minimum 100ms
  - SSL Expiry: Minimum 1 day

### Alert Channels
- Name: Required, non-empty
- Type: Required, must be 'webhook', 'discord', or 'email'
- Configuration:
  - Webhook: Valid URL format
  - Discord: Valid URL format + must be discord.com domain
  - Email: Valid email format (regex: `/^[^\s@]+@[^\s@]+\.[^\s@]+$/`)

## User Experience

### Creating an Alert Rule
1. Click "Create Rule" button
2. Fill in rule name
3. Optionally select a specific monitor (default is "All monitors")
4. Select trigger type (auto-updates threshold label)
5. Set threshold value based on trigger type
6. Select one or more notification channels
7. Toggle enabled if needed
8. Click "Create Rule"
9. Rule appears in the table immediately

### Creating an Alert Channel
1. Click "Create Channel" button
2. Fill in channel name
3. Select channel type
4. Fill in type-specific configuration (URL, email, etc.)
5. Toggle enabled if needed
6. Click "Create Channel"
7. Channel appears in the table immediately

### Editing
- Click "Edit" on any rule or channel
- Modal opens pre-populated with current values
- Make changes and click "Update"
- Changes reflected immediately in the table

### Deleting
- Click "Delete" on any rule or channel
- Confirmation dialog appears
- Confirm to permanently delete
- Item removed from table immediately

### Toggling Status
- Click the status badge (Enabled/Disabled)
- Status updates immediately without opening modal
- Visual feedback with color change (green for enabled, gray for disabled)

## Styling and Design

- Consistent with existing V-Insight components
- Uses Tailwind CSS utility classes
- Responsive tables that scroll horizontally on small screens
- Modal dialogs with backdrop click-to-close
- Color-coded status badges
- Hover effects on interactive elements
- Loading states with spinner animation
- Error states with red alert boxes

## Integration Notes

- All components use the `fetchAPI` helper from `$lib/api/client`
- Automatic token management and 401 redirect handling
- CORS-free architecture via SvelteKit proxy
- Multi-tenant: All API calls are automatically scoped to the authenticated tenant
- TypeScript interfaces for type safety
- Event dispatchers for parent-child communication

## Testing Recommendations

1. **Create Flow**: Test creating rules and channels with various configurations
2. **Edit Flow**: Verify all fields can be edited and saved correctly
3. **Delete Flow**: Ensure deletion works and shows confirmation
4. **Toggle Flow**: Test quick enable/disable functionality
5. **Validation**: Try submitting invalid data to verify error messages
6. **Multi-select**: Test selecting multiple channels for a rule
7. **Monitor Selection**: Test "All monitors" vs specific monitor selection
8. **Dynamic Forms**: Verify channel config form changes with type selection
9. **API Errors**: Test behavior when API calls fail
10. **Loading States**: Verify spinners appear during data fetching

## Future Enhancements

Potential improvements for future iterations:
- Search/filter functionality for rules and channels
- Bulk actions (enable/disable multiple rules)
- Rule testing/preview functionality
- Channel test notifications
- Advanced filtering by monitor, type, or status
- Sorting capabilities in tables
- Pagination for large datasets
- Import/export rules configuration
- Rule templates for common scenarios
- Analytics on alert frequency and channels
