# CraftsBite Technical Document

**Project:** CraftsBite
**Version:** 3.0
**Status:** Iteration 1
**Last Updated:** February 9, 2026
**Author:** Sayad Ibn Khairul Alam

---

## 1. Summary

CraftsBite is an internal web application that modernizes the Excel-based meal planning process for 100+ employees. The system implements a flexible participation model where employees set default meal preferences (opt-in or opt-out) and manage exceptions as needed, with a comprehensive audit trail for all changes.

**Key Features:**
- User default meal preferences with flexible exception management
- Bulk opt-out for date ranges (vacations, business trips)
- Real-time headcount calculation for logistics
- Role-based access control (Employee, Team Lead, Admin, Logistics)
- Comprehensive participation history tracking
- Day schedule management (holidays, celebrations, office closures)

---


## 2. Problem Statement

**Current State:**
The meal planning process relies on manual Excel spreadsheet tracking, which creates significant operational challenges:
- High administrative burden on logistics team (daily manual updates)
- Frequent errors in headcount leading to meal wastage or shortages
- No self-service capability for employees to manage preferences
- Lack of audit trail for participation changes
- Inefficient communication for exceptions (vacations, business trips)
- Time-consuming reconciliation between multiple data sources

**Impact:**
- Wasted logistics team time that could be spent on strategic activities
- Food wastage due to inaccurate headcounts
- Employee frustration with rigid opt-out process
- No accountability for last-minute changes

---

## 3. Goals and Non-Goals

### Goals
1. **Automate meal tracking** - Replace Excel-based process with digital solution
2. **Empower employees** - Enable self-service preference management
3. **Improve accuracy** - Provide real-time, accurate headcount to logistics
4. **Reduce overhead** - Minimize administrative burden by 70%
5. **Maintain transparency** - Comprehensive audit trail for all changes
6. **Flexible participation** - Support default preferences and bulk exception management

### Non-Goals
1. **External vendor integration** - Not integrating with external catering systems (future iteration)
2. **Mobile apps** - Responsive web only, no native mobile apps
3. **Menu management** - Not managing actual meal menus or recipes
4. **Payment processing** - No meal payment or billing functionality
5. **Nutritional tracking** - Not tracking dietary restrictions or nutrition info
6. **Multi-organization support** - Single organization only

---

## 4. Tech Stack and Rationale

### Frontend

| Technology | Version | Rationale |
|-----------|---------|-----------|
| **React** | 19.x | Industry standard, team familiarity, large ecosystem |
| **TypeScript** | 5.x | Type safety, reduced runtime errors, better IDE support |
| **Vite** | 7.x | Fast dev server, optimized builds |
| **Tailwind CSS** | 4.x | Rapid development, utility-first, consistent design |
| **Zustand** | 5.x | Lightweight state management, simple API |
| **Axios** | 1.x | Interceptors, better error handling |
| **React Router** | 7.x | Standard routing solution, type-safe |
| **React Hook Form** | 7.x | Performant form handling, easy validation |
| **date-fns** | 3.x | Lightweight, modular, tree-shakeable |

### Backend

| Technology | Version | Rationale |
|-----------|---------|-----------|
| **Go** | 1.23.x | Performance, concurrency, simple deployment |
| **Gin** | 1.10.x | Fast, minimal, excellent routing |
| **PostgreSQL** | 17.x | ACID compliance, reliability, JSON support |
| **GORM** | 1.25.x | Popular Go ORM, migration support |
| **JWT** | 5.x | Stateless authentication, scalable |
| **bcrypt** | Latest | Secure password hashing |
| **Zap** | 1.x | Structured logging, high performance |

**Architecture Pattern:** Three-tier architecture with clean architecture principles (Handler → Service → Repository → Model layers)

---

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

### Flow 3: Employee Creates Bulk Opt-Out
1. Employee navigates to "Bulk Opt-Out" section
2. Selects date range (start date, end date)
3. Submits request
4. System creates record in `bulk_opt_outs` table
5. System applies opt-out to all meals in range
6. Confirmation with summary displayed

### Flow 4: Team Lead Overrides Participation
1. Team Lead views team member list
2. Selects member and date
3. Toggles participation status
4. System updates `meal_participations` with `overridden_by` field
5. Logs change to `meal_participation_history`
6. Notification shown

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

## 8. Design

### System Architecture

```
┌─────────────────────────────────────────┐
│         Client (Browser)                │
│    React + TypeScript SPA               │
│    State: Zustand (global)              │
└──────────────┬──────────────────────────┘
               │ HTTPS/REST API
┌──────────────▼──────────────────────────┐
│      Application Server (Go + Gin)      │
│  ┌────────────────────────────────────┐ │
│  │  Authentication Middleware (JWT)   │ │
│  │  Authorization Middleware (RBAC)   │ │
│  │  Handler → Service → Repository    │ │
│  └────────────────────────────────────┘ │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│      PostgreSQL Database (17.x)         │
│  - users (with default_meal_preference) │
│  - teams                                │
│  - team_members                         │
│  - meal_participations                  │
│  - meal_participation_history           │
│  - bulk_opt_outs                        │
│  - day_schedules                        │
└─────────────────────────────────────────┘
```

