// Meal History Types

import type { MealType, ParticipationAction } from './meal.types';

export interface HistoryEntry {
    date: string; // YYYY-MM-DD format
    mealType: MealType;
    status: 'participated' | 'opted_out' | 'override_in' | 'override_out';
    changedAt?: string;
    changedBy?: 'self' | 'team_lead' | 'admin';
    changedByName?: string;
    reason?: string;
    isDefault?: boolean;
}

export interface ParticipationAudit {
    id: string;
    userId: string;
    date: string; // YYYY-MM-DD format
    mealType: MealType;
    action: ParticipationAction;
    previousValue?: string;
    changedByUserId: string;
    changedByName: string;
    reason?: string;
    ipAddress?: string;
    createdAt: string;
}

export interface HistorySummary {
    totalDays: number;
    participated: number;
    optedOut: number;
    participationRate: string;
}

export interface MealHistory {
    userId: string;
    startDate: string;
    endDate: string;
    history: HistoryEntry[];
    summary: HistorySummary;
}

export interface HistoryParams {
    startDate?: string;
    endDate?: string;
    mealType?: MealType;
    limit?: number;
}

export interface AuditParams {
    userId?: string;
    date?: string;
    mealType?: MealType;
}
