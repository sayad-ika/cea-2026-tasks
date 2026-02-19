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
