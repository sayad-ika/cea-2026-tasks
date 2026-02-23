import api from './api';
import type { ApiResponse } from '../types';
import type { WorkLocationValue } from '../types/work-location.types';

export interface WorkLocationData {
  user_id?: string;
  date: string;
  location: WorkLocationValue;
}

export interface SetWorkLocationRequest {
  date: string;
  location: WorkLocationValue;
}

export interface OverrideWorkLocationRequest {
  user_id: string;
  date: string;
  location: WorkLocationValue;
  reason: string;
}

export interface WorkLocationListItem {
  user_id: string;
  user_name?: string;
  date: string;
  location: WorkLocationValue;
}


export interface MonthlyWFHSummary {
    year_month: string;
    wfh_days: number;
    allowance: number;
    is_over_limit: boolean;
}

export interface MemberWFHSummary {
    user_id: string;
    wfh_days: number;
    is_over_limit: boolean;
    extra_days: number;
}

export interface TeamMonthlyReport {
    year_month: string;
    allowance: number;
    total_employees: number;
    over_limit_count: number;
    total_extra_days: number;
    members: MemberWFHSummary[];
}

// ---------- API Calls ----------

export async function setWorkLocation(
  payload: SetWorkLocationRequest
): Promise<ApiResponse<null>> {
  const response = await api.post<ApiResponse<null>>('/work-location', payload);
  return response.data;
}

export async function getWorkLocation(
  date: string
): Promise<ApiResponse<WorkLocationData>> {
  const response = await api.get<ApiResponse<WorkLocationData>>(
    `/work-location?date=${date}`
  );
  return response.data;
}

export async function overrideWorkLocation(
  payload: OverrideWorkLocationRequest
): Promise<ApiResponse<null>> {
  const response = await api.post<ApiResponse<null>>(
    '/work-location/override',
    payload
  );
  return response.data;
}

export async function getWorkLocationList(
  date: string
): Promise<ApiResponse<WorkLocationListItem[]>> {
  const response = await api.get<ApiResponse<WorkLocationListItem[]>>(
    `/work-location/list?date=${date}`
  );
  return response.data;
}

export async function getMonthlyWFHSummary(
    month?: string,
): Promise<ApiResponse<MonthlyWFHSummary>> {
    const query = month ? `?month=${month}` : "";
    const response = await api.get<ApiResponse<MonthlyWFHSummary>>(
        `/work-location/monthly-summary${query}`,
    );
    return response.data;
}

export async function fetchWFHReport(month?: string): Promise<ApiResponse<TeamMonthlyReport>> {
    const params = month ? `?month=${month}` : "";
    
    const response = await api.get<ApiResponse<TeamMonthlyReport>>(
        `/work-location/team-monthly-report${params}`,
    );

    return response.data;
}