### Entity Relationship Diagram

![Entity Relationship Diagram](./technical_docs/images/entity_relation_diagram.png)


### Frontend Component Structure

```
src/
├── components/
│   ├── layout/          # Page structure: Header, Navbar, Footer, BottomActionButtons
│   ├── cards/           # Reusable card components: EmployeeMenuCard, InteractiveCard, StandardCard, AccentBorderCard
│   ├── forms/           # Form controls: Button, IconButton, Dropdown, Input fields
│   ├── modals/          # Modal dialogs: MealModal, MealOptOutModal, confirmation dialogs
│   └── feedback/        # User feedback: Toast notifications, Loading spinners, error messages
├── pages/
│   ├── LoginPage.tsx                 # Authentication page
│   ├── DashboardPage.tsx             # Main dashboard with overview
│   ├── EmployeeMealPage.tsx          # Employee meal calendar and preference management
│   ├── AdminHeadcountPage.tsx        # Admin/Logistics headcount view and reports
│   └── TeamLeadManagementPage.tsx    # Team lead override and team management
├── services/            # API service layer: HTTP client, API endpoint wrappers
├── store/               # Zustand state management: Auth store, meal store, global state
├── types/               # TypeScript type definitions: Interfaces, enums, API response types
├── hooks/               # Custom React hooks: useAuth, useMeals, useToast, API data fetching
└── utils/               # Helper functions: Date formatting, validators, constants, utility functions
```

