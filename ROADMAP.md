# V-Insight Development Roadmap

This document outlines the development plan for the next features of V-Insight. It is divided into stages to be implemented incrementally.

## Status
- [x] Stage 1: Implement Email Notifications (Implemented)
- [x] Stage 2: Add Keyword Monitoring
- [x] Stage 3: Add Ping (ICMP) Monitoring
- [x] Stage 2: Add Keyword Monitoring
- [ ] Stage 3: Add Ping (ICMP) Monitoring
- [ ] Stage 4: Add DNS Monitoring
- [ ] Stage 5: Add Telegram Notifications
- [ ] Stage 6: User Profile Management
- [ ] Stage 7: Status Page Maintenance Windows
- [ ] Stage 8: Reporting & Analytics

---

## Stage 1: Implement Email Notifications
**Goal:** Enable email notifications for incidents.
**Implementation Details:**
- Backend: Added SMTP config and `net/smtp` logic in `worker`.
- Frontend: `AlertChannelModal` already supports email.
- **Status:** Completed.

## Stage 2: Add Keyword Monitoring
**Goal:** Check if a specific keyword exists in the HTTP response body.
**Instructions:**
1.  **Database:** Create a migration to add `keyword` (TEXT, nullable) to `monitors` table.
2.  **Backend:**
    -   Update `Monitor` entity in `backend/internal/domain/entities/monitor.go`.
    -   Update `HealthCheckJob` in `worker/internal/jobs/health_check_job.go` to read response body and check for keyword presence if set.
3.  **Frontend:**
    -   Update `frontend/src/lib/components/MonitorModal.svelte` to add "Keyword" input when type is HTTP.

## Stage 3: Add Ping (ICMP) Monitoring
**Goal:** Add ICMP Ping monitor type.
**Instructions:**
1.  **Database:** Ensure `type` column supports 'ping' (it's VARCHAR, so it should be fine).
2.  **Backend:**
    -   Update `HealthCheckJob` to handle `type="ping"`.
    -   Use a library like `github.com/prometheus-community/pro-bing` or execute `ping` command (ensure Docker has permissions).
3.  **Frontend:**
    -   Update `frontend/src/lib/components/MonitorModal.svelte` to allow selecting "Ping" type. Hide URL field (use Hostname/IP) or reuse URL field.

## Stage 4: Add DNS Monitoring
**Goal:** Check if a domain resolves.
**Instructions:**
1.  **Backend:**
    -   Update `HealthCheckJob` to handle `type="dns"`.
    -   Use `net.LookupHost`.
    -   Optionally support specifying expected IP or DNS server.
2.  **Frontend:**
    -   Update `frontend/src/lib/components/MonitorModal.svelte` to allow selecting "DNS" type.

## Stage 5: Add Telegram Notifications
**Goal:** Send alerts via Telegram.
**Instructions:**
1.  **Frontend:**
    -   Update `frontend/src/lib/components/AlertChannelModal.svelte` to add `telegram` type.
    -   Fields needed: `bot_token`, `chat_id`.
2.  **Backend:**
    -   Update `AlertChannel` entity validation if needed.
    -   Update `worker/internal/jobs/notification_job.go` to implement `sendTelegramNotification` using Telegram Bot API.

## Stage 6: User Profile Management
**Goal:** Allow users to change name and password.
**Instructions:**
1.  **Backend:**
    -   Add `PUT /api/v1/user/profile` (name updates).
    -   Add `PUT /api/v1/user/password` (password change).
2.  **Frontend:**
    -   Create `frontend/src/routes/settings/profile/+page.svelte`.
    -   Add form for profile and password change.

## Stage 7: Status Page Maintenance Windows
**Goal:** Schedule and display maintenance.
**Instructions:**
1.  **Database:** Create `maintenance_windows` table (id, tenant_id, title, description, start_time, end_time, status).
2.  **Backend:**
    -   Add CRUD API: `/api/v1/maintenance`.
3.  **Frontend:**
    -   Add Management UI in Dashboard.
    -   Update Public Status Page (`frontend/src/routes/status/[slug]`) to show active/upcoming maintenance.

## Stage 8: Reporting & Analytics
**Goal:** View historical uptime reports.
**Instructions:**
1.  **Backend:**
    -   Add `GET /api/v1/reports/uptime` (params: from, to).
    -   Calculate uptime percentage per monitor for the period.
2.  **Frontend:**
    -   Create `frontend/src/routes/reports/+page.svelte`.
    -   Display table/charts.
