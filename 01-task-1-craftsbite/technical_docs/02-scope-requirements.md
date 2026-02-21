## 5. Scope of Changes

### In Scope (Iteration 1)

- ✅ User authentication with JWT and role-based access control
- ✅ User default meal preference (opt-in or opt-out)
- ✅ Daily meal participation management with flexible defaults
- ✅ On-demand headcount calculation based on schedules and cutoff rules
- ✅ Admin/Team Lead override capabilities
- ✅ Day schedule management (holidays, celebrations, office closures)
- ✅ Comprehensive participation history tracking
- ✅ Change tracking for accountability

### In Scope (Iteration 2)

- ✅ Team-based visibility (employees see their team, Team Leads view team participation)
- ✅ Bulk actions for Team Leads/Admin (mark group as opted out)
- ✅ Enhanced special day controls with notes
- ✅ Improved headcount reporting (by meal type, team, Office/WFH split)
- ✅ Live headcount updates without page refresh (Admin/Logistics views)
- ✅ Daily announcement draft generation
- ✅ Work location tracking (Office/WFH) per date
- ✅ Company-wide WFH period management

### In Scope (Iteration 3)

- ✅ Future meal participation planning (forward window, configurable)
- ✅ Headcount forecast snapshot for upcoming dates (Admin/Logistics)
- ✅ Event day scheduling via `day_schedules` (`event_day` status)
- ✅ Auditability: changer role visible in participation history
- ✅ Admin cross-user audit trail endpoint
- ✅ Monthly WFH usage summary per employee (soft limit: 5 days)
- ✅ Over-limit indicators in Team Lead / Admin reports
- ✅ Over-limit filter in team and admin WFH reports
- ✅ Policy refinements: forward window constraint, past-date guard

### Out of Scope (Future Iterations)

- ❌ Personal meal history view for users (future iteration)
- ❌ Menu planning and recipes
- ❌ Dietary restrictions tracking
- ❌ Mobile applications
- ❌ Email/SMS notifications
- ❌ Analytics dashboard
- ❌ "Favorite Meals" or meal preferences

---

## 6. Requirements

### Functional Requirements

**FR1: User Authentication**

- Users log in with email and password
- JWT-based session management (2-hour expiry)
- Role-based access control (Employee, Team Lead, Admin, Logistics)

**FR2: Default Meal Preferences**

- Users set default preference: opt-in (default) or opt-out
- Default applies to all meals unless explicitly overridden
- Users can change default preference at any time

**FR3: Daily Meal Participation**

- View current meal participation status (based on default + exceptions)
- Toggle participation for specific dates/meals
- Cutoff time enforcement (9:00 PM on the previous day for next-day meals)
- Visual indicators for participation status

**FR4: Participation Change History**

- Users can view their own participation change records (last 3 months)
- Track change source (self, team lead, admin) with timestamp

**FR5: Admin/Team Lead Features**

- Override employee participation (restricted by cutoff time; no changes allowed after cutoff)
- View team headcount
- Manage day schedules (mark days as holidays, celebrations, office closed)
- View audit trail with change attribution

**FR6: Day Schedule Management**

- Mark days as: Normal, Office Closed, Government Holiday, Celebration
- Define available meals for special days
- System prevents opt-in/out on office closed days

**FR7: Time-Based Headcount Calculation**

- Admins/Logistics view headcount calculated at request time per meal
- Breakdown: participating count, not participating count, total employees
- Filter by date and meal type

**FR8: Team-based Visibility**

- Employees can view which team they belong to
- Team Leads can view participation for their own team for the day
- Admin/Logistics can view participation across all teams
- Team information displayed on dashboard

**FR9: Bulk Actions for Team Lead/Admin**

- Team Leads can apply bulk opt-out for their team members
- Admins can apply bulk opt-out across teams
- Bulk actions support reason field (e.g., offsite, event)
- Audit trail for bulk actions logged per user with date range

**FR10: Enhanced Special Day Controls**

- Admin/Logistics can mark day as: Office Closed, Government Holiday, Special Celebration Day
- Special Celebration Day supports optional note field
- System automatically disables meal participation for Office Closed days
- System shows special day indicators on calendar

