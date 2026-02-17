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

### Out of Scope (Future Iterations)
- ❌ Bulk opt-out for date ranges (vacations, business trips)
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

**FR8: Day-Level Participation Management** (*future iteration)
- Manage participation status for date ranges (e.g., office travel, leave)
- View all active date-range participation rules
- Remove or update participation rules
- Date-range rules override user default preferences

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
