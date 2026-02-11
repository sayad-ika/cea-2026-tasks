// Meal and Schedule Types

export type MealType =
  | "lunch"
  | "snacks"
  | "iftar"
  | "event_dinner"
  | "optional_dinner";

export type DayStatus =
  | "normal"
  | "office_closed"
  | "govt_holiday"
  | "celebration";

export type MealPreference = "opt_in" | "opt_out";

export type ParticipationAction =
  | "opted_in"
  | "opted_out"
  | "override_in"
  | "override_out";

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

// --- Today's Meals API response (matches GET /meals/today) ---
export type ParticipationSource =
  | "day_schedule"
  | "explicit"
  | "bulk_opt_out"
  | "user_default"
  | "system_default";

export interface TodayMealParticipation {
  meal_type: MealType;
  is_participating: boolean;
  source: ParticipationSource;
}

export interface TodayMealsData {
  date: string;
  day_status: DayStatus;
  available_meals: MealType[];
  participations: TodayMealParticipation[];
}

// --- Participation request (matches POST /meals/participation) ---
export interface SetParticipationRequest {
  date: string;
  meal_type: MealType;
  participating: boolean;
}

// --- Override request (matches POST /meals/participation/override) ---
export interface OverrideParticipationRequest {
  user_id: string;
  date: string;
  meal_type: MealType;
  participating: boolean;
  reason: string;
}

// --- Headcount (matches GET /headcount/today and /headcount/:date) ---
export interface HeadcountMealSummary {
  participating: number;
  opted_out: number;
}

export interface HeadcountData {
  date: string;
  day_status: DayStatus;
  total_active_users: number;
  meals: Record<string, HeadcountMealSummary>;
}

// --- Detailed headcount (matches GET /headcount/:date/:meal_type) ---
export interface HeadcountUserEntry {
  id: string;
  name: string;
  email: string;
  is_participating: boolean;
  source: ParticipationSource;
}

export interface DetailedHeadcountData {
  date: string;
  meal_type: MealType;
  participating_users: HeadcountUserEntry[];
  opted_out_users: HeadcountUserEntry[];
}

// Array of headcount data (e.g., today and tomorrow)
export type HeadcountDataArray = HeadcountData[];

// --- Team Participation (matches GET /meals/team-participation) ---
export interface MealParticipationDetail {
  meal_type: MealType;
  is_participating: boolean;
  source: ParticipationSource;
}

export interface TeamMemberParticipation {
  user_id: string;
  name: string;
  participations: MealParticipationDetail[];
}

export interface TeamData {
  team_id: string;
  team_name: string;
  members: TeamMemberParticipation[];
}

export interface TeamParticipationResponse {
  date: string;
  teams: TeamData[];
}