**FR11: Enhanced Headcount Reporting**

- Headcount breakdown by meal type
- Headcount breakdown by team
- Overall total headcount
- Office vs WFH split in headcount

**FR12: Live Headcount Updates**

- Headcount totals update in real-time without page refresh (Admin/Logistics views only)
- WebSocket or Server-Sent Events for live updates
- Updates triggered when any employee changes participation

**FR13: Daily Announcement Draft**

- Logistics/Admin can generate announcement message for selected date
- Message includes meal-wise participation totals
- Message highlights special day notes (holiday/office closed/celebration)
- Copy/paste-friendly format

**FR14: Work Location Tracking**

- Employees can set work location for a date: Office or WFH
- Default work location is Office (no user-level preference setting)
- Team Leads can correct work location for team members
- Admin can correct work location for any employee
- Work location visible in headcount reports

**FR15: Company-wide WFH Period**

- Admin/Logistics can declare date range as WFH for everyone
- During WFH period, all employees treated as WFH by default
- WFH period overrides individual work location settings
- WFH period visible on calendar with indicator

**FR16: Future Meal Participation Planning**

- Employees can set meal participation for future dates within a configurable forward window (default: 14 days ahead)
- Requests beyond the forward window are rejected with a clear error
- The forward window is server-configured (env var) and exposed to the frontend

**FR17: Event Meals**

- Admin/Logistics can create a day schedule with `day_status = event_day`
- Event day includes: date, available meals (e.g., `lunch`, `event_dinner`), optional reason/note
- Employees can opt in/out for event meals the same way as regular meals
- `event_dinner` MealType already exists; no new meal type needed

**FR18: Auditability and Accountability**

- Participation history already records `changed_by_user_id`; the changer's role must be surfaced in the API response
- Admin/Logistics can view a cross-user audit trail via a new endpoint
- Team Lead and Admin edits are identifiable by role in history records
- History records are append-only and immutable — no role (Employee, Team Lead, Admin or Logistics) may update or delete an existing history entry; the audit trail must remain tamper-proof

**FR19: Headcount Forecast Snapshot**

- Admin/Logistics can view upcoming headcount forecast for the next N days (default: 7, max: 14)
- Forecast returns one `DailyHeadcountSummary` per day, skipping days with no schedule or weekends
- Special day status and meal counts are included per day

**FR20: Monthly WFH Usage Summary**

- The system shows WFH days used per employee for a given month
- The standard allowance is 5 WFH days per month (configurable); entries beyond the allowance are still accepted
- Each employee summary returns: `used`, `allowance`, `is_over_limit`

**FR21: Over-Limit Indicators in Reports**

- Team Lead and Admin/Logistics views clearly highlight employees who exceed the monthly WFH allowance
- Reports include rollup fields: `employees_over_limit` (count) and `total_extra_wfh_days` (sum of excess)

**FR22: Over-Limit Filters**

- Team Lead and Admin/Logistics WFH report endpoints accept `?over_limit_only=true`
- When set, only employees exceeding the monthly allowance are returned

**FR23: Policy Refinements**

- Forward window enforced server-side (employees cannot set participation beyond `today + ForwardWindowDays`)
- Past-date guard: employees cannot set participation for dates before today via self-service
- Existing cutoff, RBAC, and schedule constraints remain unchanged

### Non-Functional Requirements

**NFR1: Performance**

- Page load time < 2 seconds
- API response time < 500ms (95th percentile)

**NFR2: Availability**

- 99% uptime
- Graceful degradation on failures

**NFR3: Security**

- HTTPS only
- JWT token expiry: 2 hours
- Password: bcrypt with cost 12
- RBAC enforcement at API level

**NFR4: Usability**

- Responsive design
- Accessibility: WCAG 2.1 AA compliance
- Intuitive UI requiring minimal training

**NFR4: Concurrent Capacity Estimate (Usage-Based)**

- Total users: ~100–120 employees
- Peak usage window: 10–15 minutes before cutoff
- Expected active users at peak: ~20–30
- Average request rate: ~1–2 req/sec per active user
- Target capacity: support ~30–40 concurrent active users comfortably

---
