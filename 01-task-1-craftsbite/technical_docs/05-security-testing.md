## 10. Security and Access Control

### Authentication

- **Method:** JWT-based authentication
- **Token Storage:** localStorage (client-side)
- **Token Expiry:** 2-4 hours
- **Password Hashing:** bcrypt with cost factor 12
- **HTTPS Only:** All communications encrypted in transit

### Authorization (RBAC)

| Role          | Permissions                                                                                                                                                    |
| ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Employee**  | View/update own preferences, view own participation, manage own bulk opt-outs, view own history, view own team membership, set own work location               |
| **Team Lead** | All employee permissions + override team member participation, view team headcount, apply bulk actions for team, correct team member work location             |
| **Admin**     | All team lead permissions + manage day schedules, view all users, system-wide headcount, manage WFH periods, generate announcements, bulk actions across teams |
| **Logistics** | View headcounts, view schedules, generate announcements, view all teams participation                                                                          |

### Middleware Stack

1. **CORS** - Whitelist allowed origins
2. **Request ID** - Unique ID for request tracing
3. **Authentication** - JWT validation
4. **Authorization** - Role-based access checks
5. **Rate Limiting** - 100 requests/minute per user
6. **Recovery** - Graceful panic recovery

### Security Best Practices

- Input validation on all API endpoints (using go-playground/validator)
- SQL injection prevention (GORM parameterized queries)
- XSS prevention (React's built-in escaping)
- CSRF protection (SameSite cookie attribute)
- Audit logging for sensitive operations
- Password complexity requirements (min 8 chars, mix of upper/lower/numbers)

### Data Privacy

- Participation history upto 1 year
- No PII beyond name and email
- Audit logs show "who changed what" for accountability

---

## 11. Testing Plan

### Unit Testing

**Backend (Go)**

- Framework: Go's built-in `testing` package + testify
- Coverage Target: >80%
- Focus Areas:
  - Business logic in service layer
  - Participation resolution algorithm
  - Date/time calculations
  - Permission checks

**Frontend (React)**

- Framework: Vitest + React Testing Library
- Coverage Target: >70%
- Focus Areas:
  - Component rendering
  - User interactions
  - Form validation
  - State management

### Integration Testing

**API Testing**

- Framework: Go's httptest + testify
- Test Scenarios:
  - Full request/response cycles
  - Middleware chain
  - Database interactions
  - Error handling

**Database Testing**

- Framework: testcontainers-go (PostgreSQL in Docker)
- Test Scenarios:
  - Schema migrations
  - Complex queries
  - Transaction rollbacks
  - Constraint validation

### Integration Scenarios

#### No external system integrations are implemented in iteration 1.

All scenarios documented in this section focus on internal application behavior and data flows.
External integrations (e.g., email/SMS notifications, third-party vendors, SSO) will be introduced
and documented in future iterations if required.

### Manual QA Checklist

#### Pre-Test Setup

- [ ] Fresh database with test data (10 users, various roles)
- [ ] System time set to known date (e.g., Feb 10, 2026, 8:00 AM)
- [ ] Test user credentials available for each role
- [ ] Browser cache cleared

---

#### Test Suite 1: User Authentication & Authorization

**TC1.1: Successful Login**

- [ ] Navigate to login page (`/login`)
- [ ] Enter valid employee credentials
- [ ] Click "Login" button
- [ ] **Verify:** Redirected to dashboard
- [ ] **Verify:** User name displayed in header
- [ ] **Verify:** JWT token stored in localStorage

**TC1.2: Failed Login - Invalid Credentials**

- [ ] Enter incorrect password
- [ ] Click "Login"
- [ ] **Verify:** Error message "Invalid credentials" displayed
- [ ] **Verify:** No redirect occurs
- [ ] **Verify:** No token stored

**TC1.3: Session Persistence**

- [ ] Login successfully
- [ ] Refresh browser page
- [ ] **Verify:** User remains logged in
- [ ] **Verify:** Dashboard still accessible

**TC1.4: Logout**

- [ ] Click "Logout" button
- [ ] **Verify:** Redirected to login page
- [ ] **Verify:** Token removed from localStorage
- [ ] Attempt to access `/meals` directly
- [ ] **Verify:** Redirected to login page

**TC1.5: Role-Based Access Control**

- [ ] Login as Employee
- [ ] Attempt to access `/admin/headcount`
- [ ] **Verify:** 403 error or redirect to dashboard
- [ ] Logout and login as Admin
- [ ] Access `/admin/headcount`
- [ ] **Verify:** Page loads successfully

---

#### Test Suite 2: Default Meal Preference

**TC2.1: View Current Default Preference**

- [ ] Login as employee
- [ ] Navigate to Settings/Preferences page
- [ ] **Verify:** Current default displayed (e.g., "Opt-In")
- [ ] **Verify:** Radio buttons or toggle reflects current setting

**TC2.2: Change Default from Opt-In to Opt-Out**

- [ ] Select "Opt-Out" option
- [ ] Click "Save" button
- [ ] **Verify:** Success message displayed
- [ ] Refresh page
- [ ] **Verify:** Selection persisted to "Opt-Out"

**TC2.3: Default Preference Applied to Future Meals**

- [ ] Set default to "Opt-Out"
- [ ] Navigate to meal calendar
- [ ] View tomorrow's meals (no specific records)
- [ ] **Verify:** All meals show "Not Participating" status
- [ ] Change default to "Opt-In"
- [ ] Refresh calendar
- [ ] **Verify:** All meals now show "Participating" status

---

#### Test Suite 3: Daily Meal Participation

**TC3.1: View Current Week Meal Calendar**

- [ ] Navigate to `/meals` or meal calendar page
- [ ] **Verify:** Current week displayed with all days
- [ ] **Verify:** Each day shows available meals (lunch, snacks, dinner)
- [ ] **Verify:** Participation status visible for each meal

**TC3.2: Toggle Participation (Before Cutoff)**

- [ ] Set system time to 8:30 AM
- [ ] Find today's lunch meal
- [ ] Click toggle/checkbox to opt-out
- [ ] **Verify:** Status changes to "Not Participating"
- [ ] **Verify:** Success toast notification appears
- [ ] Refresh page
- [ ] **Verify:** Change persisted

**TC3.3: Attempt Toggle After Cutoff**

- [ ] Set system time to 9:30 AM (after lunch cutoff)
- [ ] Attempt to toggle today's lunch
- [ ] **Verify:** Error message "Cutoff time has passed"
- [ ] **Verify:** Status does not change
- [ ] **Verify:** Toggle/button disabled or warning shown

**TC3.4: Toggle for Future Date**

- [ ] Select tomorrow's lunch
- [ ] Toggle participation
- [ ] **Verify:** Change allowed (no cutoff restriction)
- [ ] **Verify:** Status updated successfully

**TC3.5: Visual Indicators**

- [ ] **Verify:** "Participating" meals have green checkmark or highlight
- [ ] **Verify:** "Not Participating" meals have red X or different styling
- [ ] **Verify:** Past dates are greyed out or read-only
- [ ] **Verify:** Cutoff-passed meals show lock icon or disabled state

---

#### Test Suite 4: Date-Range Participation Rule Management

**TC4.1: Create Date-Range Participation Rule for Vacation**

- [ ] Navigate to "Date-Range Participation Rule" or "Manage Exceptions" page
- [ ] Click "Create Date-Range Participation Rule" button
- [ ] Enter start date: Feb 20, 2026
- [ ] Enter end date: Feb 25, 2026
- [ ] Enter reason: "Vacation"
- [ ] Click "Submit"
- [ ] **Verify:** Success message displayed
- [ ] **Verify:** New entry appears in date-range participation list

**TC4.2: Verify Date-Range Participation Rule Applied to Calendar**

- [ ] Navigate to meal calendar
- [ ] View Feb 20-25 date range
- [ ] **Verify:** All meals in range show "Not Participating"
- [ ] **Verify:** Visual indicator (e.g., "Date-Range Participation Rule" label) shown
- [ ] Hover or click for details
- [ ] **Verify:** Reason "Vacation" displayed

**TC4.3: View All Active Date-Range Participations**

- [ ] Return to date-range participation page
- [ ] **Verify:** List shows all active/upcoming bulk opt-outs
- [ ] **Verify:** Each entry shows: date range, reason, created date
- [ ] **Verify:** Past bulk opt-outs hidden or in separate section

**TC4.4: Delete Date-Range Participation Rule**

- [ ] Locate previously created date-range participation
- [ ] Click "Delete" or trash icon
- [ ] Confirm deletion in modal
- [ ] **Verify:** Entry removed from list
- [ ] Navigate to calendar for affected dates
- [ ] **Verify:** Meals revert to default preference status

**TC4.5: Overlapping Date-Range Participation (Edge Case)**

- [ ] Create date-range participation: Feb 10-15
- [ ] Create antoher: Feb 12-18
- [ ] **Verify:** Both saved successfully
- [ ] Check calendar for Feb 12-15 (overlap)
- [ ] **Verify:** Meals still show "Not Participating"
- [ ] Delete first date-range participation
- [ ] **Verify:** Feb 12-18 still opted-out (second rule applies)

---

#### Test Suite 5: Participation History

**TC5.1: View Personal History**

- [ ] Navigate to "My History" or history tab
- [ ] **Verify:** List of recent participation changes displayed
- [ ] **Verify:** Each entry shows: date, meal type, old status, new status, timestamp
- [ ] **Verify:** Changes from last 3 months visible

**TC5.2: History After Manual Toggle**

- [ ] Toggle tomorrow's lunch (opt-out)
- [ ] Navigate to history page
- [ ] **Verify:** New entry at top of list
- [ ] **Verify:** Shows: changed_by = self, change_type = "manual"

**TC5.3: History After Admin Override**

- [ ] (Admin account) Override user's meal participation
- [ ] (Employee account) View history
- [ ] **Verify:** Entry shows changed_by = admin name
- [ ] **Verify:** Change type = "override" or similar indicator

**TC5.4: Filter/Sort History**

- [ ] Use date filter to show last week only
- [ ] **Verify:** Only relevant entries shown
- [ ] Click column header to sort by date
- [ ] **Verify:** List reorders correctly

---

#### Test Suite 6: Team Lead Override

**TC6.1: View Team Members**

- [ ] Login as Team Lead
- [ ] Navigate to "Team Management" page
- [ ] **Verify:** List of team members displayed
- [ ] **Verify:** Each member shows current meal participation status

**TC6.2: Override Team Member Participation**

- [ ] Select team member "John Doe"
- [ ] View his Feb 15 lunch status (currently "Participating")
- [ ] Click "Override" button
- [ ] Toggle to "Not Participating"
- [ ] Add reason: "Team meeting scheduled"
- [ ] Click "Save Override"
- [ ] **Verify:** Success message displayed
- [ ] **Verify:** John's status updated immediately

**TC6.3: Verify Override Persisted**

- [ ] Refresh page
- [ ] **Verify:** John's Feb 15 lunch still shows "Not Participating"
- [ ] **Verify:** Override indicator visible (e.g., "Overridden by Team Lead")
- [ ] (Login as John Doe) View his own calendar
- [ ] **Verify:** Feb 15 lunch shows "Not Participating" with override notice

**TC6.4: Cannot Override Outside Team**

- [ ] Attempt to access another team's member
- [ ] **Verify:** User not in list OR
- [ ] Attempt override via API directly
- [ ] **Verify:** 403 Forbidden error

---

#### Test Suite 7: Admin Headcount View

**TC7.1: View Today's Headcount**

- [ ] Login as Admin or Logistics
- [ ] Navigate to "Headcount" page
- [ ] Select today's date
- [ ] Select meal type: "Lunch"
- [ ] Click "View Headcount"
- [ ] **Verify:** Summary displays:
  - Total employees: [number]
  - Participating: [number]
  - Not participating: [number]
- [ ] **Verify:** Numbers add up correctly

**TC7.2: Drill Down to User List**

- [ ] Click "View Details" or expand participating users
- [ ] **Verify:** List of participating employees shown
- [ ] **Verify:** List shows user names
- [ ] Click "Not Participating" tab
- [ ] **Verify:** List of opted-out users shown

**TC7.3: Export Headcount Report**

- [ ] Click "Export" or "Download CSV" button
- [ ] **Verify:** File downloads successfully
- [ ] Open CSV file
- [ ] **Verify:** Contains: date, meal type, user names, participation status

**TC7.4: Headcount for Future Date**

- [ ] Select date: Feb 25, 2026
- [ ] View headcount
- [ ] **Verify:** Shows projected numbers based on current opt-outs/preferences
- [ ] Create new date-range participation affecting Feb 25
- [ ] Refresh headcount
- [ ] **Verify:** Numbers updated to reflect new opt-out

---

#### Test Suite 8: Day Schedule Management

**TC8.1: Mark Day as Government Holiday**

- [ ] Login as Admin
- [ ] Navigate to "Day Schedules" page
- [ ] Select date: Feb 26, 2026
- [ ] Set status: "Government Holiday"
- [ ] Enter reason: "Independence Day"
- [ ] Set available meals: (none)
- [ ] Click "Save"
- [ ] **Verify:** Success message displayed

**TC8.2: Verify Holiday Blocks Participation**

- [ ] (Login as Employee)
- [ ] Navigate to calendar for Feb 26
- [ ] **Verify:** Day marked as "Government Holiday - Independence Day"
- [ ] **Verify:** No meal toggles available
- [ ] Attempt to opt-in via API (if accessible)
- [ ] **Verify:** 400/403 error returned

**TC8.3: View Headcount for Holiday**

- [ ] (Login as Admin) View headcount for Feb 26
- [ ] **Verify:** Shows 0 participating
- [ ] **Verify:** Message "Office closed - no meals scheduled"

**TC8.4: Change Day Status Back to Normal**

- [ ] Edit Feb 26 schedule
- [ ] Change status to "Normal"
- [ ] **Verify:** Employees can now manage participation for that day

**TC8.5: Mark Day as Celebration (Special Meals)**

- [ ] Select date: Feb 28
- [ ] Set status: "Celebration"
- [ ] Reason: "Company Anniversary"
- [ ] Available meals: ["lunch", "event_dinner"]
- [ ] Save
- [ ] **Verify:** Calendar shows special indicator
- [ ] **Verify:** Only lunch and event dinner available for opt-in

---

#### Test Suite 9: UI/UX Validation

**TC9.1: Responsive Design - Mobile**

- [ ] Resize browser to mobile width (375px)
- [ ] **Verify:** Navigation collapses to hamburger menu
- [ ] **Verify:** Calendar cards stack vertically
- [ ] **Verify:** All buttons accessible and tappable
- [ ] **Verify:** No horizontal scroll on any page

**TC9.2: Responsive Design - Tablet**

- [ ] Resize to tablet width (768px)
- [ ] **Verify:** Layout adapts appropriately
- [ ] **Verify:** Two-column layouts where applicable

**TC9.3: Loading States**

- [ ] Throttle network to slow 3G
- [ ] Navigate to headcount page
- [ ] **Verify:** Loading spinner displayed while data fetches
- [ ] **Verify:** UI doesn't show stale data during load

**TC9.4: Error Handling**

- [ ] Disconnect network
- [ ] Attempt to toggle participation
- [ ] **Verify:** Error message "Network error - please try again"
- [ ] **Verify:** No silent failure

**TC9.5: Accessibility - Keyboard Navigation**

- [ ] Use Tab key to navigate entire page
- [ ] **Verify:** Focus indicator visible on each element
- [ ] **Verify:** All interactive elements reachable
- [ ] Press Enter/Space on toggles
- [ ] **Verify:** Actions trigger correctly

**TC9.6: Accessibility - Screen Reader**

- [ ] Enable screen reader (NVDA/JAWS)
- [ ] Navigate participation calendar
- [ ] **Verify:** Each meal status announced correctly
- [ ] **Verify:** Button labels descriptive ("Opt out of lunch on Feb 10")

---

#### Test Suite 10: Edge Cases & Error Scenarios

**TC10.1: Invalid Date Selection**

- [ ] Attempt to create date-range participation with end date before start date
- [ ] **Verify:** Validation error displayed
- [ ] **Verify:** Form submission blocked

**TC10.2: Special Characters in Reason Field**

- [ ] Enter reason: `<script>alert('XSS')</script>`
- [ ] Save date-range participation
- [ ] View reason in list
- [ ] **Verify:** Displayed as plain text (no script execution)

**TC10.3: Simultaneous Admin Override**

- [ ] Two admins override same user's meal simultaneously
- [ ] **Verify:** Both changes processed (last-write-wins) OR
- [ ] **Verify:** Second admin sees "conflict" warning

**TC10.4: Token Expiration During Session**

- [ ] Login and let session sit for 24+ hours (or manipulate token expiry)
- [ ] Attempt to toggle participation
- [ ] **Verify:** 401 Unauthorized response
- [ ] **Verify:** Redirected to login page
- [ ] **Verify:** User-friendly message: "Session expired - please login again"

**TC10.5: SQL Injection Attempt**

- [ ] Enter email: `admin@test.com' OR '1'='1`
- [ ] Attempt login
- [ ] **Verify:** Login fails (parameterized queries prevent injection)
- [ ] **Verify:** No database error exposed to user

**TC10.6: Extremely Long Reason Text**

- [ ] Enter 1000+ character reason for date-range participation
- [ ] Attempt save
- [ ] **Verify:** Validation error or text truncated with warning

---

#### Test Suite 11: Data Integrity & Audit

**TC11.1: History Accuracy**

- [ ] Make 5 participation changes throughout the day
- [ ] View history at end of day
- [ ] **Verify:** All 5 changes logged in correct chronological order
- [ ] **Verify:** Timestamps accurate

**TC11.2: Default Preference Change Logging**

- [ ] Change default from opt-in to opt-out
- [ ] Check if this change is logged
- [ ] **Verify:** System behavior documented (may or may not log preference changes)

**TC11.3: Date-Range Participation Rule Audit Trail**

- [ ] Create date-range participation
- [ ] Delete it
- [ ] Check if deletion logged
- [ ] **Verify:** Audit trail shows creation and deletion events

---

#### Test Suite 12: Team-based Visibility

**TC12.1: Employee Views Team Membership**

- [ ] Login as employee
- [ ] View dashboard
- [ ] **Verify:** Team name displayed
- [ ] **Verify:** Team lead name displayed

**TC12.2: Team Lead Views Team Participation**

- [ ] Login as Team Lead
- [ ] Navigate to Team View
- [ ] **Verify:** All team members listed
- [ ] **Verify:** Meal Participation status for each member
- [ ] **Verify:** Cannot view other teams

**TC12.3: Admin Views All Teams**

- [ ] Login as Admin
- [ ] Navigate to Team
- [ ] **Verify:** All teams visible

**TC12.4: Logistics Views All Teams**

- [ ] Login as Logistics
- [ ] Navigate to Headcount
- [ ] **Verify:** Can view all teams participation
- [ ] **Verify:** Can generate announcements

---

#### Test Suite 13: Bulk Actions

**TC13.1: Team Lead Bulk Opt-Out**

- [ ] Login as Team Lead
- [ ] Select multiple team members
- [ ] Apply bulk opt-out for date range
- [ ] Enter reason
- [ ] **Verify:** All selected members opted out
- [ ] **Verify:** Audit log created per user with date range

**TC13.2: Admin Bulk Opt-Out Across Teams**

- [ ] Login as Admin
- [ ] Select users from multiple teams
- [ ] Apply bulk opt-out
- [ ] **Verify:** Action applied to all selected users

**TC13.3: Bulk Action Audit Trail**

- [ ] Perform bulk opt-out action
- [ ] View audit log
- [ ] **Verify:** Each user has individual entry with date range
- [ ] **Verify:** Reason recorded

---

#### Test Suite 14: Work Location

**TC14.1: Employee Sets Work Location**

- [ ] Login as employee
- [ ] Set location to WFH
- [ ] **Verify:** Location saved
- [ ] View headcount as Admin or Team Lead
- [ ] **Verify:** WFH count includes user

**TC14.2: Employee Changes Work Location**

- [ ] Set work location to Office for a date
- [ ] Change to WFH for same date
- [ ] **Verify:** Location updated

**TC14.3: Team Lead Corrects Work Location**

- [ ] Login as Team Lead
- [ ] View team member work location
- [ ] Correct missing entry
- [ ] **Verify:** Change saved with audit

**TC14.4: Default Work Location**

- [ ] Check employee without explicit work location
- [ ] **Verify:** Default is Office

---

#### Test Suite 15: Live Updates

**TC15.1: Headcount Updates Live**

- [ ] Open headcount page as Admin
- [ ] Open second browser tab as Employee
- [ ] Employee changes meal participation
- [ ] **Verify:** Admin headcount totals update without page refresh

**TC15.2: SSE Connection Handling**

- [ ] Open Admin headcount page
- [ ] Open browser DevTools → Network tab → filter by `EventSource`
- [ ] **Verify:** `headcount/report/live` connection shows status `101` / pending (open)
- [ ] Close browser tab
- [ ] **Verify:** Server cleans up connection (no goroutine leak)

**TC15.3: Live Updates Scope**

- [ ] Login as Employee → view meal calendar
- [ ] **Verify:** No SSE connection to `headcount/report/live` in Network tab
- [ ] Login as Logistics → view headcount page
- [ ] **Verify:** SSE connection established and live updates received
- [ ] Login as Admin → view headcount page
- [ ] **Verify:** SSE connection established and live updates received

**TC15.4: Admin Override Triggers Live Update**

- [ ] Open headcount page as Admin (tab 1)
- [ ] Override an employee's participation as Admin (tab 2)
- [ ] **Verify:** Headcount in tab 1 updates without refresh

**TC15.5: SSE Reconnection**

- [ ] Open headcount page as Admin
- [ ] **Verify:** SSE connection established
- [ ] Restart the backend server
- [ ] **Verify:** Browser automatically reconnects within ~3 seconds
- [ ] **Verify:** Live updates resume after reconnection

---

#### Test Suite 16: Daily Announcement

**TC16.1: Generate Announcement**

- [ ] Login as Logistics
- [ ] Navigate to Announcement page
- [ ] Select date
- [ ] **Verify:** Announcement generated with meal totals
- [ ] **Verify:** Copy button works

**TC16.2: Announcement with Special Day**

- [ ] Mark date as Special Celebration
- [ ] Generate announcement
- [ ] **Verify:** Celebration note included

**TC16.3: Announcement with Office Closed**

- [ ] Mark date as Office Closed
- [ ] Generate announcement
- [ ] **Verify:** Announcement indicates office closed

**TC16.4: Announcement with WFH Breakdown**

- [ ] Set some employees as WFH for date
- [ ] Generate announcement
- [ ] **Verify:** Office vs WFH breakdown included

---

#### Test Suite 17: Company-wide WFH Period

**TC17.1: Create WFH Period**

- [ ] Login as Admin
- [ ] Create WFH period for date range
- [ ] **Verify:** Period created successfully
- [ ] View calendar
- [ ] **Verify:** WFH indicator shown

**TC17.2: WFH Period Overrides Individual Location**

- [ ] Create WFH period
- [ ] Employee sets Office location for same date
- [ ] View headcount
- [ ] **Verify:** Employee counted as WFH

**TC17.3: WFH Period Visibility**

- [ ] Create WFH period
- [ ] Login as Employee
- [ ] View calendar
- [ ] **Verify:** WFH period indicator visible

**TC17.4: Delete WFH Period**

- [ ] Delete active WFH period
- [ ] **Verify:** Period removed
- [ ] **Verify:** Individual work locations respected again

---

#### Test Suite 18: Enhanced Headcount Reporting

**TC18.1: Headcount by Meal Type**

- [ ] Login as Admin
- [ ] View headcount for date
- [ ] **Verify:** Breakdown by meal type (breakfast, lunch, dinner)

**TC18.2: Headcount by Team**

- [ ] View headcount
- [ ] **Verify:** Team-wise breakdown available
- [ ] Filter by specific team
- [ ] **Verify:** Filtered results correct

**TC18.3: Office vs WFH Split**

- [ ] Set some employees as WFH
- [ ] View headcount
- [ ] **Verify:** Office count and WFH count displayed
- [ ] **Verify:** Totals add up correctly

---
