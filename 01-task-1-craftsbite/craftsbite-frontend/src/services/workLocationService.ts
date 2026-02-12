import api from "./api";
import type { ApiResponse } from "../types";

export type WorkLocation = "office" | "wfh";

export interface WorkLocationStatus {
  user_id: string;
  date: string;
  location: WorkLocation;
  source: string;
}

export interface UpdateWorkLocationPayload {
  date: string;
  location: WorkLocation;
}

/**
 * Get current work location status for authenticated user
 * GET /work-locations/me
 */
export async function getMyWorkLocation(
  date: string,
): Promise<WorkLocationStatus> {
  const response = await api.get<ApiResponse<WorkLocationStatus>>(
    "/work-locations/me",
    {
      params: { date },
    },
  );

  if (!response.data.data) {
    throw new Error("Work location data is empty.");
  }

  return response.data.data;
}

/**
 * Update work location status for authenticated user
 * PUT /work-locations/me
 */
export async function updateMyWorkLocation(
  payload: UpdateWorkLocationPayload,
): Promise<ApiResponse<null>> {
  const response = await api.put<ApiResponse<null>>(
    "/work-locations/me",
    payload,
  );
  return response.data;
}
