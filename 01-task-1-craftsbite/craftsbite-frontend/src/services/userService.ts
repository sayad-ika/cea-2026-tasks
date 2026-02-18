// User Service - API calls for user management

import api from './api';
import type { ApiResponse, User } from '../types';

// --- Team Members types (GET /users/me/team-members) ---
export interface TeamMember {
    id: string;
    name: string;
    email: string;
    role: string;
    team_id: string;
    team_name: string;
}

export interface TeamMembersData {
    team_lead_id: string;
    team_lead_name: string;
    total_members: number;
    members: TeamMember[];
}

export interface TeamDetails {
    team_id: string;
    team_name: string;
    description: string;
    team_lead_name: string;
}

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

/**
 * Get team members for the current team lead
 * GET /users/me/team-members
 */
export async function getTeamMembers(): Promise<ApiResponse<TeamMembersData>> {
    const response = await api.get<ApiResponse<TeamMembersData>>('/users/me/team-members');
    return response.data;
}

export async function getTeamDetails(): Promise<ApiResponse<TeamDetails>> {
    const response = await api.get<ApiResponse<TeamDetails>>('/users/me/team');
    return response.data;
}
