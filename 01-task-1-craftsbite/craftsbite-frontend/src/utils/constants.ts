// Application Constants

import type { MealType, DayStatus, UserRole } from '../types';

// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

// Meal Types
export const MEAL_TYPES: Record<MealType, string> = {
    lunch: 'Lunch',
    snacks: 'Snacks',
    iftar: 'Iftar',
    event_dinner: 'Event Dinner',
    optional_dinner: 'Optional Dinner',
};

// Day Statuses
export const DAY_STATUSES: Record<DayStatus, string> = {
    normal: 'Normal',
    office_closed: 'Office Closed',
    govt_holiday: 'Government Holiday',
    celebration: 'Celebration',
};

// User Roles
export const USER_ROLES: Record<UserRole, string> = {
    employee: 'Employee',
    team_lead: 'Team Lead',
    admin: 'Admin',
    logistics: 'Logistics',
};

// Cutoff Times (in 24-hour format HH:mm)
export const CUTOFF_TIMES: Record<MealType, string> = {
    lunch: '10:00',
    snacks: '14:00',
    iftar: '16:00',
    event_dinner: '14:00',
    optional_dinner: '14:00',
};

// Storage Keys
export const STORAGE_KEYS = {
    TOKEN: 'craftsbite_token',
    USER: 'craftsbite_user',
} as const;

// Date Formats
export const DATE_FORMATS = {
    API: 'yyyy-MM-dd',
    DISPLAY: 'MMM dd, yyyy',
    DISPLAY_LONG: 'MMMM dd, yyyy',
    TIME: 'HH:mm',
    DATETIME: 'MMM dd, yyyy HH:mm',
} as const;
