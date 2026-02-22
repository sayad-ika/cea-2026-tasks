import type { DayStatus, MealType } from "./meal.types";

export const DAY_STATUS_OPTIONS: { value: DayStatus; label: string }[] = [
    { value: "normal", label: "Normal" },
    { value: "office_closed", label: "Office Closed" },
    { value: "govt_holiday", label: "Govt Holiday" },
    { value: "celebration", label: "Celebration" },
    { value: "weekend", label: "Weekend" },
    { value: "event_day", label: "Event Day" },
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
    event_day: { icon: "star", label: "Event Day" },
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

export const STATUS_COLORS_FOR_SCHEDULE: Record<string, string> = {
    normal: "bg-emerald-700 text-emerald-800 border-emerald-400",
    office_closed: "bg-red-700 text-red-800 border-red-400",
    govt_holiday: "bg-sky-700 text-sky-800 border-sky-400",
    celebration: "bg-violet-700 text-violet-800 border-violet-400",
    weekend: "bg-[#ede8df] text-[var(--color-text-sub)] border-[#d6ccc0]",
    event_day: "bg-orange-400 text-orange-800 border-orange-400",
};
