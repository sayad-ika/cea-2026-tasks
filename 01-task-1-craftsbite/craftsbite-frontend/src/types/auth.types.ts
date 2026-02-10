// Authentication Types

export type UserRole = 'employee' | 'team_lead' | 'admin' | 'logistics';

export interface User {
    id: string;
    email: string;
    name: string;
    role: UserRole;
    active: boolean;
    defaultMealPreference: 'opt_in' | 'opt_out';
    createdAt: string;
    updatedAt: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    success: string;
    data: {
        token: string,
        user: {
            id: string,
            employee_id: string,
            name: string,
            role: string,
        }
    },
    message: string
}

export interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (email: string, password: string) => Promise<void>;
    register: (name: string, email: string, role: any, password: string) => Promise<void>;
    logout: () => void;
    setUser: (user: User | null) => void;
    checkAuth: () => Promise<void>;
    initialize: () => void;
}
