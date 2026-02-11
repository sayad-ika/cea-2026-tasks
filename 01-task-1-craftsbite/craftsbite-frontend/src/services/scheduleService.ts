import api from './api';
import type { ApiResponse, ScheduleEntry, DayStatus, MealType } from '../types';

export interface CreateScheduleRequest {
    date: string; // YYYY-MM-DD
    day_status: DayStatus;
    reason?: string;
    available_meals?: MealType[];
}

/**
 * Create a new schedule entry
 * POST /schedules
 */
export async function createSchedule(
    payload: CreateScheduleRequest
): Promise<ApiResponse<ScheduleEntry>> {
    const response = await api.post<ApiResponse<ScheduleEntry>>('/schedules', payload);
    return response.data;
}

/**
 * Update a schedule entry by date
 * PUT /schedules/:date
 */
export async function updateSchedule(
    date: string,
    payload: CreateScheduleRequest
): Promise<ApiResponse<ScheduleEntry>> {
    const response = await api.put<ApiResponse<ScheduleEntry>>(`/schedules/${date}`, payload);
    return response.data;
}

/**
 * Get a single schedule entry by date
 * GET /schedules/:date
 */
export async function getScheduleByDate(
    date: string
): Promise<ApiResponse<ScheduleEntry>> {
    const response = await api.get<ApiResponse<ScheduleEntry>>(`/schedules/${date}`);
    return response.data;
}

/**
 * Delete a schedule entry by date
 * DELETE /schedules/:date
 */
export async function deleteSchedule(
    date: string
): Promise<ApiResponse<null>> {
    const response = await api.delete<ApiResponse<null>>(`/schedules/${date}`);
    return response.data;
}

/**
 * Get schedules for a date range
 * GET /schedules/range?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
 */
export async function getScheduleRange(
    startDate: string,
    endDate: string
): Promise<ApiResponse<ScheduleEntry[]>> {
    const response = await api.get<ApiResponse<ScheduleEntry[]>>('/schedules/range', {
        params: { start_date: startDate, end_date: endDate },
    });
    return response.data;
}
