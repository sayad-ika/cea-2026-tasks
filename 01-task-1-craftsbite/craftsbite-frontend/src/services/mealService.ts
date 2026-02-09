// Meal Service - API calls for meal management

import api from './api';
import type { MealType, MealParticipation } from '../types';

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