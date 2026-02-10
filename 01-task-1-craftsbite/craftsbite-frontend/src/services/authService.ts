// Authentication Service

import api from './api';
import type { LoginResponse, User, UserRole } from '../types';

/**
 * Login user with email and password
 */
export async function login(email: string, password: string): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/auth/login', {
        email,
        password,
    });

    return response.data;
}

/**
 * Register a new user
 */
export async function register(
    name: string,
    email: string,
    role: UserRole,
    password: string
): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/auth/register', {
        name,
        email,
        role,
        password,
    });

    return response.data;
}

/**
 * Logout current user
 */
export async function logout(): Promise<void> {
    await api.post('/auth/logout');
}

/**
 * Get current authenticated user
 */
export async function getCurrentUser(): Promise<User> {
    const response = await api.get<{ data: User }>('/auth/me');
    return response.data.data;
}
