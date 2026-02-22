import api from './api';
import type { MealHistoryEntry, MealHistoryEntryHistoryResponse, MealHistoryEntryHistoryFilters } from '../types/history.types';

// Get a user's meal participation history (Admin and Logistics only)
export async function getUserMealHistory(
    userId: string,
    filters?: MealHistoryEntryHistoryFilters
): Promise<MealHistoryEntryHistoryResponse> {
    const params: Record<string, string> = {};
    if (filters?.start_date) params.start_date = filters.start_date;
    if (filters?.end_date) params.end_date = filters.end_date;

    const response = await api.get<MealHistoryEntryHistoryResponse>(
        `/admin/meals/history/${userId}`,
        { params }
    );
    return response.data;
}

export type { MealHistoryEntry };
