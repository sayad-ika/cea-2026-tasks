import type { DayStatus, MealType } from "./meal.types";

export const DAY_STATUS_OPTIONS: { value: DayStatus; label: string }[] = [
    { value: "normal", label: "Normal" },
    { value: "office_closed", label: "Office Closed" },
    { value: "govt_holiday", label: "Govt Holiday" },
    { value: "celebration", label: "Celebration" },
];

export const MEAL_OPTIONS: { value: MealType; label: string }[] = [
    { value: "lunch", label: "Lunch" },
    { value: "snacks", label: "Snacks" },
    { value: "iftar", label: "Iftar" },
    { value: "event_dinner", label: "Event Dinner" },
    { value: "optional_dinner", label: "Optional Dinner" },
];

export const STATUS_CONFIG: Record<
    DayStatus | "weekend",
    { icon: string; label: string }
> = {
    normal: { icon: "wb_sunny", label: "Regular Day" },
    govt_holiday: { icon: "flag", label: "Government Holiday" },
    celebration: { icon: "celebration", label: "Celebration" },
    office_closed: { icon: "lock", label: "Office Closed" },
    weekend: { icon: "weekend", label: "Weekend" },
};

export const MEAL_LABELS: Record<string, string> = {
    lunch: "Lunch",
    snacks: "Snacks",
    iftar: "Iftar",
    event_dinner: "Event Dinner",
    optional_dinner: "Optional Dinner",
};

export interface ScheduleEntry {
    id: string;
    date: string;
    day_status: DayStatus;
    reason: string;
    available_meals: string;
}

export interface DaySchedule {
    id: string;
    date: string; // YYYY-MM-DD format
    day_status: DayStatus;
    reason?: string;
    available_meals: MealType[];
    createdBy: string;
    createdAt: string;
    updatedAt: string;
}

export interface CreateScheduleRequest {
    date: string;
    day_status: DayStatus;
    reason?: string;
    available_meals: MealType[];
}

export interface UpdateScheduleRequest {
    day_status?: DayStatus;
    reason?: string;
    available_meals?: MealType[];
}
