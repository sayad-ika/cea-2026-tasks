// Headcount Service - API calls for headcount management (Admin/Logistics)

import api from "./api";
import type {
  ApiResponse,
  HeadcountData,
  HeadcountDataArray,
  DetailedHeadcountData,
} from "../types";

/**
 * Get today's and tomorrow's headcount summary (Admin only)
 * GET /headcount/today
 */
export async function getTodayHeadcount(): Promise<
  ApiResponse<HeadcountDataArray>
> {
  const response =
    await api.get<ApiResponse<HeadcountDataArray>>("/headcount/today");
  return response.data;
}

/**
 * Get headcount for a specific date (Admin only)
 * GET /headcount/:date
 */
export async function getHeadcountByDate(
  date: string,
): Promise<ApiResponse<HeadcountData>> {
  const response = await api.get<ApiResponse<HeadcountData>>(
    `/headcount/${date}`,
  );
  return response.data;
}

/**
 * Get detailed headcount for a date + meal type (Admin only)
 * GET /headcount/:date/:meal_type
 */
export async function getDetailedHeadcount(
  date: string,
  mealType: string,
): Promise<ApiResponse<DetailedHeadcountData>> {
  const response = await api.get<ApiResponse<DetailedHeadcountData>>(
    `/headcount/${date}/${mealType}`,
  );
  return response.data;
}
