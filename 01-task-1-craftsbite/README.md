# Task-1 of cea-2026-tasks

# CraftsBite

> Modern meal planning system replacing manual Excel-based tracking for 100+ employees

## Overview

CraftsBite is an internal web application that automates daily meal headcount tracking using an opt-out model. Employees are assumed to participate in meals unless they explicitly opt out, reducing administrative overhead and improving accuracy.

## Features

- 🔐 **Role-Based Access Control** - Employee, Team Lead, and Admin roles
- 📅 **Daily Meal Management** - Opt-out system with configurable cutoff times
- 📊 **Real-Time Headcount** - Live visibility for logistics team
- 🎯 **Admin Override** - Team leads can manage team member participation
- 🏢 **Day Scheduling** - Mark office closures, holidays, and celebrations
- 📱 **Responsive Design** - Works seamlessly on desktop and mobile browsers

## Tech Stack

### Frontend
- **React 18** with TypeScript
- **Vite** for fast development and optimized builds
- **Tailwind CSS & shadcn/ui** for styling
- **Zustand** for state management
- **React Hook Form** for form handling
- **Axios** for API communication

### Backend
- **Go 1.22** with Gin web framework
- **PostgreSQL 16** for data persistence
- **GORM** for ORM and migrations
- **JWT** for authentication
- **Zap** for structured logging

## Prerequisites

- **Node.js** 18+ and npm/yarn
- **Go** 1.22+
- **PostgreSQL** 16+x

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd craftsbite
```

### 2. Database Setup

**Local PostgreSQL**
```bash
createdb craftsbite_dev
```

### 3. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env with your database credentials

# Run migrations
make migrate-up

# Start the server
go run cmd/server/main.go
```

Backend will run on `http://localhost:8080`

### 4. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Copy environment template
cp .env.example .env

# Start development server
npm run dev
```

Frontend will run on `http://localhost:5173`

## Project Structure

```
craftsbite/
├── backend/
│   ├── cmd/
│   │   └── server/          # Application entrypoint
│   ├── internal/
│   │   ├── config/          # Configuration management
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # Auth, CORS, logging
│   │   ├── models/          # Database models
│   │   ├── repository/      # Data access layer
│   │   ├── routes/          # API route definitions
│   │   └── services/        # Business logic
│   ├── migrations/          # Database migrations
│   └── tests/              # Unit and integration tests
│
├── frontend/
│   ├── src/
│   │   ├── components/      # Reusable UI components
│   │   ├── pages/          # Page components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── services/       # API services
│   │   ├── store/          # Zustand stores
│   │   ├── types/          # TypeScript types
│   │   └── utils/          # Helper functions
│   └── public/             # Static assets
│
└── docs/                   # Documentation
```

## Development

### Running Tests

**Backend:**
```bash
cd backend
go test ./... -v
```

**Frontend:**
```bash
cd frontend
npm run test
```

### Database Migrations

**Create a new migration:**
```bash
cd backend
make migrate-create name=add_users_table
```

**Run migrations:**
```bash
make migrate-up
```

**Rollback migrations:**
```bash
make migrate-down
```

### Code Quality

**Backend:**
```bash
golangci-lint run
go fmt ./...
```

**Frontend:**
```bash
npm run lint
npm run format
```

## Environment Variables

### Backend (.env)

```env
# Server
PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=craftsbite_dev
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=http://localhost:5173
```

### Frontend (.env)

```env
VITE_API_BASE_URL=http://localhost:8080/api
```

## API Documentation

API endpoints are documented at `/api/docs` when running in development mode.

### Key Endpoints

- `POST /api/auth/login` - User authentication
- `GET /api/meals/today` - Today's meal status
- `POST /api/meals/opt-out` - Opt out of a meal
- `GET /api/admin/headcount` - Get real-time headcount
- `POST /api/admin/schedules` - Manage day schedules

## Default Users

After running migrations, test users are available:

| Email | Password | Role |
|-------|----------|------|
| admin@craftsmensoftware.com | admin123 | Admin |
| teamlead@craftsmensoftware.com | lead123 | Team Lead |
| employee@craftsmensoftware.com | emp123 | Employee |

**⚠️ Change these credentials in production!**

## Deployment

See [deployment documentation](./docs/DEPLOYMENT.md) for production setup instructions.

## Contributing

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Commit your changes: `git commit -am 'Add new feature'`
3. Push to the branch: `git push origin feature/your-feature`
4. Submit a pull request

### Commit Message Convention

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
pg_isready

# Check connection string
psql -h localhost -U postgres -d craftsbite_dev
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Frontend Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

## License

Internal use only - Proprietary

## Support

For issues or questions:
- Create an issue in the repository
- Contact: [Your Team Email]
- Documentation: `/docs` folder

---

**Version:** 1.0.0  
**Last Updated:** February 2026
