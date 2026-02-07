// Bulk Opt-Out Types

import type { MealType } from './meal.types';

export interface BulkOptOut {
    id: string;
    userId: string;
    startDate: string; // YYYY-MM-DD format
    endDate: string; // YYYY-MM-DD format
    mealType: MealType;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

export interface CreateBulkOptOutRequest {
    startDate: string;
    endDate: string;
    mealType: MealType;
}