**Key Directories:**
- **components/**: Reusable UI components organized by category
- **pages/**: Top-level route components (one per main route)
- **services/**: Centralized API communication layer
- **store/**: Global application state (authentication, user data)
- **types/**: Shared TypeScript interfaces and types
- **hooks/**: Custom React hooks for logic reuse
- **utils/**: Pure utility functions and constants


### Backend package Structure

```
cmd/
└── server/
    └── main.go                 # Application entry point

internal/
├── config/
│   └── config.go              # Configuration management
├── handlers/
│   ├── auth_handler.go        # Authentication endpoints
│   ├── meal_handler.go        # Meal participation endpoints
│   ├── user_handler.go        # User management endpoints
│   ├── headcount_handler.go   # Headcount reporting endpoints
│   ├── schedule_handler.go    # Day schedule management endpoints
│   ├── preference_handler.go  # User preference endpoints
│   ├── bulk_optout_handler.go # Bulk opt-out endpoints
│   └── history_handler.go     # Meal history endpoints
├── middleware/
│   ├── auth.go                # JWT authentication
│   ├── cors.go                # CORS configuration
│   ├── logger.go              # Request logging
│   ├── recovery.go            # Panic recovery
│   └── request_id.go          # Request ID tracking
├── models/
│   ├── user.go                # User entity
│   ├── meal.go                # Meal entity
│   ├── participation.go       # Participation entity
│   ├── schedule.go            # Day schedule entity
│   ├── role.go                # Role definitions
│   ├── history.go             # Participation history entity
│   ├── bulk_optout.go         # Bulk opt-out entity
│   └── team.go                # Team and TeamMember entities
├── repository/
│   ├── user_repository.go     # User data access
│   ├── meal_repository.go     # Meal data access
│   ├── schedule_repository.go # Schedule data access
│   ├── history_repository.go  # History data access
│   ├── bulk_optout_repository.go # Bulk opt-out data access
│   ├── team_repository.go     # Team data access
│   └── database.go            # Database connection & initialization
├── services/
│   ├── auth_service.go        # Authentication logic
│   ├── meal_service.go        # Meal business logic (ENHANCED)
│   ├── schedule_service.go    # Schedule business logic
│   ├── user_service.go        # User management logic
│   ├── headcount_service.go   # Headcount calculations (ENHANCED)
│   ├── preference_service.go  # User preference logic
│   ├── bulk_optout_service.go # Bulk opt-out logic
│   ├── history_service.go     # History tracking logic
│   └── participation_resolver.go # Participation status resolution
├── jobs/
│   └── cleanup_job.go         # History cleanup cron job
└── utils/
    ├── jwt.go                 # JWT utilities
    ├── password.go            # Password hashing
    ├── validator.go           # Custom validators
    └── response.go            # Standard response formats
```

---

## 9. Key Decisions and Trade-offs

### Decision 1: Default Preference Model
**Decision:** Implement user-level default preference (opt-in or opt-out) rather than requiring explicit opt-in/out for every meal.

**Rationale:**
- Reduces daily interaction burden for users
- Matches real-world behavior (most people have consistent patterns)
- Exceptions handled via specific records or bulk opt-outs

**Trade-offs:**
- More complex participation resolution logic
- Need to clearly communicate how defaults work
- ✅ Chosen for better UX

### Decision 2: Database Choice - PostgreSQL
**Decision:** PostgreSQL over MySQL or MongoDB.

**Rationale:**
- ACID compliance critical for meal counts
- Strong JSON support for flexible data (day schedules)
- Team familiarity
- Excellent performance for this scale

**Trade-offs:**
- More resource-intensive than MySQL
- ✅ Chosen for reliability and features

### Decision 3: JWT for Authentication
**Decision:** Stateless JWT tokens over session-based authentication.

**Rationale:**
- Scalability (no server-side session storage)
- Simpler deployment (no session store needed)
- Standard approach for API-first design

**Trade-offs:**
- Cannot revoke tokens before expiry (mitigated with 24-hour expiry)
- ✅ Chosen for simplicity and scalability

### Decision 4: Go + Gin for Backend
**Decision:** Go with Gin framework over Node.js/Express or Python/Django.

**Rationale:**
- Better performance and concurrency
- Single binary deployment (simpler operations)
- Strong typing and compile-time checks
- Team expanding Go expertise

**Trade-offs:**
- Smaller ecosystem than Node.js
- Less team familiarity initially
- ✅ Chosen for performance and operational simplicity

### Decision 5: Zustand over Redux
**Decision:** Zustand for state management instead of Redux or Context API.

**Rationale:**
- Much simpler API than Redux
- Better performance than Context API
- Smaller bundle size
- Sufficient for application complexity

**Trade-offs:**
- Less tooling/devtools than Redux
- ✅ Chosen for simplicity and team velocity

### Decision 6: Automatic 3-Month History Cleanup
**Decision:** Automatically delete participation history older than 3 months.

**Rationale:**
- Limits database growth
- Sufficient for accountability needs
- Privacy consideration (minimize data retention)

**Trade-offs:**
- Cannot retrieve older historical data
- ✅ Chosen to balance retention needs with data minimization

---


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

## 12. Operations

### Deployment Architecture

**Environment:** Single server deployment (initial iteration)
- **Frontend:** Static files served via Nginx
- **Backend:** Go binary running as systemd service
- **Database:** PostgreSQL 17.x
- **Reverse Proxy:** Nginx (SSL termination, static file serving)

### Deployment Process

1. **Build:**
   - Frontend: `npm run build` → static files
   - Backend: `go build` → single binary

2. **Deploy:**
   - Upload static files to `/var/www/craftsbite`
   - Upload Go binary to `/opt/craftsbite/`
   - Restart systemd service: `systemctl restart craftsbite`
   - Nginx reload: `systemctl reload nginx`

3. **Database Migrations:**
   - Run migrations before deploying new binary
   - Use `golang-migrate/migrate` tool
   - Rollback capability for failed migrations

### Monitoring

**Application Metrics:**
- Request latency (p50, p95, p99)
- Error rates (4xx, 5xx)
- Active users
- API endpoint usage

**Infrastructure Metrics:**
- CPU/Memory usage
- Disk space
- Database connection pool
- Network I/O

**Logging:**
- Application logs: Structured JSON (Zap logger)
- Access logs: Nginx format
- Log retention: 30 days
- Log aggregation: Centralized logging system (future)

### Backup and Recovery

**Database Backups:**
- Frequency: Daily at 2 AM
- Retention: 30 days
- Method: pg_dump
- Storage: Off-server backup location

**Recovery Time Objective (RTO):** 4 hours
**Recovery Point Objective (RPO):** 24 hours

### Maintenance

**Scheduled Jobs:**
- **Participation history cleanup:** Daily at 3 AM
  - Delete records older than 3 months
- **Database vacuum:** Weekly on Sunday at 1 AM
  - Reclaim storage, update statistics

**Dependency Updates:**
- Security patches: Apply within 48 hours
- Minor updates: Monthly review
- Major updates: Quarterly review with testing

---


## 13. Risks, Assumptions, and Open Questions

### Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Low user adoption** | Medium | High | Comprehensive training, intuitive UI, stakeholder engagement |
| **Performance issues at scale** | Low | Medium | Performance testing before launch, database indexing, query optimization |
| **Cutoff time confusion** | Medium | Medium | Clear UI indicators, email reminders, grace period for first week |
| **Security breach** | Low | High | Security best practices, regular audits, HTTPS only, input validation |
| **Database corruption** | Low | High | Daily backups, transaction support, database replication (future) |
| **Team lead override abuse** | Low | Medium | Audit logging, monthly access reviews, change notifications |

### Assumptions

1. **Network Connectivity:** Users have stable internet access during work hours
2. **Browser Support:** Users have modern browsers (Chrome 100+, Firefox 100+, Safari 15+, Edge 100+)
3. **User Base:** Maximum 100 concurrent users for iteration 1
4. **Data Volume:** Average 200 meal participation changes per day
5. **Cutoff Times:** 9 AM for lunch/snacks
6. **Working Days:** Monday-Friday (weekends assumed office closed unless specified)
7. **Single Organization:** No multi-tenancy required
9. **Role Assignment:** User roles managed manually by admin (no self-registration)
10. **Language:** English only (no i18n required for iteration 1)

### Open Questions

1. **Q:** Should we send email notifications for opt-out reminders?
   **Status:** Deferred to iteration 2
   **Decision needed by:** N/A

2. **Q:** Should we allow employees to see team-wide headcount?
   **Status:** Open - privacy concerns
   **Decision needed by:** Before launch

3. **Q:** How to handle conflicting bulk opt-outs (overlapping date ranges)?
   **Status:** Last-created wins (current implementation)
   **Decision needed by:** Resolved

4. **Q:** Should Team Leads be able to bulk-change their entire team's participation?
   **Status:** No for iteration 1 - individual overrides only
   **Decision needed by:** Resolved

5. **Q:** What happens to meal participation when user role changes?
   **Status:** Permissions change immediately, historical data preserved
   **Decision needed by:** Resolved

6. **Q:** Should we implement "favorite meals" or meal preferences?
   **Status:** Out of scope for iteration 1
   **Decision needed by:** N/A

7. **Q:** How granular should audit logs be (field-level vs record-level)?
   **Status:** Record-level (simpler, sufficient for accountability)
   **Decision needed by:** Resolved

8. **Q:** Should the system support recurring bulk opt-outs (e.g., every Friday)?
   **Status:** No for iteration 1 - manual creation required
   **Decision needed by:** N/A

---

## 14. Appendix

### A. Glossary

| Term | Definition |
|------|------------|
| **Opt-In** | Default assumption that employee will participate in meal |
| **Opt-Out** | Employee explicitly indicates they will not participate |
| **Default Meal Preference** | User's preferred default behavior (opt-in or opt-out) for all meals unless explicitly changed |
| **Participation History** | Audit trail of all explicit meal participation changes |
| **Cutoff Time** | Deadline for changing meal participation (9 AM lunch/snacks) |
| **Headcount** | Total number of employees participating in a meal |
| **Override** | Admin/Team Lead action to modify employee's participation |
| **Participation Resolution** | Process of determining user's meal status by checking day schedule → specific records → bulk opt-outs → default preference in priority order |
| **Auto-Cleanup** | Scheduled job that deletes participation history records older than 3 months |
| **Bulk Opt-Out** | Date range where user is automatically opted-out of all meals |
| **Day Schedule** | Admin-defined status for a specific day (normal, holiday, celebration, office closed) |

### B. Participation Resolution Algorithm

Priority order for determining if a user is participating in a meal:

1. **Check day_schedules:** If day is marked "office_closed", return not participating
2. **Check meal_participations:** If explicit record exists, use its value
3. **Check bulk_opt_outs:** If date falls in range, return not participating
4. **Check default_meal_preference:** Use user's default (opt-in or opt-out)

### C. Database Indexes

```sql
-- Performance-critical indexes
CREATE INDEX idx_meal_participations_user_date ON meal_participations(user_id, date);
CREATE INDEX idx_meal_participations_date_meal ON meal_participations(date, meal_type);
CREATE INDEX idx_bulk_opt_outs_user_dates ON bulk_opt_outs(user_id, start_date, end_date);
CREATE INDEX idx_history_user_created ON meal_participation_history(user_id, created_at);
CREATE INDEX idx_history_cleanup ON meal_participation_history(created_at);
CREATE INDEX idx_day_schedules_date ON day_schedules(date);
```

### D. Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=craftsbite
DB_USER=craftsbite_user
DB_PASSWORD=<secure_password>

# Application
APP_PORT=8080
APP_ENV=production
JWT_SECRET=<secure_random_string>
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=https://craftsbite.company.com

# Logging
LOG_LEVEL=info
```

### E. References

- [Go Official Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Security Guidelines](https://owasp.org/)

---

**Document Revision History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | Feb 6, 2026 | Sayad Ibn Khairul Alam | Initial draft |
| 2.0 | Feb 9, 2026 | Sayad Ibn Khairul Alam | Added teams/team_members tables, updated frontend architecture |
| 3.0 | Feb 9, 2026 | Sayad Ibn Khairul Alam | Restructured to follow standard tech doc format, condensed content |

---

*End of Document*


