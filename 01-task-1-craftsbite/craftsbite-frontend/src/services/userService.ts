// User Service - API calls for user management

import api from './api';
import type { ApiResponse, User } from '../types';

/**
 * Get all users (Admin only)
 * GET /users
 */
export async function getUsers(): Promise<ApiResponse<User[]>> {
    const response = await api.get<ApiResponse<User[]>>('/users');
    return response.data;
}

/**
 * Get user by ID
 * GET /users/:id
 */
export async function getUserById(id: string): Promise<ApiResponse<User>> {
    const response = await api.get<ApiResponse<User>>(`/users/${id}`);
    return response.data;
}
