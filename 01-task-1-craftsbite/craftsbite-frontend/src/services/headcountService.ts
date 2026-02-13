// Headcount Service - API calls for headcount management (Admin/Logistics)

import api from "./api";
import type {
  ApiResponse,
  HeadcountData,
  HeadcountDataArray,
  DetailedHeadcountData,
  DailyAnnouncementResponse,
  HeadcountReportData,
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

/**
 * Generate daily announcement for a specific date (Admin/Logistics only)
 * GET /headcount/announcement/:date
 */
export async function generateDailyAnnouncement(
  date: string,
): Promise<ApiResponse<DailyAnnouncementResponse>> {
  const response = await api.get<ApiResponse<DailyAnnouncementResponse>>(
    `/headcount/announcement/${date}`,
  );
  return response.data;
}

/**
 * Get headcount report for today and tomorrow (Admin/Logistics only)
 * GET /headcount/report
 */
export async function getHeadcountReport(): Promise<
  ApiResponse<HeadcountReportData>
> {
  const response =
    await api.get<ApiResponse<HeadcountReportData>>("/headcount/report");
  return response.data;
}
