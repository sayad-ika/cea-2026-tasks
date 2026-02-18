## 7. User Flows

### Flow 1: Employee Sets Default Preference

1. Employee logs in
2. Navigates to Settings
3. Selects default meal preference (opt-in or opt-out)
4. Saves changes
5. System updates `users.default_meal_preference`
6. Confirmation displayed

### Flow 2: Employee Manages Daily Meals

1. Employee views meal calendar
2. System shows participation status (based on: day schedule → specific record → bulk opt-out → default preference)
3. Employee toggles participation for specific date/meal
4. System checks cutoff time
5. If before cutoff: update `meal_participations` table, log to `meal_participation_history`
6. If after cutoff: show error message
7. Display updated status

### Flow 3: Employee Manages Date-Range Participation

1. Employee navigates to "Participation Rules" or "Exceptions" section
2. Selects date range (start date, end date)
3. System creates a date-range participation rule
4. System applies the rule to all meals within the range
5. Confirmation with summary displayed

### Flow 4: Team Lead Overrides Participation

1. Team Lead views team member list
2. Selects member and date
3. Attempts to toggle participation status
4. System validates cutoff time (changes allowed only before cutoff)
5. If valid, system updates `meal_participations` with `overridden_by` field
6. Logs change to `meal_participation_history`
7. Notification shown

### Flow 5: Admin Views Headcount

1. Admin navigates to Headcount page
2. Selects date and meal type
3. System calculates:
   - Checks `day_schedules` (if office closed, return 0)
   - Queries `meal_participations` for explicit records
   - For users without explicit records, checks `bulk_opt_outs`
   - For users without records or bulk opt-outs, uses `default_meal_preference`
4. Returns aggregated counts
5. Display breakdown with drill-down capability

### Flow 6: Employee Views Team Information

1. Employee logs in
2. Dashboard displays team membership
3. Employee can view team name and team lead
4. System shows team-specific information

### Flow 7: Team Lead Views Team Participation

1. Team Lead navigates to Team View
2. System displays team members list
3. Team Lead selects date
4. System shows participation status for all team members
5. Headcount summary displayed

### Flow 8: Team Lead Applies Bulk Opt-Out

1. Team Lead selects team members (multiple)
2. Selects date range
3. Enters reason (e.g., Team offsite)
4. Confirms bulk action
5. System updates participation for all selected members
6. Audit log entry created per user with date range

### Flow 9: Admin Generates Daily Announcement

1. Admin navigates to Announcement page
2. Selects date
3. System generates announcement draft with:
   - Meal-wise participation totals
   - Special day notes if applicable
   - Office/WFH breakdown
4. Admin copies message for distribution

### Flow 10: Employee Sets Work Location

1. Employee navigates to calendar
2. Selects date
3. Chooses work location: Office or WFH
4. System saves work location
5. Headcount reports reflect work location

### Flow 11: Admin Declares Company-wide WFH Period

1. Admin navigates to WFH Period management
2. Creates new WFH period
3. Enters start date and end date
4. Optionally enters reason
5. System applies WFH status to all employees for period
6. Calendar shows WFH period indicator

---
