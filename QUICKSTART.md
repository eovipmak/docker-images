# Quick Start - Monitors Management

## One-Time Setup

Due to tooling limitations with creating dynamic routes (`[id]`), you need to manually create the monitor detail page directory:

```bash
cd frontend/src/routes/domains
mkdir '[id]'
```

Then copy the monitor detail page content from `frontend/MONITOR_DETAIL_PAGE.md` into `frontend/src/routes/domains/[id]/+page.svelte`.

**Note:** This is a ONE-TIME setup for the route structure, NOT per-monitor. Once created, all monitors automatically use this detail page.

## Start the Application

```bash
# From repository root
make up
```

Wait ~30 seconds for all services to start.

## Access the Monitors Page

Open your browser to: **http://localhost:3000/domains**

(You'll need to login first if not already authenticated)

## Features You Can Test

1. **Create Monitor**: Click "Add Monitor" button
2. **View List**: See all monitors in sortable table
3. **Search**: Type in search box to filter
4. **Sort**: Click column headers
5. **Edit**: Click "Edit" button on any row
6. **Delete**: Click "Delete" button (with confirmation)
7. **View Details**: Click on any monitor row to see:
   - Statistics (uptime, response time, status)
   - Uptime chart (last 24 hours)
   - Response time chart
   - SSL certificate info (HTTPS only)
   - Recent checks table

## Troubleshooting

### If TypeScript errors:

Review the output from `npm run check` and ensure all components are properly typed.

### If 404 on detail page:

1. Verify the file exists: `ls src/routes/domains/[id]/+page.svelte`
2. Restart the dev server: `make restart`

### If modal doesn't open:

1. Check browser console for errors
2. Verify you're logged in
3. Check network tab for API errors

## What's Already Working

✅ Backend API endpoints are all implemented
✅ MonitorTable component is ready
✅ MonitorModal component is ready
✅ Monitors list page is updated
✅ All CRUD operations are integrated
✅ Search, sort, filter functionality
✅ Responsive design for mobile

## What Needs Testing

- [ ] Create a monitor with valid data
- [ ] Create a monitor with invalid data (test validation)
- [ ] Edit an existing monitor
- [ ] Delete a monitor (test confirmation)
- [ ] Search/filter monitors
- [ ] Sort by each column
- [ ] View monitor details
- [ ] Check charts display correctly
- [ ] View SSL info for HTTPS monitor
- [ ] Test on mobile device/small screen
- [ ] Test with no monitors (empty state)
- [ ] Test with many monitors (scrolling)

## Next Steps After Testing

1. Review the implementation
2. Provide feedback or request changes
3. Test in production environment
4. Deploy when ready

## Documentation

- `IMPLEMENTATION_COMPLETE.md` - Complete guide
- `frontend/MONITORS_IMPLEMENTATION.md` - Technical details
- `frontend/MONITOR_DETAIL_PAGE.md` - Detail page info

## Need Help?

Check the troubleshooting sections in the documentation files above.
