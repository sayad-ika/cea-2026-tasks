# CraftsBite Frontend 🍽️

A modern, enterprise-grade **meal management system** built with React and TypeScript. CraftsBite streamlines daily meal planning, participation tracking, and headcount reporting for organizations with an elegant **Claymorphism** design aesthetic.

---

## 🎯 Overview

CraftsBite Frontend is the client-side application for managing organizational meal services. It provides role-based interfaces for employees to manage their meal preferences, team leads to oversee their teams, and administrators to track headcount and make operational decisions.

**Key Capabilities:**
- 🍽️ **Real-time Meal Participation** - Employees can opt-in/opt-out of daily meals
- 📊 **Headcount Analytics** - Comprehensive reporting for logistics and planning
- 👥 **Team Management** - Role-based access control for team leads and admins
- ⚙️ **Manual Overrides** - Administrative controls with audit logging
- 🌙 **Dark Mode** - Full theme support for day and night usage

---

## 🚀 Technology Stack

| Category | Technology | Purpose |
|----------|-----------|---------|
| **Framework** | React 19 | Modern component-based UI library |
| **Language** | TypeScript | Type-safe development |
| **Build Tool** | Vite | Lightning-fast development and builds |
| **Styling** | Tailwind CSS 4 | Utility-first CSS framework |
| **State** | Zustand + Context API | Global state management |
| **Routing** | React Router 7 | Client-side routing |
| **HTTP** | Axios | API communication |
| **Forms** | React Hook Form | Form state management |
| **Icons** | Material Symbols + Lucide | Icon sets |
| **Notifications** | React Hot Toast | Toast notifications |
| **Dates** | date-fns | Date manipulation |
| **Error Handling** | React Error Boundary | Graceful error recovery |

---

## ✨ Core Features

### 👤 Employee Dashboard (`/home`)
The primary interface for all employees to manage their meal participation.

- **📅 Next-Day Menu View** - Displays tomorrow's available meals (Lunch & Snacks)
- **✅ One-Click Toggle** - Simple opt-in/opt-out mechanism for each meal
- **⏰ Cutoff Time Enforcement** - Automatic locking after configured deadline (e.g., 9 PM)
- **📍 Status Indicators** - Clear visual feedback for current participation state
- **🔒 Weekend/Holiday Handling** - Automatic opt-out when office is closed

**User Roles:** All authenticated users

---

### 📊 Headcount Dashboard (`/headcount`)
Analytics and reporting interface for logistics planning.

- **📈 Real-time Metrics** - Total active users, office/WFH split, participation rates
- **🍽️ Meal-Type Breakdown** - Separate counts for Lunch and Snacks
- **📊 Visual Progress Bars** - Participation vs. opt-out ratios at a glance
- **📆 Date Navigation** - Browse historical and future headcount data
- **👥 Team Breakdown** - Expandable sections showing team-specific statistics
- **🎨 Day Status Banner** - Visual indicator for normal days, weekends, holidays

**User Roles:** `admin`, `logistics`

---

### 🛠️ Override Panel (`/override`)
Manual intervention tools for team leads and administrators.

- **👥 User Management Table** - Searchable, filterable list of team members
- **✏️ Manual Status Override** - Force opt-in/opt-out for specific users
- **📝 Reason Logging** - Mandatory justification for all manual changes
- **🔐 Role-Based Scope**
  - **Team Leads**: Can only manage their assigned team members
  - **Admins**: Global access to all users
- **📆 Date Selection** - Apply overrides for specific dates
- **📜 Audit Trail** - All changes tracked with timestamp and reason

**User Roles:** `admin`, `team_lead`

---

### 🔐 Authentication System
Secure user authentication and authorization.

- **🔑 Login/Register** - Email and password authentication
- **👤 User Profiles** - Name, email, role, default meal preference
- **🎫 JWT Tokens** - Secure session management with automatic refresh
- **🚪 Protected Routes** - Route guards based on user roles
- **♻️ Auto Re-authentication** - Persistent sessions across page refreshes

**User Roles Supported:**
- `employee` - Standard users
- `team_lead` - Team managers with override capabilities
- `admin` - Full system access
- `logistics` - Operations and reporting access

---

## 📁 Project Structure

