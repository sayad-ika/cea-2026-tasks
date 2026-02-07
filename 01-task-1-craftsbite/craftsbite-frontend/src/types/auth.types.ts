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
    token: string;
    user: User;
}

export interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (email: string, password: string) => Promise<void>;
    logout: () => void;
    setUser: (user: User | null) => void;
    checkAuth: () => Promise<void>;
    initialize: () => void;
}
