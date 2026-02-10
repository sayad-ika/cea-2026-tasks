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

---
