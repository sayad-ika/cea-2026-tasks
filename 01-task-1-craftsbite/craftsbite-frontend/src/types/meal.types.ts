// Meal and Schedule Types

export type MealType =
    | 'lunch'
    | 'snacks'
    | 'iftar'
    | 'event_dinner'
    | 'optional_dinner';

export type DayStatus =
    | 'normal'
    | 'office_closed'
    | 'govt_holiday'
    | 'celebration';

export type MealPreference = 'opt_in' | 'opt_out';

export type ParticipationAction =
    | 'opted_in'
    | 'opted_out'
    | 'override_in'
    | 'override_out';

export interface MealParticipation {
    id: string;
    userId: string;
    date: string; // YYYY-MM-DD format
    mealType: MealType;
    isParticipating: boolean;
    optedOutAt?: string;
    overrideBy?: string;
    overrideReason?: string;
    canModify: boolean;
    cutoffTime?: string;
    createdAt: string;
    updatedAt: string;
}

export interface UpdateParticipationRequest {
    date: string;
    mealType: MealType;
    isParticipating: boolean;
}

export interface MealHeadcount {
    mealType: MealType;
    participatingCount: number;
    optedOutCount: number;
    totalEmployees: number;
}

export interface DailyHeadcountSummary {
    date: string;
    meals: MealHeadcount[];
    generatedAt: string;
}

export interface DetailedHeadcount {
    date: string;
    mealType: MealType;
    participating: Array<{
        id: string;
        name: string;
        email: string;
    }>;
    optedOut: Array<{
        id: string;
        name: string;
        email: string;
        optedOutAt: string;
    }>;
}

export interface DaySchedule {
    id: string;
    date: string; // YYYY-MM-DD format
    dayStatus: DayStatus;
    reason?: string;
    availableMeals: MealType[];
    createdBy: string;
    createdAt: string;
    updatedAt: string;
}

export interface CreateScheduleRequest {
    date: string;
    dayStatus: DayStatus;
    reason?: string;
    availableMeals: MealType[];
}

export interface UpdateScheduleRequest {
    dayStatus?: DayStatus;
    reason?: string;
    availableMeals?: MealType[];
}

export interface UserPreference {
    userId: string;
    defaultMealPreference: MealPreference;
    updatedAt: string;
}
