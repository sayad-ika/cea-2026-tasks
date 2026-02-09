## 10. Security and Access Control

### Authentication
- **Method:** JWT-based authentication
- **Token Storage:** localStorage (client-side)
- **Token Expiry:** 2-4 hours
- **Password Hashing:** bcrypt with cost factor 12
- **HTTPS Only:** All communications encrypted in transit

### Authorization (RBAC)

| Role | Permissions |
|------|-------------|
| **Employee** | View/update own preferences, view own participation, manage own bulk opt-outs, view own history |
| **Team Lead** | All employee permissions + override team member participation, view team headcount |
| **Admin** | All team lead permissions + manage day schedules, view all users, system-wide headcount |
| **Logistics** | View headcounts, view schedules (read-only for planning) |

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
- Participation history limited to 3 months
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

**Scenario 1: Authentication & Authorization Flow**
- **Test:** User login → JWT generation → Protected resource access
- **Steps:**
  1. POST /api/auth/login with valid credentials
  2. Verify JWT token returned with correct claims (user_id, role, expiry)
  3. Use JWT to access protected endpoint (GET /api/meals/participation)
  4. Verify authorization middleware allows access based on role
  5. Attempt access with expired token → verify 401 response
  6. Attempt access to admin endpoint as employee → verify 403 response
- **Expected:** Proper authentication and role-based access control enforced

**Scenario 2: Participation Resolution with Multiple Factors**
- **Test:** System correctly resolves participation status with overlapping rules
- **Steps:**
  1. Set user default preference to "opt-in"
  2. Create bulk opt-out for Feb 10-15
  3. Create specific participation record (opt-in) for Feb 12
  4. Query participation status for Feb 12
  5. Verify specific record overrides bulk opt-out
  6. Query participation for Feb 13 (no specific record)
  7. Verify bulk opt-out applies (not participating)
- **Expected:** Priority order respected: specific record > bulk opt-out > default preference

**Scenario 3: Bulk Opt-Out Creation and Cascade**
- **Test:** Bulk opt-out affects all meals in date range
- **Steps:**
  1. User has default preference "opt-in"
  2. POST /api/bulk-opt-outs with date range Feb 20-25
  3. Verify bulk_opt_outs record created
  4. Query participation for Feb 21, Feb 23 (different meals: lunch, dinner)
  5. Verify all meals marked as "not participating"
  6. DELETE bulk opt-out
  7. Re-query participation → verify reverts to default (opt-in)
- **Expected:** Bulk opt-out correctly applied and removed across all meals

**Scenario 4: Day Schedule Impact on Participation**
- **Test:** Office closed day prevents participation
- **Steps:**
  1. Admin creates day schedule: Feb 18 = "office_closed"
  2. Employee attempts to opt-in for Feb 18 lunch
  3. Verify API returns error or blocks action
  4. Query headcount for Feb 18 → verify returns 0
  5. Admin changes day to "normal"
  6. Re-query → verify participation allowed
- **Expected:** Day schedule takes highest priority, blocking participation

**Scenario 5: Team Lead Override with Audit Trail**
- **Test:** Override creates proper audit log
- **Steps:**
  1. Employee has participation = true for Feb 25 lunch
  2. Team lead executes PUT /api/admin/override-participation (change to false)
  3. Verify meal_participations updated with overridden_by = team_lead_id
  4. Query meal_participation_history
  5. Verify history record shows: old_value=true, new_value=false, changed_by=team_lead_id
  6. Employee views own history → verify change visible
- **Expected:** Override persisted, audit trail complete, employee notified

**Scenario 6: Cutoff Time Enforcement**
- **Test:** System blocks changes after cutoff time
- **Steps:**
  1. Set server time to 8:55 AM (before lunch cutoff)
  2. Employee toggles lunch participation for today
  3. Verify update succeeds
  4. Set server time to 9:05 AM (after cutoff)
  5. Employee attempts same action
  6. Verify API returns 400/403 with "cutoff time passed" message
- **Expected:** Cutoff time strictly enforced, clear error messages

**Scenario 7: Default Preference Change Retroactive Effect**
- **Test:** Changing default doesn't affect past specific records
- **Steps:**
  1. User default = "opt-in"
  2. Create specific opt-out for Feb 10
  3. Query Feb 10 → verify not participating
  4. Change default to "opt-out"
  5. Re-query Feb 10 → verify still not participating (specific record preserved)
  6. Query Feb 11 (no specific record) → verify now uses new default (opt-out)