```plaintext
craftsbite-frontend/
├── public/                      # Static assets
├── src/
│   ├── components/              # Reusable UI components
│   │   ├── cards/              # Menu cards, stat cards
│   │   ├── feedback/           # Loading, toast, error states
│   │   ├── forms/              # Input fields, buttons
│   │   ├── guards/             # Route protection components
│   │   ├── layout/             # Header, Footer, Navbar, Layout
│   │   └── modals/             # Modal dialogs
│   ├── contexts/               # React Context providers
│   │   ├── AuthContext.tsx    # Authentication state
│   │   └── ThemeContext.tsx   # Dark/light mode
│   ├── pages/                  # Route components
│   │   ├── Home.tsx            # Employee dashboard
│   │   ├── Login.tsx           # Login page
│   │   ├── Register.tsx        # Registration page
│   │   ├── HeadcountDashboard.tsx  # Analytics dashboard
│   │   ├── OverridePanel.tsx   # Manual override interface
│   │   └── ComponentShowcase.tsx   # Design system demo
│   ├── services/               # API integration layer
│   │   ├── api.ts              # Axios instance & interceptors
│   │   ├── authService.ts      # Authentication endpoints
│   │   ├── mealService.ts      # Meal participation endpoints
│   │   ├── headcountService.ts # Reporting endpoints
│   │   └── userService.ts      # User management endpoints
│   ├── store/                  # Zustand stores
│   │   └── authStore.ts        # Authentication store
│   ├── types/                  # TypeScript type definitions
│   │   ├── auth.types.ts       # Authentication types
│   │   ├── meal.types.ts       # Meal & participation types
│   │   ├── history.types.ts    # History & override types
│   │   └── api.types.ts        # Generic API types
│   ├── utils/                  # Helper functions
│   │   ├── constants.ts        # App-wide constants
│   │   └── validators.ts       # Form validation utilities
│   ├── App.tsx                 # Root application component
│   ├── main.tsx                # Application entry point
│   └── index.css               # Global styles & theme tokens
├── .env.example                # Environment variable template
├── package.json                # Dependencies and scripts
├── tsconfig.json               # TypeScript configuration
├── vite.config.ts              # Vite build configuration
└── README.md                   # This file
```

---

## 🛠️ Local Development Setup

### Prerequisites

- **Node.js** >= 18.x
- **npm** >= 9.x (comes with Node.js)
- **CraftsBite Backend** running on `http://localhost:8080` (or configured API URL)

### Step-by-Step Guide

#### 1️⃣ Clone the Repository

```bash
git clone <repository-url>
cd craftsbite-frontend
```

#### 2️⃣ Install Dependencies

```bash
npm install
```

This will install all production and development dependencies listed in `package.json`.

#### 3️⃣ Configure Environment Variables

Create a `.env` file in the **root** directory of the project:

```bash
cp .env.example .env
```

Edit the `.env` file to set your backend API URL:

```env
# API Base URL (without trailing slash)
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

> **Note:** The `VITE_` prefix is required for Vite to expose the variable to the client-side code.

#### 4️⃣ Start the Development Server

```bash
npm run dev
```

**Expected Output:**
```
  VITE v7.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

The application will be available at **`http://localhost:5173`**

#### 5️⃣ Access the Application

Open your browser and navigate to:
```
http://localhost:5173
```

**Default Routes:**
- `/login` - Login page
- `/register` - User registration
- `/home` - Employee dashboard (requires authentication)
- `/headcount` - Headcount reporting (admin/logistics only)
- `/override` - Manual overrides (admin/team_lead only)
- `/showcase` - Component showcase (development only)

---

## 🏗️ Build & Deployment

### Production Build

Compile TypeScript and create an optimized production bundle:

```bash
npm run build
```

**Output:**
- Compiled files will be in the `dist/` directory
- Assets are minified and optimized for production
- Source maps are generated for debugging

### Preview Production Build

Test the production build locally before deployment:

```bash
npm run preview
```

This serves the `dist/` folder at `http://localhost:4173`

### Lint Code

Run ESLint to check for code quality issues:

```bash
npm run lint
```

---

## 🎨 Design System

CraftsBite uses a custom **Claymorphism** design language with a warm orange color palette.

### Color Palette

