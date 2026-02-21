import api from './api';
import type {
    MealType,
    MealParticipation,
    ApiResponse,
    SetParticipationRequest,
    OverrideParticipationRequest,
    BatchBulkOptOutRequest,
} from '../types';
import type { TeamParticipationResponse } from '../types/team.types';

export interface TodaysMealsResponse {
    success: boolean;
    data: {
        date: string;
        day_status: string;
        available_meals: MealType[];
        participations: Array<{
            meal_type: MealType;
            is_participating: boolean;
            source: string;
        }>;
    };
    message?: string;
}

/**
 * Get today's meals for the authenticated user
 */
export async function getTodaysMeals(): Promise<TodaysMealsResponse> {
    const response = await api.get<TodaysMealsResponse>('/meals/today');
    return response.data;
}

/**
 * Toggle meal participation for a specific meal type
 */
export async function toggleMealParticipation(
    mealType: MealType,
    date: string
): Promise<{ success: boolean; data: MealParticipation }> {
    const response = await api.post(`/meals/${mealType}/toggle`, { date });
    return response.data;
}

/**
 * Update meal participation status
 */
export async function updateMealParticipation(
    date: string,
    mealType: MealType,
    isParticipating: boolean
): Promise<{ success: boolean; data: MealParticipation }> {
    const response = await api.put('/meals/participation', {
        date,
        mealType,
        isParticipating,
    });
    return response.data;
}

/**
 * Set meal participation explicitly
 * POST /meals/participation
 */
export async function setMealParticipation(
    payload: SetParticipationRequest
): Promise<ApiResponse<null>> {
    const response = await api.post<ApiResponse<null>>('/meals/participation', payload);
    return response.data;
}

/**
 * Override participation (Admin/Team Lead only)
 * POST /meals/participation/override
 */
export async function overrideParticipation(
    payload: OverrideParticipationRequest
): Promise<ApiResponse<null>> {
    const response = await api.post<ApiResponse<null>>('/meals/participation/override', payload);
    return response.data;
}

export async function getTeamParticipation(): Promise<ApiResponse<TeamParticipationResponse>> {
    const response = await api.get<ApiResponse<TeamParticipationResponse>>(`/meals/team-participation`);
    return response.data;
}

export async function getAllTeamsParticipation(): Promise<ApiResponse<TeamParticipationResponse>> {
    const response = await api.get<ApiResponse<TeamParticipationResponse>>(`/meals/all-teams-participation`);
    return response.data;
}

export async function createBatchBulkOptOut(payload: BatchBulkOptOutRequest): Promise<ApiResponse<unknown[]>> {
    const response = await api.post<ApiResponse<unknown[]>>('admin/meals/bulk-optouts', payload);
    return response.data;
}
