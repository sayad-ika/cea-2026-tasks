## 5. Scope of Changes

### In Scope (Iteration 1)
- ✅ User authentication with JWT and role-based access control
- ✅ User default meal preference (opt-in or opt-out)
- ✅ Daily meal participation management with flexible defaults
- ✅ Bulk opt-out for date ranges (vacations, business trips)
- ✅ Real-time headcount calculation
- ✅ Admin/Team Lead override capabilities
- ✅ Day schedule management (holidays, celebrations, office closures)
- ✅ Comprehensive participation history tracking
- ✅ Personal meal history view for users
- ✅ Automatic 3-month audit trail cleanup
- ✅ Change tracking for accountability

### Out of Scope (Future Iterations)
- ❌ Menu planning and recipes
- ❌ Dietary restrictions tracking
- ❌ Mobile applications
- ❌ Email/SMS notifications
- ❌ Analytics dashboard

---

## 6. Requirements

### Functional Requirements

**FR1: User Authentication**
- Users log in with email and password
- JWT-based session management (24-hour expiry)
- Role-based access control (Employee, Team Lead, Admin, Logistics)

**FR2: Default Meal Preferences**
- Users set default preference: opt-in (default) or opt-out
- Default applies to all meals unless explicitly overridden
- Users can change default preference at any time

**FR3: Daily Meal Participation**
- View current meal participation status (based on default + exceptions)
- Toggle participation for specific dates/meals
- Cutoff time enforcement (9 AM for lunch/snacks)
- Visual indicators for participation status

**FR4: Bulk Opt-Out Management**
- Create opt-out for date ranges (e.g., vacation Feb 10-20)
- View all active bulk opt-outs
- Delete bulk opt-outs
- Bulk opt-outs override default preference

**FR5: Participation History**
- Users view their own participation changes (last 3 months)
- Track who made changes and when
- Automatic cleanup after 3 months

**FR6: Admin/Team Lead Features**
- Override employee participation
- View team headcount
- Manage day schedules (mark days as holidays, celebrations, office closed)
- View audit trail with change attribution

**FR7: Day Schedule Management**
- Mark days as: Normal, Office Closed, Government Holiday, Celebration
- Define available meals for special days
- System prevents opt-in/out on office closed days

**FR8: Real-Time Headcount**
- Admins/Logistics view current headcount per meal
- Breakdown: participating count, opted-out count, total employees
- Filter by date and meal type

### Non-Functional Requirements

**NFR1: Performance**
- Page load time < 2 seconds
- API response time < 500ms (95th percentile)
- Support 100 concurrent users

**NFR2: Availability**
- 99% uptime
- Graceful degradation on failures

**NFR3: Security**
- HTTPS only
- JWT token expiry: 2-4 hours
- Password: bcrypt with cost 12
- RBAC enforcement at API level

**NFR4: Usability**
- Mobile-responsive design
- Accessibility: WCAG 2.1 AA compliance
- Intuitive UI requiring minimal training

---