| Token | Light Mode | Dark Mode | Usage |
|-------|-----------|-----------|-------|
| `--color-primary` | `#fa8c47` | `#ff9f5f` | Primary actions, accents |
| `--color-background-light` | `#fff5e6` | `#1a1210` | Page background |
| `--color-clay-light` | `#ffffff` | `#2a1f18` | Card backgrounds |
| `--color-text-main` | `#4a4a4a` | `#e8ddd0` | Primary text |
| `--color-text-sub` | `#8c705f` | `#b09880` | Secondary text |

### Claymorphism Shadows

All UI components use soft, layered shadows to create a tactile 3D effect:

```css
--shadow-clay: 20px 20px 60px #e6dccf, -20px -20px 60px #ffffff;
--shadow-clay-md: 10px 10px 20px #e6dccf, -10px -10px 20px #ffffff;
--shadow-clay-inset: inset 6px 6px 12px #e6dccf, inset -6px -6px 12px #ffffff;
```

### Typography

- **Font Family:** Inter (loaded from Google Fonts)
- **Weights:** 300 (Light), 400 (Regular), 500 (Medium), 600 (Semi-Bold), 700 (Bold), 800 (Extra-Bold), 900 (Black)

### Component Principles

- **Rounded Corners:** Heavy use of `border-radius` (16px-24px) for soft, friendly aesthetics
- **Interactive Feedback:** All buttons and cards have hover/active states
- **Smooth Transitions:** 200-300ms transitions for state changes
- **Accessibility:** ARIA labels, keyboard navigation, focus states

---

## 🔑 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Backend API base URL (without trailing slash) | `http://localhost:8080/api/v1` |

---

## 📝 Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot reload |
| `npm run build` | Build production bundle with TypeScript compilation |
| `npm run preview` | Preview production build locally |
| `npm run lint` | Run ESLint for code quality checks |

---

## 🧪 Testing

> **Note:** Unit tests are not currently implemented but can be added using Vitest (already configured in dependencies).

To add tests:
1. Create test files with `.test.tsx` or `.spec.tsx` extensions
2. Run tests with `npm run test` (add script to `package.json`)

---

## 🔒 Authentication Flow

1. **Login:** User enters email and password at `/login`
2. **Token Storage:** JWT token stored in localStorage
3. **API Interceptor:** Axios automatically attaches token to all requests
4. **Route Protection:** `ProtectedRoute` component guards authenticated pages
5. **Role Verification:** Routes check user role before rendering
6. **Token Refresh:** Auto-logout on 401/403 responses

---

## 🛣️ Route Access Matrix

| Route | Employee | Team Lead | Admin | Logistics |
|-------|----------|-----------|-------|-----------|
| `/home` | ✅ | ✅ | ✅ | ✅ |
| `/headcount` | ❌ | ❌ | ✅ | ✅ |
| `/override` | ❌ | ✅ | ✅ | ❌ |
| `/login` | Public | Public | Public | Public |
| `/register` | Public | Public | Public | Public |

---

## 🐛 Common Issues & Troubleshooting

### Issue: "Cannot connect to backend"
**Solution:** Ensure the backend is running and `VITE_API_BASE_URL` is correctly set in `.env`

### Issue: "Blank page after login"
**Solution:** Check browser console for errors. Verify API responses match expected types.

### Issue: "Changes not reflected after opt-in/opt-out"
**Solution:** Check if cutoff time has passed. The system locks changes after the deadline.

### Issue: "Dark mode not working"
**Solution:** Theme preference is stored in localStorage. Clear browser storage and try again.

---

## 📚 API Integration

All API calls are centralized in the `src/services/` directory. The application communicates with the backend using these services:

- **authService:** Login, registration, token management
- **mealService:** Get today's meals, toggle participation, get meal history
- **headcountService:** Get headcount reports by date
- **userService:** Get team members, manage user overrides

All services use the Axios instance configured in `src/services/api.ts` with automatic token injection and error handling.

---

## 👥 Contributing

When contributing to this codebase:

1. Follow the existing code style (TypeScript strict mode)
2. Use functional components with React Hooks
3. Maintain the Claymorphism design language
4. Add proper TypeScript types for all new code
5. Test across light and dark modes
6. Ensure responsive design (mobile, tablet, desktop)

---

## 📄 License

This project is proprietary software developed for organizational use.

---

## 📞 Support

For questions or issues, please contact the development team or create an issue in the repository.

---

**Built with ❤️ using React, TypeScript, and Claymorphism design principles**
