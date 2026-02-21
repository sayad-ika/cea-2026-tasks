## 4. Tech Stack and Rationale

### Frontend

| Technology          | Version | Rationale                                               |
| ------------------- | ------- | ------------------------------------------------------- |
| **React**           | 19.x    | Industry standard, team familiarity, large ecosystem    |
| **TypeScript**      | 5.x     | Type safety, reduced runtime errors, better IDE support |
| **Vite**            | 7.x     | Fast dev server, optimized builds                       |
| **Tailwind CSS**    | 4.x     | Rapid development, utility-first, consistent design     |
| **Zustand**         | 5.x     | Lightweight state management, simple API                |
| **Axios**           | 1.x     | Interceptors, better error handling                     |
| **React Router**    | 7.x     | Standard routing solution, type-safe                    |
| **React Hook Form** | 7.x     | Performant form handling, easy validation               |
| **date-fns**        | 3.x     | Lightweight, modular, tree-shakeable                    |

### Backend

| Technology     | Version | Rationale                                   |
| -------------- | ------- | ------------------------------------------- |
| **Go**         | 1.23.x  | Performance, concurrency, simple deployment |
| **Gin**        | 1.10.x  | Fast, minimal, excellent routing            |
| **PostgreSQL** | 17.x    | ACID compliance, reliability, JSON support  |
| **GORM**       | 1.25.x  | Popular Go ORM, migration support           |
| **JWT**        | 5.x     | Stateless authentication, scalable          |
| **bcrypt**     | Latest  | Secure password hashing                     |
| **Zap**        | 1.x     | Structured logging, high performance        |

**Architecture Pattern:** Three-tier architecture with clean architecture principles (Handler → Service → Repository → Model layers)

---

## 8. Design

### System Architecture

```
┌─────────────────────────────────────────┐
│         Client (Browser)                │
│    React + TypeScript SPA               │
│    State: Zustand (global)              │
│    SSE Client (live headcount updates)  │
└──────────────┬──────────────────────────┘
               │ HTTPS/REST API + SSE
┌──────────────▼──────────────────────────┐
│      Application Server (Go + Gin)      │
│  ┌────────────────────────────────────┐ │
│  │  Authentication Middleware (JWT)   │ │
│  │  Authorization Middleware (RBAC)   │ │
│  │  Handler → Service → Repository    │ │
│  │  SSE Hub (live headcount updates)  │ │
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
│  - work_locations                       │
│  - wfh_periods                          │
│  - team_bulk_actions                    │
└─────────────────────────────────────────┘
```

### Entity Relationship Diagram

![Entity Relationship Diagram](./images/entity_relation_diagram.png)

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

---