- **Expected:** Default change only affects future/unspecified dates

**Scenario 8: Headcount Calculation Accuracy**
- **Test:** Admin headcount view shows correct aggregated data
- **Steps:**
  1. Setup: 10 users total
     - 3 users: default "opt-in", no exceptions
     - 2 users: default "opt-out", no exceptions
     - 3 users: default "opt-in", bulk opt-out for Feb 15
     - 2 users: default "opt-in", specific opt-out for Feb 15
  2. Admin queries GET /api/admin/headcount?date=2026-02-15&mealType=lunch
  3. Verify calculation:
     - Participating: 3 (only those with default opt-in, no exceptions)
     - Not participating: 7 (2 default opt-out + 3 bulk + 2 specific)
     - Total: 10
- **Expected:** Headcount matches manual calculation

**Scenario 9: History Auto-Cleanup Job**
- **Test:** Scheduled job deletes old records
- **Steps:**
  1. Insert test history records with created_at dates:
     - Record A: 2 months ago
     - Record B: 3 months ago
     - Record C: 4 months ago
  2. Trigger cleanup job (or simulate cron)
  3. Query meal_participation_history
  4. Verify Records A and B exist, Record C deleted
  5. Verify user can still view recent history (A, B)
- **Expected:** Only records >3 months deleted, recent history preserved

**Scenario 10: Concurrent Update Handling**
- **Test:** Database handles simultaneous participation updates
- **Steps:**
  1. User A and Admin B both update same meal participation simultaneously
  2. Verify database constraint (UNIQUE user_id, date, meal_type) prevents conflicts
  3. Verify one transaction succeeds, other fails with proper error
  4. Verify successful update logged to history
- **Expected:** Data integrity maintained, last-write-wins or proper error handling

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

#### Test Suite 4: Bulk Opt-Out Management

**TC4.1: Create Bulk Opt-Out for Vacation**
- [ ] Navigate to "Bulk Opt-Out" or "Manage Exceptions" page
- [ ] Click "Create Bulk Opt-Out" button
- [ ] Enter start date: Feb 20, 2026
- [ ] Enter end date: Feb 25, 2026
- [ ] Enter reason: "Vacation"
- [ ] Click "Submit"
- [ ] **Verify:** Success message displayed
- [ ] **Verify:** New entry appears in bulk opt-out list

**TC4.2: Verify Bulk Opt-Out Applied to Calendar**
- [ ] Navigate to meal calendar
- [ ] View Feb 20-25 date range
- [ ] **Verify:** All meals in range show "Not Participating"
- [ ] **Verify:** Visual indicator (e.g., "Bulk Opt-Out" label) shown
- [ ] Hover or click for details
- [ ] **Verify:** Reason "Vacation" displayed

**TC4.3: View All Active Bulk Opt-Outs**
- [ ] Return to bulk opt-out page
- [ ] **Verify:** List shows all active/upcoming bulk opt-outs
- [ ] **Verify:** Each entry shows: date range, reason, created date
- [ ] **Verify:** Past bulk opt-outs hidden or in separate section

**TC4.4: Delete Bulk Opt-Out**
- [ ] Locate previously created bulk opt-out
- [ ] Click "Delete" or trash icon
- [ ] Confirm deletion in modal
- [ ] **Verify:** Entry removed from list
- [ ] Navigate to calendar for affected dates
- [ ] **Verify:** Meals revert to default preference status

**TC4.5: Overlapping Bulk Opt-Outs (Edge Case)**
- [ ] Create bulk opt-out: Feb 10-15
- [ ] Create another: Feb 12-18
- [ ] **Verify:** Both saved successfully
- [ ] Check calendar for Feb 12-15 (overlap)
- [ ] **Verify:** Meals still show "Not Participating"
- [ ] Delete first bulk opt-out
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
- [ ] Create new bulk opt-out affecting Feb 25
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
- [ ] Attempt to create bulk opt-out with end date before start date
- [ ] **Verify:** Validation error displayed
- [ ] **Verify:** Form submission blocked

**TC10.2: Special Characters in Reason Field**
- [ ] Enter reason: `<script>alert('XSS')</script>`
- [ ] Save bulk opt-out
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
- [ ] Enter 1000+ character reason for bulk opt-out
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

**TC11.3: Bulk Opt-Out Audit Trail**
- [ ] Create bulk opt-out
- [ ] Delete it
- [ ] Check if deletion logged
- [ ] **Verify:** Audit trail shows creation and deletion events

---
