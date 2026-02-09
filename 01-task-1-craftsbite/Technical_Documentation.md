# CraftsBite
## Technical Architecture & Design Document

**Version:** 2.1
**Date:** February 6, 2026
**Last Updated:** February 9, 2026
**Author:** Sayad Ibn Khairul Alam
**Project Status:** Iteration 1 - Enhanced Development (with User Preference & History Features)

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [Architecture Design](#architecture-design)
4. [Technology Stack](#technology-stack)
5. [System Components](#system-components)
6. [Data Model](#data-model)
7. [API Design](#api-design)
8. [Security Architecture](#security-architecture)
9. [Frontend Architecture](#frontend-architecture)
10. [Deployment Strategy](#deployment-strategy)
11. [Development Guidelines](#development-guidelines)
12. [Testing Strategy](#testing-strategy)
13. [Future Iterations](#future-iterations)
14. [Risk Assessment](#risk-assessment)
15. [Timeline & Milestones](#timeline--milestones)

---

## 1. Executive Summary

### 1.1 Project Overview
CraftsBite is an internal web application designed to modernize the current Excel-based meal planning process for 100+ employees. The system implements a flexible participation model where employees can set their default meal preference (opt-in or opt-out) and manage exceptions as needed. The system maintains a comprehensive audit trail of all participation changes for accountability and transparency.

### 1.2 Business Objectives
- Replace manual Excel-based meal tracking with an automated digital solution
- Reduce administrative overhead for logistics team
- Improve accuracy of daily meal headcount
- Provide real-time visibility into meal participation
- Enable self-service for employees to manage their meal preferences
- Empower users with flexible default preferences and bulk management
- Maintain transparent audit trail for all participation changes

### 1.3 Key Success Metrics
- 100% employee onboarding within first 2 weeks
- Daily usage rate >95% for meal opt-out updates
- Reduction in meal wastage by accurate headcount prediction
- Administrative time savings of 70% for logistics team
- User satisfaction score >4.5/5 for preference management features

---

## 2. System Overview

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Layer (Browser)                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │         React + TypeScript SPA                        │  │
│  │  (State Management: Zustand for global state)         │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTPS/REST API
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                  Application Layer                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │         Go + Gin Framework (Backend API)              │  │
│  │                                                       │  │
│  │  ├─ Authentication Middleware (JWT)                   │  │
│  │  ├─ Authorization Middleware (RBAC)                   │  │
│  │  ├─ API Handlers                                      │  │
│  │  ├─ Business Logic Layer                              │  │
│  │  └─ Data Access Layer                                 │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Data Layer                                │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  PostgreSQL Database                                  │  │
│  │  - users table (with default_meal_preference)         │  │
│  │  - teams table                                        │  │
│  │  - team_members table                                 │  │
│  │  - meal_participations table                          │  │
│  │  - meal_participation_history table (audit trail)     │  │
│  │  - bulk_opt_outs table (date range management)        │  │
│  │  - day_schedules table                                │  │
│  │                                                       │  │
│  │  Accessed via GORM (Go ORM)                           │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 System Boundaries
- **In Scope (Iteration 1):**
  - User authentication and role-based access control
  - Daily meal opt-in/out management with flexible defaults
  - **User default meal preference (opt-in or opt-out)**
  - **Bulk opt-out for date ranges (vacations, business trips)**
  - **Comprehensive participation history tracking**
  - **Personal meal history view for users**
  - Real-time headcount calculation
  - Admin/Team Lead override capabilities
  - Day schedule management (office closed, government holidays, celebrations)
  - Audit trail with automatic 3-month cleanup
  - Change tracking for accountability and transparency

---

## 3. Architecture Design

### 3.1 Architectural Patterns

#### 3.1.1 Overall Architecture Style
- **Pattern:** Three-tier architecture with RESTful API
- **Rationale:** Clear separation of concerns, scalability, maintainability

#### 3.1.2 Backend Architecture
- **Pattern:** Layered architecture (Clean Architecture principles)
  - Handler Layer (HTTP handlers)
  - Service Layer (Business logic)
  - Repository Layer (Data access)
  - Model Layer (Domain entities)

```
┌─────────────────────────────────────────────────┐
│            Handler Layer (Gin)                  │
│  - HTTP request/response handling               │
│  - Input validation                             │
│  - Response formatting                          │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│            Service Layer                        │
│  - Business logic                               │
│  - Transaction management                       │
│  - Authorization checks                         │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│         Repository Layer                        │
│  - Data persistence                             │
│  - Database queries (GORM)                      │
│  - Data mapping                                 │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│            Model Layer                          │
│  - Domain entities                              │
│  - Business rules                               │
└─────────────────────────────────────────────────┘
```

#### 3.1.3 Frontend Architecture
- **Pattern:** Component-based architecture with React
- **State Management:** Zustand for global state (auth, user data), local state for component UI
- **Routing:** React Router v6 for client-side navigation

### 3.2 Design Principles
1. **Separation of Concerns:** Clear boundaries between layers
2. **DRY (Don't Repeat Yourself):** Reusable components and utilities
3. **SOLID Principles:** Applied in backend service design
4. **API-First Design:** Backend exposes well-defined REST API
5. **Security by Default:** Authentication and authorization at every layer

---

## 4. Technology Stack

### 4.1 Frontend Stack

| Component | Technology | Version | Justification |
|-----------|-----------|---------|---------------|
| Core Framework | React | 19.x | Industry standard, large ecosystem, team familiarity |
| Language | TypeScript | 5.x | Type safety, better IDE support, reduced runtime errors |
| Build Tool | Vite | 7.x | Fast development server, optimized production builds |
| UI Framework | Tailwind CSS | 4.x | Utility-first, rapid development, consistent design |
| State Management | Zustand | 5.x | Lightweight, simple API, better than Context API for global state |
| HTTP Client | Axios | 1.x | Interceptors, request/response transformation, better error handling |
| Routing | React Router | 7.x | Standard React routing solution, type-safe routes |
| Form Handling | React Hook Form | 7.x | Performance, easy validation, TypeScript support |
| Date Handling | date-fns | 3.x | Lightweight, modular, immutable, tree-shakeable |
| Error Boundaries | react-error-boundary | 4.x | Production error handling, graceful failure recovery |
| Toast Notifications | react-hot-toast | 2.x | Lightweight, accessible user feedback for actions |
| Icons | lucide-react | Latest | Consistent, customizable icon library |
| Environment Config | Vite Env Variables | Built-in | Type-safe build-time environment configuration |

**State Management Strategy:**
- **Zustand** for global state (authentication, user profile, preferences)
- **Local state** (useState) for component-specific UI state
- **React Hook Form** for form state management

### 4.2 Backend Stack

| Component | Technology | Version | Justification |
|-----------|-----------|---------|---------------|
| Language | Go | 1.23.x | Performance, concurrency, simple deployment (single binary) |
| Web Framework | Gin | 1.10.x | Fast, minimal, excellent routing, security patches |
| Database | PostgreSQL | 17.x | ACID compliance, reliability, JSON improvements, proven at scale |
| ORM | GORM | 1.25.x | Most popular Go ORM, migrations support, clean API |
| Database Driver | pgx | 5.x | High-performance PostgreSQL driver, better than lib/pq |
| Database Migrations | golang-migrate/migrate | 4.x | Version control for schema changes, rollback support |
| Authentication | golang-jwt/jwt | 5.x | Stateless, scalable, industry standard |
| Password Hashing | bcrypt (golang.org/x/crypto) | Latest | Secure password storage with configurable cost |
| Validation | go-playground/validator | 10.x | Declarative validation, struct tags, custom validators |
| Configuration | Viper | 1.x | Environment-based config, multiple formats support |
| Logging | Zap | 1.x | Structured logging, high performance, production-ready |
| CORS Middleware | gin-contrib/cors | 1.x | Cross-origin request handling for SPA architecture |
| Request ID Middleware | Custom | - | Request tracing and debugging across services |
| Recovery Middleware | gin.Recovery() | Built-in | Graceful panic recovery, prevents server crashes |

**Backend Architecture Notes:**
- Connection pool size: 25 (max), 5 (min idle) - tuned for 100 concurrent users
- JWT token expiry: 24 hours (configurable)
- Password bcrypt cost: 12 (balance between security and performance)

### 4.3 Testing Stack

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Backend Testing** |
| Unit Testing | Testify | 1.x | Assertions, mocking, suite support |
| HTTP Testing | httptest (stdlib) | Built-in | API endpoint testing without network calls |
| Database Testing | testcontainers-go | 0.28.x | Integration testing with real PostgreSQL instance |
| Coverage Tool | go test -cover | Built-in | Code coverage analysis |
| **Frontend Testing** |
| Unit Testing | Vitest | 1.x | Fast, Vite-native testing framework |
| Component Testing | React Testing Library | 14.x | User-centric component testing |
| E2E Testing | Playwright | 1.x | Browser automation for critical user flows |
| Coverage Tool | Vitest Coverage | Built-in | Istanbul-based coverage reporting |

**Testing Targets:**
- Backend: >80% code coverage
- Frontend: >70% code coverage
- E2E: Critical user paths (login, opt-out, admin override)

### 4.4 Development Tools

| Tool | Technology | Purpose |
|------|-----------|---------|
| Version Control | Git | Source code management |
| IDE | VS Code | Recommended IDE with Go and TypeScript extensions |
| API Testing | Postman / Insomnia | Manual API endpoint testing and documentation |
| Code Formatting (Frontend) | ESLint + Prettier | Consistent code style, auto-formatting |
| Code Linting (Backend) | golangci-lint | Static analysis, code quality enforcement |
| Live Reload (Backend) | Air | Hot reload for Go development |
| Containerization | Docker & Docker Compose | Local PostgreSQL, reproducible dev environment |
| Database Management | DBeaver / pgAdmin 4 | GUI for database inspection and queries |
| API Documentation | Swagger/OpenAPI | Auto-generated API documentation (optional) |
| Git Hooks | Husky (Frontend) | Pre-commit linting and formatting |

### 4.7 Database Configuration

| Setting | Value | Rationale |
|---------|-------|-----------|
| Max Connections | 100 | PostgreSQL default, suitable for 100+ users |
| Connection Pool (Max) | 25 | Prevents pool exhaustion under load |
| Connection Pool (Min Idle) | 5 | Keeps connections warm for fast response |
| Statement Timeout | 30s | Prevents long-running queries from blocking |
| Idle Transaction Timeout | 10min | Cleans up abandoned transactions |
| Work Memory | 4MB | Per-query memory for sorting/hashing |
| Shared Buffers | 256MB | Cache for frequently accessed data |
| WAL Level | replica | Enables point-in-time recovery |
| Archive Mode | ON | Required for backup strategy |

**Indexing Strategy:**
- Primary keys on all tables (auto-indexed)
- Foreign keys: `user_id`, `date` (composite index on meal_participations)
- Composite index: `(user_id, date)` for fast user meal lookups
- Index on `day_schedules.date` for holiday checks

**Backup Strategy:**
- Automated daily backups at 2 AM UTC via `pg_dump`
- 30-day retention policy
- WAL archiving for point-in-time recovery
- Weekly full backups, daily incrementals (Iteration 2)

---

## 5. System Components

### 5.1 Frontend Components

#### 5.1.1 Component Hierarchy

```
App
├── Layout
│   ├── Header
│   │   ├── Navigation
│   │   └── UserMenu
│   ├── Sidebar (optional)
│   └── Footer
├── Pages
│   ├── LoginPage
│   ├── DashboardPage
│   │   ├── TodayMealsCard
│   │   ├── DefaultPreferenceIndicator
│   ├── EmployeeMealPage
│   │   ├── MealList
│   │   ├── MealOptOutForm
│   ├── Settings/UserPreferencesPage
│   │   ├── DefaultPreferenceToggle
│   ├── MealHistoryPage
│   │   ├── HistoryCalendarView
│   │   ├── HistoryListView
│   ├── AdminHeadcountPage
│   │   ├── HeadcountSummary
│   │   └── DetailedBreakdown
│   ├── AdminSchedulePage
│   │   ├── ScheduleCalendar
│   │   ├── ScheduleForm
│   │   └── HolidayList
│   └── TeamLeadManagementPage
│       ├── EmployeeOverrideForm
│       └── TeamMealHistoryView
├── components/
│    ├── layout/
│    │   ├── Header.tsx
│    │   ├── Navbar.tsx
│    │   ├── Footer.tsx
│    │   ├── BottomActionButtons.tsx
│    │   └── index.ts
│    ├── cards/
│    │   ├── EmployeeMenuCard.tsx
│    │   ├── InteractiveCard.tsx
│    │   ├── StandardCard.tsx
│    │   ├── AccentBorderCard.tsx
│    │   └── index.ts
│    ├── forms/
│    │   ├── Button.tsx
│    │   ├── IconButton.tsx
│    │   ├── Dropdown.tsx
│    │   └── index.ts
│    ├── feedback/
│    │   ├── Toast.tsx
│    │   ├── Loading.tsx
│    │   └── index.ts
│    ├── modals/
│    │   ├── MealModal.tsx
│    │   ├── MealOptOutModal.tsx
│    │   └── index.ts
│    └── index.ts (barrel export)
└── Services
    ├── authService
    ├── mealService
    ├── scheduleService
    ├── userService
    ├── preferenceService
    ├── bulkOptOutService
    └── historyService
```

#### 5.1.2 Key Pages & Responsibilities

**LoginPage**
- User authentication
- Role-based redirection after login
- Remember me functionality

**DashboardPage (Employee)**
- Today's date and meal schedule
- Current participation status
- Day status indicator (normal, holiday, office closed)
- Quick opt-out actions
- Display user's default meal preference
- Show active bulk opt-outs affecting today

**EmployeeMealPage**
- Detailed view of all meals for the day
- Opt-in/opt-out toggles
- View cutoff times
- Create bulk opt-outs for date ranges
- Submission confirmation

**UserPreferencesPage**
- Set default meal preference (opt-in or opt-out)
- Manage bulk opt-outs (create, view, delete)
- View list of active bulk opt-outs with date ranges
- Explanation of how preferences work

**MealHistoryPage**
- Calendar view of past 30/60/90 days participation
- List view with filtering by date range and meal type
- Participation summary statistics
- Audit trail showing who changed what and when
- Visual indicators for opted-in, opted-out, overrides

**AdminHeadcountPage**
- Real-time headcount per meal type
- Export functionality (future)
- Date selection for historical view (future)

**AdminSchedulePage**
- Calendar view of scheduled days
- Create/edit day schedules (office closed, holidays, celebrations)
- Set available meals for special days
- Manage government holidays and company events

**TeamLeadManagementPage**
- View team members
- Override meal participation
- View team members' meal history
- Bulk update capabilities (future)

### 5.2 Backend Components

#### 5.2.1 Package Structure

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

migrations/                    # Database migration files
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_meal_participations_table.up.sql
├── 000002_create_meal_participations_table.down.sql
├── 000003_create_day_schedules_table.up.sql
├── 000003_create_day_schedules_table.down.sql
├── 000004_add_default_meal_preference.up.sql
├── 000004_add_default_meal_preference.down.sql
├── 000005_create_meal_participation_history_table.up.sql
├── 000005_create_meal_participation_history_table.down.sql
├── 000006_create_bulk_opt_outs_table.up.sql
├── 000006_create_bulk_opt_outs_table.down.sql
├── 000007_add_team_lead_relationship.up.sql
└── 000007_add_team_lead_relationship.down.sql

pkg/                           # Shared/public packages
└── logger/
    └── logger.go              # Logger configuration
```

---

## 6. Data Model

### 6.1 Entity Relationship Diagram

```
┌─────────────────────────────┐
│           User              │
├─────────────────────────────┤
│ id (string)                 │◄─────────┐
│ email (string)              │          │
│ name (string)               │          │
│ password (hash)             │          │
│ role (enum)                 │          │
│ active (bool)               │          │
│ default_meal_preference     │          │  
│   (enum: opt_in/opt_out)    │          │
│ created_at                  │          │
│ updated_at                  │          │
└─────────────────────────────┘          │
        │                                │ user_id (FK)
        │ user_id (FK)                   │
        ▼                                │
┌─────────────────────────────────────┐  │
│            Team                     │  │  
├─────────────────────────────────────┤  │
│ id (string)                         │  │
│ name (string)                       │  │
│ description (string)                │  │
│ team_lead_id (string)               │──┘
│ active (bool)                       │
│ created_at                          │
│ updated_at                          │
└─────────────────────────────────────┘
        │
        │ team_id (FK)
        ▼
┌─────────────────────────────────────┐
│        TeamMember                   │  
├─────────────────────────────────────┤
│ team_id (string) PK                 │
│ user_id (string) PK                 │
│ joined_at (datetime)                │
└─────────────────────────────────────┘
        │
        │ user_id (FK)
        ▼
┌─────────────────────────────────────┐
│      MealParticipation              │
├─────────────────────────────────────┤
│ id (string)                         │
│ user_id (string)                    │
│ date (string)                       │──┐
│ meal_type (enum)                    │  │
│ is_participating (bool)             │  │
│ opted_out_at (datetime)             │  │
│ override_by (string)                │  │
│ override_reason (string)            │  │
│ created_at                          │  │
│ updated_at                          │  │
└─────────────────────────────────────┘  │ date (FK)
        │                                │
        │                                │
        │ participation_id (FK)          │
        │                                │
┌─────────────────────────────────────┐  │
│  MealParticipationHistory           │  │  
├─────────────────────────────────────┤  │
│ id (string)                         │  │
│ user_id (string)                    │──┘
│ date (string)                       │
│ meal_type (enum)                    │
│ action (enum: opted_in/opted_out)   │
│ previous_value (enum)               │
│ changed_by_user_id (string)         │
│ reason (string)                     │
│ ip_address (string)                 │
│ created_at                          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│         BulkOptOut                  │  
├─────────────────────────────────────┤
│ id (string)                         │
│ user_id (string)                    │──┐
│ start_date (string)                 │  │
│ end_date (string)                   │  │
│ meal_type (enum)                    │  │
│ is_active (bool)                    │  │
│ created_at                          │  │
│ updated_at                          │  │
└─────────────────────────────────────┘  │
                                         │
                             ┌───────────┘
                             │
┌─────────────────────────────▼───────┐
│         DaySchedule                 │
├─────────────────────────────────────┤
│ id (string)                         │
│ date (string)                       │
│ day_status (enum)                   │
│ reason (string)                     │
│ available_meals (array)             │
│ created_by (string)                 │
│ created_at                          │
│ updated_at                          │
└─────────────────────────────────────┘

┌─────────────────────────┐
│   MealPreference        │  (Enum/Constant) 
├─────────────────────────┤
│ OPT_IN (default)        │
│ OPT_OUT                 │
└─────────────────────────┘

┌─────────────────────────┐
│   ParticipationAction   │  (Enum/Constant) 
├─────────────────────────┤
│ OPTED_IN                │
│ OPTED_OUT               │
│ OVERRIDE_IN             │
│ OVERRIDE_OUT            │
└─────────────────────────┘

┌─────────────────┐
│   MealType      │  (Enum/Constant)
├─────────────────┤
│ LUNCH           │
│ SNACKS          │
│ IFTAR           │
│ EVENT_DINNER    │
│ OPTIONAL_DINNER │
└─────────────────┘

┌─────────────────┐
│   DayStatus     │  (Enum/Constant)
├─────────────────┤
│ NORMAL          │
│ OFFICE_CLOSED   │
│ GOVT_HOLIDAY    │
│ CELEBRATION     │
└─────────────────┘

┌─────────────────┐
│      Role       │  (Enum/Constant)
├─────────────────┤
│ EMPLOYEE        │
│ TEAM_LEAD       │
│ ADMIN           │
│ LOGISTICS       │
└─────────────────┘
```

### 6.2 Data Models (Go Structs)

#### 6.2.1 User Model

```go
type Role string

const (
    RoleEmployee  Role = "employee"
    RoleTeamLead  Role = "team_lead"
    RoleAdmin     Role = "admin"
    RoleLogistics Role = "logistics"
)

// *** NEW: MealPreference Enum ***
type MealPreference string

const (
    MealPreferenceOptIn  MealPreference = "opt_in"  // Default: User participates unless they opt out
    MealPreferenceOptOut MealPreference = "opt_out" // User doesn't participate unless they opt in
)

type User struct {
    ID                    string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Email                 string         `gorm:"uniqueIndex;not null;size:255" json:"email" validate:"required,email"`
    Name                  string         `gorm:"not null;size:255" json:"name" validate:"required"`
    Password              string         `gorm:"not null" json:"-"` // Never expose in JSON
    Role                  Role           `gorm:"type:varchar(50);not null;default:'employee'" json:"role" validate:"required"`
    Active                bool           `gorm:"default:true" json:"active"`
    DefaultMealPreference MealPreference `gorm:"type:varchar(20);not null;default:'opt_in'" json:"default_meal_preference"` // 
    CreatedAt             time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt             time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
    return "users"
}
```

#### 6.2.2 Meal Participation Model

```go
type MealType string

const (
    MealTypeLunch          MealType = "lunch"
    MealTypeSnacks         MealType = "snacks"
    MealTypeIftar          MealType = "iftar"
    MealTypeEventDinner    MealType = "event_dinner"
    MealTypeOptionalDinner MealType = "optional_dinner"
)

type DayStatus string

const (
    DayStatusNormal         DayStatus = "normal"          // Regular working day
    DayStatusOfficeClosed   DayStatus = "office_closed"   // Office closed (no meals)
    DayStatusGovtHoliday    DayStatus = "govt_holiday"    // Government holiday (no meals)
    DayStatusCelebration    DayStatus = "celebration"     // Special celebration day
)

type MealParticipation struct {
    ID              string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    UserID          string     `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
    Date            string     `gorm:"type:date;not null;index" json:"date" validate:"required"` // Format: YYYY-MM-DD
    MealType        MealType   `gorm:"type:varchar(50);not null" json:"meal_type" validate:"required"`
    IsParticipating bool       `gorm:"default:true" json:"is_participating"`
    OptedOutAt      *time.Time `gorm:"type:timestamp" json:"opted_out_at,omitempty"`
    OverrideBy      string     `gorm:"type:uuid" json:"override_by,omitempty"` // User ID who made override
    OverrideReason  string     `gorm:"type:text" json:"override_reason,omitempty"`
    CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
    
    // Foreign key relationship
    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for GORM
func (MealParticipation) TableName() string {
    return "meal_participations"
}

// Composite unique index to prevent duplicate entries for same user/date/meal
func (MealParticipation) TableIndexes() []string {
    return []string{
        "idx_user_date_meal:user_id,date,meal_type,unique",
    }
}
```

#### 6.2.3 Day Schedule Model

```go
type DaySchedule struct {
    ID             string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Date           string     `gorm:"type:date;uniqueIndex;not null" json:"date" validate:"required"` // Format: YYYY-MM-DD
    DayStatus      DayStatus  `gorm:"type:varchar(50);not null;default:'normal'" json:"day_status" validate:"required"`
    Reason         string     `gorm:"type:text" json:"reason,omitempty"` // Reason for office closure or holiday
    AvailableMeals string     `gorm:"type:text" json:"-"` // Stored as JSON string in DB
    CreatedBy      string     `gorm:"type:uuid;not null" json:"created_by"` // Admin who set the schedule
    CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
    
    // Transient field for available meals (populated from AvailableMeals JSON string)
    AvailableMealsArray []MealType `gorm:"-" json:"available_meals"`
}

// TableName specifies the table name for GORM
func (DaySchedule) TableName() string {
    return "day_schedules"
}
```

#### 6.2.4 Meal Participation History Model

```go
// ParticipationAction represents the action taken on meal participation
type ParticipationAction string

const (
    ActionOptedIn     ParticipationAction = "opted_in"
    ActionOptedOut    ParticipationAction = "opted_out"
    ActionOverrideIn  ParticipationAction = "override_in"  // Admin/Team Lead overrode to opt-in
    ActionOverrideOut ParticipationAction = "override_out" // Admin/Team Lead overrode to opt-out
)

// MealParticipationHistory tracks all explicit changes to meal participation
// Default behavior (following user preference) is NOT tracked
// Records are automatically deleted after 3 months
type MealParticipationHistory struct {
    ID               string              `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    UserID           string              `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
    Date             string              `gorm:"type:date;not null;index" json:"date" validate:"required"` // Format: YYYY-MM-DD
    MealType         MealType            `gorm:"type:varchar(50);not null" json:"meal_type" validate:"required"`
    Action           ParticipationAction `gorm:"type:varchar(20);not null" json:"action" validate:"required"`
    PreviousValue    string              `gorm:"type:varchar(20)" json:"previous_value,omitempty"` // Previous participation status
    ChangedByUserID  string              `gorm:"type:uuid;not null" json:"changed_by_user_id"` // Who made the change (self or admin)
    Reason           string              `gorm:"type:varchar(255)" json:"reason,omitempty"` // Optional reason for change
    IPAddress        string              `gorm:"type:varchar(45)" json:"ip_address,omitempty"` // IPv4 or IPv6
    CreatedAt        time.Time           `gorm:"autoCreateTime;index" json:"created_at"`
    
    // Foreign key relationships
    User      User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
    ChangedBy User `gorm:"foreignKey:ChangedByUserID;constraint:OnDelete:SET NULL" json:"-"`
}

// TableName specifies the table name for GORM
func (MealParticipationHistory) TableName() string {
    return "meal_participation_history"
}

// Composite index for efficient querying
func (MealParticipationHistory) TableIndexes() []string {
    return []string{
        "idx_history_user_date:user_id,date",
        "idx_history_created_at:created_at", // For auto-cleanup queries
    }
}
```

#### 6.2.5 Bulk Opt-Out Model

```go
// BulkOptOut represents a date range where user has opted out
// Used for vacations, business trips, extended leave
type BulkOptOut struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    UserID    string    `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
    StartDate string    `gorm:"type:date;not null" json:"start_date" validate:"required"` // Format: YYYY-MM-DD
    EndDate   string    `gorm:"type:date;not null" json:"end_date" validate:"required,gtfield=StartDate"` // Must be >= StartDate
    MealType  MealType  `gorm:"type:varchar(50);not null" json:"meal_type" validate:"required"`
    IsActive  bool      `gorm:"default:true" json:"is_active"` // Can be deactivated without deletion
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    
    // Foreign key relationship
    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for GORM
func (BulkOptOut) TableName() string {
    return "bulk_opt_outs"
}

// Composite index for efficient date range queries
func (BulkOptOut) TableIndexes() []string {
    return []string{
        "idx_bulk_user_dates:user_id,start_date,end_date",
        "idx_bulk_active:is_active",
    }
}

// IsDateInRange checks if a given date falls within this bulk opt-out period
func (b *BulkOptOut) IsDateInRange(date string) bool {
    return b.IsActive && date >= b.StartDate && date <= b.EndDate
}
```

#### 6.2.6 Team Models

```go
// Team represents a team within the organization
type Team struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Name        string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
    Description string    `gorm:"type:text" json:"description,omitempty"`
    TeamLeadID  uuid.UUID `gorm:"type:uuid;not null;index" json:"team_lead_id" validate:"required"`
    Active      bool      `gorm:"not null;default:true" json:"active"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // Relationships
    TeamLead *User  `gorm:"foreignKey:TeamLeadID;constraint:OnDelete:CASCADE" json:"team_lead,omitempty"`
    Members  []User `gorm:"many2many:team_members" json:"members,omitempty"`
}

// TableName specifies the table name for GORM
func (Team) TableName() string {
    return "teams"
}

// TeamMember represents the junction table for team membership
type TeamMember struct {
    TeamID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
    UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
    JoinedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"joined_at"`

    // Relationships
    Team *Team `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"-"`
    User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for GORM
func (TeamMember) TableName() string {
    return "team_members"
}
```

#### 6.2.7 Headcount Summary Model

```go
type MealHeadcount struct {
    Date               string            `json:"date"`
    MealType           MealType          `json:"meal_type"`
    ParticipatingCount int               `json:"participating_count"`
    OptedOutCount      int               `json:"opted_out_count"`
    TotalEmployees     int               `json:"total_employees"`
    LastUpdated        time.Time         `json:"last_updated"`
}

type DailyHeadcountSummary struct {
    Date      string          `json:"date"`
    Meals     []MealHeadcount `json:"meals"`
    GeneratedAt time.Time     `json:"generated_at"`
}
```

### 6.3 Database Schema

The application uses PostgreSQL as the relational database. GORM handles automatic migrations and schema creation.

#### 6.3.1 users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'employee',
    active BOOLEAN DEFAULT true,
    default_meal_preference VARCHAR(20) NOT NULL DEFAULT 'opt_in', -- 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(active);
```

#### 6.3.2 meal_participations Table

```sql
CREATE TABLE meal_participations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL,
    is_participating BOOLEAN DEFAULT true,
    opted_out_at TIMESTAMP,
    override_by UUID REFERENCES users(id),
    override_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_user_date_meal UNIQUE (user_id, date, meal_type)
);

CREATE INDEX idx_meal_participations_user_id ON meal_participations(user_id);
CREATE INDEX idx_meal_participations_date ON meal_participations(date);
CREATE INDEX idx_meal_participations_user_date ON meal_participations(user_id, date);
```

#### 6.3.3 day_schedules Table

```sql
CREATE TABLE day_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE UNIQUE NOT NULL,
    day_status VARCHAR(50) NOT NULL DEFAULT 'normal',
    reason TEXT,
    available_meals TEXT, -- JSON array stored as text
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_day_schedules_date ON day_schedules(date);
CREATE INDEX idx_day_schedules_status ON day_schedules(day_status);
```

#### 6.3.4 meal_participation_history Table

```sql
CREATE TABLE meal_participation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL, -- 'opted_in', 'opted_out', 'override_in', 'override_out'
    previous_value VARCHAR(20), -- Previous participation status
    changed_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(255), -- Optional reason for change
    ip_address VARCHAR(45), -- IPv4 or IPv6 address
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_history_user_id ON meal_participation_history(user_id);
CREATE INDEX idx_history_user_date ON meal_participation_history(user_id, date);
CREATE INDEX idx_history_created_at ON meal_participation_history(created_at); -- For auto-cleanup

-- Auto-cleanup function to delete records older than 3 months
CREATE OR REPLACE FUNCTION cleanup_old_participation_history()
RETURNS void AS $$
BEGIN
    DELETE FROM meal_participation_history
    WHERE created_at < NOW() - INTERVAL '3 months';
END;
$$ LANGUAGE plpgsql;

-- Optional: Create a scheduled job using pg_cron extension (if available)
-- SELECT cron.schedule('cleanup-history', '0 2 * * *', 'SELECT cleanup_old_participation_history()');
```

#### 6.3.5 bulk_opt_outs Table

```sql
CREATE TABLE bulk_opt_outs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Validation constraint: end_date must be >= start_date
    CONSTRAINT valid_date_range CHECK (end_date >= start_date)
);

CREATE INDEX idx_bulk_user_id ON bulk_opt_outs(user_id);
CREATE INDEX idx_bulk_user_dates ON bulk_opt_outs(user_id, start_date, end_date);
CREATE INDEX idx_bulk_active ON bulk_opt_outs(is_active);
CREATE INDEX idx_bulk_date_range ON bulk_opt_outs(start_date, end_date); -- For range queries
```

#### 6.3.6 teams Table

```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    team_lead_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_teams_team_lead_id ON teams(team_lead_id);
CREATE INDEX idx_teams_active ON teams(active);

-- Add comments for documentation
COMMENT ON TABLE teams IS 'Teams within the organization, each led by a team lead';
COMMENT ON COLUMN teams.team_lead_id IS 'User ID of the team lead who manages this team';
COMMENT ON COLUMN teams.active IS 'Whether the team is currently active';
```

#### 6.3.7 team_members Table

```sql
-- Create team_members junction table (many-to-many relationship)
CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id, user_id)
);

-- Create indexes for efficient lookups
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- Add comments for documentation
COMMENT ON TABLE team_members IS 'Junction table for many-to-many relationship between teams and users';
```

#### 6.3.8 Sample Data

**Users:**
```sql
INSERT INTO users (id, email, name, password, role) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'admin@company.com', 'Admin User', '$2a$10$...', 'admin'),
('550e8400-e29b-41d4-a716-446655440001', 'john.doe@company.com', 'John Doe', '$2a$10$...', 'employee');
```

**Day Schedules:**
```sql
INSERT INTO day_schedules (date, day_status, available_meals, created_by) VALUES
('2026-02-07', 'normal', '["lunch","snacks"]', '550e8400-e29b-41d4-a716-446655440000'),
('2026-02-08', 'govt_holiday', '[]', '550e8400-e29b-41d4-a716-446655440000'),
('2026-02-14', 'celebration', '["lunch","snacks","event_dinner"]', '550e8400-e29b-41d4-a716-446655440000');
```

---

## 7. API Design

### 7.1 API Principles
- RESTful design
- JSON request/response format
- Standard HTTP status codes
- Consistent error response structure
- JWT bearer token authentication

### 7.2 Standard Response Format

#### Success Response
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### 7.3 API Endpoints

#### 7.3.1 Authentication Endpoints

**POST /api/v1/auth/login**
- Description: Authenticate user and return JWT token
- Access: Public
- Request:
```json
{
  "email": "user@company.com",
  "password": "password123"
}
```
- Response:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid-1",
      "email": "user@company.com",
      "name": "John Doe",
      "role": "employee"
    }
  }
}
```

**POST /api/v1/auth/logout**
- Description: Invalidate current session (optional for JWT)
- Access: Authenticated
- Response: 200 OK

**GET /api/v1/auth/me**
- Description: Get current user information
- Access: Authenticated
- Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-1",
    "email": "user@company.com",
    "name": "John Doe",
    "role": "employee"
  }
}
```

#### 7.3.2 Meal Participation Endpoints

**GET /api/v1/meals/today**
- Description: Get today's meal schedule with user's participation status
- Access: Authenticated (All roles)
- Response:
```json
{
  "success": true,
  "data": {
    "date": "2026-02-06",
    "meals": [
      {
        "meal_type": "lunch",
        "is_participating": true,
        "can_modify": true,
        "cutoff_time": "10:30:00"
      },
      {
        "meal_type": "snacks",
        "is_participating": true,
        "can_modify": true,
        "cutoff_time": "14:00:00"
      }
    ]
  }
}
```

**GET /api/v1/meals/participation**
- Description: Get meal participation for a specific date
- Access: Authenticated (All roles)
- Query Parameters: `date` (optional, defaults to today)
- Response: Similar to /meals/today

**POST /api/v1/meals/participation**
- Description: Create or update meal participation
- Access: Authenticated (All roles)
- Request:
```json
{
  "date": "2026-02-06",
  "meal_type": "lunch",
  "is_participating": false
}
```
- Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-p1",
    "user_id": "uuid-1",
    "date": "2026-02-06",
    "meal_type": "lunch",
    "is_participating": false,
    "opted_out_at": "2026-02-06T09:30:00Z"
  },
  "message": "Meal participation updated successfully"
}
```

**PUT /api/v1/meals/participation/:id**
- Description: Update existing participation record
- Access: Authenticated (All roles for own records)
- Request: Same as POST
- Response: Updated participation object

#### 7.3.3 Admin/Team Lead Override Endpoints

**POST /api/v1/admin/meals/participation/override**
- Description: Override meal participation for an employee
- Access: Team Lead, Admin, Logistics
- Request:
```json
{
  "user_id": "uuid-2",
  "date": "2026-02-06",
  "meal_type": "lunch",
  "is_participating": true,
  "reason": "Employee forgot to opt-in before cutoff"
}
```
- Response: Participation object with override information

#### 7.3.4 Headcount Reporting Endpoints

**GET /api/v1/headcount/today**
- Description: Get headcount summary for today
- Access: Admin, Logistics
- Response:
```json
{
  "success": true,
  "data": {
    "date": "2026-02-06",
    "meals": [
      {
        "meal_type": "lunch",
        "participating_count": 95,
        "opted_out_count": 5,
        "total_employees": 100
      },
      {
        "meal_type": "snacks",
        "participating_count": 98,
        "opted_out_count": 2,
        "total_employees": 100
      }
    ],
    "generated_at": "2026-02-06T08:00:00Z"
  }
}
```

**GET /api/v1/headcount/date/:date**
- Description: Get headcount for specific date
- Access: Admin, Logistics
- Parameters: `date` (YYYY-MM-DD)
- Response: Similar to /headcount/today

**GET /api/v1/headcount/detailed**
- Description: Get detailed breakdown with employee names
- Access: Admin, Logistics
- Query Parameters: `date`, `meal_type`
- Response:
```json
{
  "success": true,
  "data": {
    "date": "2026-02-06",
    "meal_type": "lunch",
    "participating": [
      {
        "id": "uuid-1",
        "name": "John Doe",
        "email": "john@company.com"
      }
    ],
    "opted_out": [
      {
        "id": "uuid-2",
        "name": "Jane Smith",
        "email": "jane@company.com",
        "opted_out_at": "2026-02-06T09:00:00Z"
      }
    ]
  }
}
```

#### 7.3.5 Day Schedule Management Endpoints

**GET /api/v1/schedule/date/:date**
- Description: Get day schedule for a specific date
- Access: All authenticated users
- Parameters: `date` (YYYY-MM-DD)
- Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-s1",
    "date": "2026-02-07",
    "day_status": "normal",
    "available_meals": ["lunch", "snacks"],
    "reason": null
  }
}
```

**GET /api/v1/schedule/range**
- Description: Get day schedules for a date range
- Access: Admin, Logistics
- Query Parameters: `start_date`, `end_date`
- Response: Array of day schedules

**POST /api/v1/schedule**
- Description: Create day schedule (set office closed, holiday, etc.)
- Access: Admin
- Request:
```json
{
  "date": "2026-03-21",
  "day_status": "govt_holiday",
  "reason": "Independence Day",
  "available_meals": []
}
```
- Response: Created schedule object

**PUT /api/v1/schedule/:id**
- Description: Update existing day schedule
- Access: Admin
- Request: Partial schedule object
- Response: Updated schedule object

**DELETE /api/v1/schedule/:id**
- Description: Delete day schedule (revert to normal)
- Access: Admin
- Response: 200 OK

#### 7.3.6 User Management Endpoints (Admin only)

**GET /api/v1/users**
- Description: List all users
- Access: Admin
- Query Parameters: `role`, `active`, `page`, `limit`
- Response: Paginated user list

**GET /api/v1/users/:id**
- Description: Get user details
- Access: Admin, or self
- Response: User object

**POST /api/v1/users**
- Description: Create new user
- Access: Admin
- Request:
```json
{
  "email": "newuser@company.com",
  "name": "New User",
  "password": "tempPassword123",
  "role": "employee"
}
```

**PUT /api/v1/users/:id**
- Description: Update user
- Access: Admin, or self (limited fields)
- Request: Partial user object

**DELETE /api/v1/users/:id**
- Description: Deactivate user (soft delete)
- Access: Admin

#### 7.3.7 User Preference Endpoints

**GET /api/v1/users/me/preferences**
- Description: Get current user's meal preferences
- Access: Authenticated
- Response:
```json
{
  "success": true,
  "data": {
    "user_id": "uuid-1",
    "default_meal_preference": "opt_in",
    "updated_at": "2026-02-06T10:00:00Z"
  }
}
```

**PUT /api/v1/users/me/preferences**
- Description: Update user's default meal preference
- Access: Authenticated
- Request:
```json
{
  "default_meal_preference": "opt_out"
}
```
- Response:
```json
{
  "success": true,
  "data": {
    "user_id": "uuid-1",
    "default_meal_preference": "opt_out",
    "updated_at": "2026-02-06T10:30:00Z"
  },
  "message": "Meal preference updated successfully"
}
```

#### 7.3.8 Bulk Opt-Out Endpoints

**GET /api/v1/meals/bulk-optouts**
- Description: Get all active bulk opt-outs for current user
- Access: Authenticated
- Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-b1",
      "user_id": "uuid-1",
      "start_date": "2026-03-10",
      "end_date": "2026-03-14",
      "meal_type": "lunch",
      "is_active": true,
      "created_at": "2026-02-06T10:00:00Z"
    }
  ]
}
```

**POST /api/v1/meals/bulk-optouts**
- Description: Create bulk opt-out for date range (vacation, business trip)
- Access: Authenticated
- Request:
```json
{
  "start_date": "2026-03-10",
  "end_date": "2026-03-14",
  "meal_type": "lunch"
}
```
- Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-b1",
    "user_id": "uuid-1",
    "start_date": "2026-03-10",
    "end_date": "2026-03-14",
    "meal_type": "lunch",
    "is_active": true,
    "created_at": "2026-02-06T10:00:00Z"
  },
  "message": "Bulk opt-out created successfully"
}
```

**DELETE /api/v1/meals/bulk-optouts/:id**
- Description: Delete/deactivate a bulk opt-out
- Access: Authenticated (own records only)
- Response: 200 OK

#### 7.3.9 Meal History Endpoints

**GET /api/v1/meals/history**
- Description: Get meal participation history for current user
- Access: Authenticated
- Query Parameters: `start_date`, `end_date`, `meal_type`, `limit` (default: 30 days)
- Response:
```json
{
  "success": true,
  "data": {
    "user_id": "uuid-1",
    "start_date": "2026-01-07",
    "end_date": "2026-02-06",
    "history": [
      {
        "date": "2026-02-06",
        "meal_type": "lunch",
        "status": "opted_out",
        "changed_at": "2026-02-06T09:30:00Z",
        "changed_by": "self"
      },
      {
        "date": "2026-02-05",
        "meal_type": "lunch",
        "status": "participated",
        "is_default": true
      },
      {
        "date": "2026-02-04",
        "meal_type": "lunch",
        "status": "override_in",
        "changed_at": "2026-02-04T09:45:00Z",
        "changed_by": "team_lead",
        "changed_by_name": "Sarah Johnson",
        "reason": "Team meeting with lunch"
      }
    ],
    "summary": {
      "total_days": 30,
      "participated": 25,
      "opted_out": 5,
      "participation_rate": "83.3%"
    }
  }
}
```

**GET /api/v1/admin/meals/history/:user_id**
- Description: Get meal participation history for any user (admin view)
- Access: Admin, Team Lead (for their team)
- Query Parameters: `start_date`, `end_date`, `meal_type`
- Response: Similar to /meals/history with additional admin details

**GET /api/v1/meals/participation-audit**
- Description: Get detailed audit trail of participation changes
- Access: Authenticated (own records), Admin (all records)
- Query Parameters: `user_id`, `date`, `meal_type`
- Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-h1",
      "user_id": "uuid-1",
      "date": "2026-02-06",
      "meal_type": "lunch",
      "action": "opted_out",
      "previous_value": "opted_in",
      "changed_by_user_id": "uuid-1",
      "changed_by_name": "John Doe",
      "reason": null,
      "ip_address": "192.168.1.100",
      "created_at": "2026-02-06T09:30:00Z"
    }
  ]
}
```

### 7.4 HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Successful GET/PUT/DELETE request |
| 201 | Successful POST request (resource created) |
| 400 | Bad request (validation error) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Resource not found |
| 409 | Conflict (duplicate resource) |
| 500 | Internal server error |

--- 

## 8. Security Architecture

### 8.1 Authentication

#### JWT Token Structure
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user-uuid",
    "email": "user@company.com",
    "role": "employee",
    "exp": 1675785600,
    "iat": 1675699200
  }
}
```

#### Token Configuration
- Algorithm: HS256
- Expiration: 24 hours
- Refresh: Requires re-login (future: refresh tokens)
- Storage: localStorage (client-side)

### 8.2 Authorization (RBAC)

#### Role Permissions Matrix

| Action | Employee | Team Lead | Admin | Logistics |
|--------|----------|-----------|-------|-----------|
| View own meals | ✓ | ✓ | ✓ | ✓ |
| Update own participation | ✓ | ✓ | ✓ | ✓ |
| View day schedule | ✓ | ✓ | ✓ | ✓ |
| Override team member participation | ✗ | ✓ | ✓ | ✓ |
| View headcount summary | ✗ | ✗ | ✓ | ✓ |
| View detailed headcount | ✗ | ✗ | ✓ | ✓ |
| Manage day schedules (holidays, closures) | ✗ | ✗ | ✓ | ✗ |
| Manage users | ✗ | ✗ | ✓ | ✗ |
| System configuration | ✗ | ✗ | ✓ | ✗ |

### 8.3 Security Measures

#### Backend Security
1. **Password Hashing:** bcrypt with cost factor 10
2. **SQL Injection Prevention:** GORM ORM with parameterized queries, prepared statements
3. **XSS Prevention:** Input sanitization, output encoding
4. **CSRF Protection:** SameSite cookies, CORS configuration
5. **Rate Limiting:** 100 requests/minute per IP (future)
6. **Input Validation:** Server-side validation on all endpoints
7. **Database Security:**
   - Connection pooling with max connections limit
   - Prepared statements for all queries
   - Foreign key constraints for referential integrity
   - Row-level permissions (future enhancement)
8. **Secure Headers:** 
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - Content-Security-Policy

#### Frontend Security
1. **Token Storage:** localStorage (consider httpOnly cookies in future)
2. **Auto-logout:** On token expiration
3. **Sensitive Data:** Never log passwords or tokens
4. **HTTPS Only:** Force HTTPS in production
5. **Input Sanitization:** Client-side validation

### 8.4 Data Privacy
- Personal data limited to: name, email, meal preferences
- No sensitive personal information collected
- Audit logs for admin actions
- Data retention policy: 90 days for meal participation history

---

## 9. Frontend Architecture

### 9.1 Folder Structure

```
src/
├── assets/               # Static assets (images, fonts)
├── components/           # Reusable components
│   ├── layout/
│   │   ├── Header.tsx
│   │   ├── Navbar.tsx
│   │   ├── Footer.tsx
│   │   ├── BottomActionButtons.tsx
│   │   └── index.ts
│   ├── cards/
│   │   ├── EmployeeMenuCard.tsx
│   │   ├── InteractiveCard.tsx
│   │   ├── StandardCard.tsx
│   │   ├── AccentBorderCard.tsx
│   │   └── index.ts
│   ├── forms/
│   │   ├── Button.tsx
│   │   ├── IconButton.tsx
│   │   ├── Dropdown.tsx
│   │   └── index.ts
│   ├── feedback/
│   │   ├── Toast.tsx
│   │   ├── Loading.tsx
│   │   └── index.ts
│   ├── modals/
│   │   ├── MealModal.tsx
│   │   ├── MealOptOutModal.tsx
│   │   └── index.ts
│   └── index.ts             # Barrel export
├── pages/
│   ├── LoginPage.tsx
│   ├── DashboardPage.tsx
│   ├── EmployeeMealPage.tsx
│   ├── AdminHeadcountPage.tsx
│   └── TeamLeadManagementPage.tsx
├── services/
│   ├── api.ts            # Axios instance
│   ├── authService.ts
│   ├── mealService.ts
│   ├── scheduleService.ts
│   └── userService.ts
├── store/                # Zustand stores
│   ├── authStore.ts
│   └── mealStore.ts
├── types/                # TypeScript types
│   ├── auth.types.ts
│   ├── meal.types.ts
│   ├── schedule.types.ts
│   └── user.types.ts
├── utils/
│   ├── dateUtils.ts
│   ├── validators.ts
│   └── constants.ts
├── hooks/                # Custom React hooks
│   ├── useAuth.ts
│   ├── useMeals.ts
│   └── useToast.ts
├── routes/
│   ├── ProtectedRoute.tsx
│   └── routes.tsx
├── App.tsx
├── main.tsx
└── vite-env.d.ts
```

### 9.2 State Management Strategy

#### Global State (Zustand)
- Authentication state (user, token, isAuthenticated)
- Current meal data
- App-wide settings

#### Local Component State
- Form inputs
- UI state (modals, dropdowns)
- Temporary data

### 9.3 Routing Structure

```typescript
const routes = [
  {
    path: '/login',
    element: <LoginPage />,
    public: true
  },
  {
    path: '/',
    element: <ProtectedRoute />,
    children: [
      {
        path: '/',
        element: <DashboardPage />
      },
      {
        path: '/meals',
        element: <EmployeeMealPage />
      },
      {
        path: '/admin/headcount',
        element: <AdminHeadcountPage />,
        roles: ['admin', 'logistics']
      },
      {
        path: '/team-lead/manage',
        element: <TeamLeadManagementPage />,
        roles: ['team_lead', 'admin']
      }
    ]
  }
];
```

### 9.4 TypeScript Types

```typescript
// auth.types.ts
export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
}

export type UserRole = 'employee' | 'team_lead' | 'admin' | 'logistics';

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

// meal.types.ts
export type MealType = 
  | 'lunch' 
  | 'snacks' 
  | 'iftar' 
  | 'event_dinner' 
  | 'optional_dinner';

export type DayStatus = 
  | 'normal' 
  | 'office_closed' 
  | 'govt_holiday' 
  | 'celebration';

export interface MealParticipation {
  id: string;
  userId: string;
  date: string;
  mealType: MealType;
  isParticipating: boolean;
  optedOutAt?: string;
  canModify: boolean;
}

export interface MealHeadcount {
  mealType: MealType;
  participatingCount: number;
  optedOutCount: number;
  totalEmployees: number;
}

export interface DaySchedule {
  id: string;
  date: string;
  dayStatus: DayStatus;
  reason?: string;
  availableMeals: MealType[];
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}
```

---

## Appendices

### Appendix A: Glossary

- **Opt-In:** Default assumption that employee will participate in meal
- **Opt-Out:** Employee explicitly indicates they will not participate
- **Default Meal Preference:** User's preferred default behavior (opt-in or opt-out) for all meals unless explicitly changed
- **Participation History:** Audit trail of all explicit meal participation changes
- **Cutoff Time:** Deadline for changing meal participation
- **Headcount:** Total number of employees participating in a meal
- **Override:** Admin/Team Lead action to modify employee's participation
- **Participation Resolution:** Process of determining user's meal status by checking day schedule, specific records, bulk opt-outs, and default preference in priority order
- **Auto-Cleanup:** Scheduled job that deletes participation history records older than 3 months

### Appendix B: References

- [Go Official Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

**Document Version History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-06 | Sayad IKA | Initial draft |
| 2.0 | 2026-02-09 | Sayad IKA | Added teams and team_members tables for team management; Updated frontend components architecture with new organized structure (layout, cards, forms, feedback, modals) |

---

*End of Document*
