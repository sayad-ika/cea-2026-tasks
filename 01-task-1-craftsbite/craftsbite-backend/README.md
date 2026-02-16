# CraftsBite Backend

> A comprehensive meal management system backend for office cafeterias, built with Go and Gin framework.

## 📋 Project Overview

CraftsBite Backend is a production-ready server-side application for managing daily office meals, user participation, headcount reporting, and team management. The system supports role-based access control, flexible scheduling, bulk operations, and real-time participation tracking.

Built following clean architecture principles with distinct layers for handlers, services, repositories, and models, ensuring maintainability, testability, and scalability.

## ✨ Key Features

### 🔐 Authentication & Authorization
- JWT-based authentication with secure token management
- Role-based access control (RBAC)
- User roles: **Admin**, **Logistics**, **Team Lead**, **Employee**
- Session management with logout functionality

### 🍽️ Meal Management
- Daily meal participation tracking (Lunch & Snacks)
- Real-time opt-in/opt-out with cutoff time enforcement
- Participation override capabilities for Admin/Team-Leads

### 👥 Team Management
- Hierarchical team structure with Team Leads
- Team-based meal participation visibility
- Team member override panels for supervisors

### 📊 Headcount & Reporting
- Aggregated headcount reporting by date and meal type
- Meal-specific participation statistics
- Admin and Logistics dashboard support

## 🛠️ Technology Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.21+ |
| **Web Framework** | Gin Web Framework |
| **Database** | PostgreSQL |
| **ORM** | GORM v1.31.1 |
| **Configuration** | Viper |
| **Authentication** | JWT (golang-jwt/jwt) |
| **Logging** | Uber Zap |
| **Password Hashing** | bcrypt (golang.org/x/crypto) |
| **UUID Generation** | google/uuid |
| **Job Scheduling** | robfig/cron |

## 📁 Project Structure

```
craftsbite-backend/
├── cmd/
│   └── server/              # Application entry point (main.go)
├── internal/
│   ├── config/              # Configuration loading with Viper
│   ├── database/            # Database connection & pool management
│   ├── handlers/            # HTTP request handlers (8 handlers)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── meal_handler.go
│   │   ├── schedule_handler.go
│   │   ├── headcount_handler.go
│   │   ├── preference_handler.go
│   │   ├── bulk_optout_handler.go
│   │   └── history_handler.go
│   ├── middleware/          # Gin middleware (Auth, CORS, Logging, Recovery, RequestID)
│   ├── models/              # Database models and enums
│   │   ├── user.go
│   │   ├── meal_participation.go
│   │   ├── day_schedule.go
│   │   ├── team.go
│   │   ├── bulk_opt_out.go
│   │   └── meal_participation_history.go
│   ├── repository/          # Data access layer (6 repositories)
│   ├── services/            # Business logic layer (10 services)
│   ├── jobs/                # Background jobs (cleanup scheduler)
│   └── utils/               # Utility functions
├── pkg/
│   └── logger/              # Structured logging with Zap
├── migrations/              # Database migration SQL files (14 migrations)
├── .env.example             # Environment variables template
├── go.mod                   # Go module dependencies
└── README.md
```

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed on your system:

- **Go:** Version 1.21 or higher ([Download](https://golang.org/dl/))
- **PostgreSQL:** Version 12+ ([Download](https://www.postgresql.org/download/))
- **Git:** For version control

### Installation & Local Setup

#### 1. Clone the Repository

```bash
git clone <repository-url>
cd craftsbite-backend
```

#### 2. Setup PostgreSQL Database

Create a new PostgreSQL database for the project:

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database and user
CREATE DATABASE craftsbite_db;
CREATE USER craftsbite WITH ENCRYPTED PASSWORD 'craftsbite_secret';
GRANT ALL PRIVILEGES ON DATABASE craftsbite_db TO craftsbite;

# Exit psql
\q
```

#### 3. Configure Environment Variables

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Update `.env` with your configuration:

```env
# Environment
ENV=development

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=craftsbite
DB_PASSWORD=craftsbite_secret
DB_NAME=craftsbite_db
DB_SSLMODE=disable

# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRATION=24h

# CORS Configuration (update with your frontend URLs)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Meal Cutoff Configuration
MEAL_CUTOFF_TIME=21:00
MEAL_CUTOFF_TIMEZONE=Asia/Dhaka

# History Cleanup
HISTORY_RETENTION_MONTHS=3
CLEANUP_CRON=0 0 * * *
```

#### 4. Install Go Dependencies

```bash
go mod download
```

#### 5. Run Database Migrations

The migrations are located in the `migrations/` directory. Apply them manually in order:

```bash
# Connect to your database
psql -U craftsbite -d craftsbite_db

# Apply migrations in order
\i migrations/000001_create_users_table.up.sql
\i migrations/000002_create_meal_participations_table.up.sql
\i migrations/000003_create_day_schedules_table.up.sql
\i migrations/000004_add_default_meal_preference.up.sql
\i migrations/000005_create_meal_participation_history_table.up.sql
\i migrations/000006_create_bulk_opt_outs_table.up.sql
\i migrations/000007_add_team_lead_relationship.up.sql

# Verify tables were created
\dt
```

Alternatively, you can apply all migrations at once:

```bash
# On Windows
for /r migrations %i in (*.up.sql) do psql -U craftsbite -d craftsbite_db -f "%i"

# On Linux/macOS
for file in migrations/*.up.sql; do psql -U craftsbite -d craftsbite_db -f "$file"; done
```

#### 6. Run the Application

Start the development server:

```bash
go run cmd/server/main.go
```

You should see output similar to:

```
=================================
CraftsBite Backend Configuration
=================================
Environment: development
Server Address: localhost:8080
Database: craftsbite@localhost:5432/craftsbite_db
JWT Expiration: 24h
CORS Allowed Origins: [http://localhost:3000 http://localhost:5173]
Log Level: debug
=================================

Starting CraftsBite API server on localhost:8080
```

The server is now running at `http://localhost:8080` 🎉

#### 7. Verify Installation

Test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "healthy",
  "service": "craftsbite-api",
  "environment": "development"
}
```

### Building for Production

To build a production-ready binary:

```bash
go build -o craftsbite-server cmd/server/main.go
```

Run the binary:

```bash
# Linux/macOS
./craftsbite-server

# Windows
craftsbite-server.exe
```

## 🐳 Docker Support (Optional)

A `docker/` directory exists for containerization. Docker setup can be configured based on deployment requirements.

---

**Built with ❤️ using Go and Gin Framework**

