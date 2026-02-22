import api from "./api";
import type { ApiResponse, CreateScheduleRequest } from "../types";
import type { ScheduleEntry } from "../types/schedule.types";

export async function createSchedule(
    payload: CreateScheduleRequest,
): Promise<ApiResponse<ScheduleEntry>> {
    const response = await api.post<ApiResponse<ScheduleEntry>>(
        "/schedules",
        payload,
    );
    return response.data;
}

export async function getScheduleByDate(
    date: string,
): Promise<ApiResponse<ScheduleEntry>> {
    const response = await api.get<ApiResponse<ScheduleEntry>>(
        `/schedules/${date}`,
    );
    return response.data;
}

export async function getScheduleRange(
    startDate: string,
    endDate: string,
): Promise<ApiResponse<ScheduleEntry[]>> {
    const response = await api.get<ApiResponse<ScheduleEntry[]>>(
        `/schedules/range?start_date=${startDate}&end_date=${endDate}`,
    );
    return response.data;
}
