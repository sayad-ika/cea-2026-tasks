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

export interface TeamWorkLocationMember {
  user_id: string;
  name: string;
  email: string;
  location: WorkLocation;
  source: string;
}

export interface TeamWorkLocationTeam {
  team_id: string;
  team_name: string;
  members: TeamWorkLocationMember[];
}

export interface TeamWorkLocationsData {
  date: string;
  teams: TeamWorkLocationTeam[];
}

export interface OverrideWorkLocationPayload {
  user_id: string;
  date: string;
  location: WorkLocation;
  reason: string;
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

/**
 * Get work locations for teams led by authenticated team lead
 * GET /work-locations/team?date=YYYY-MM-DD
 */
export async function getTeamWorkLocations(
  date: string,
): Promise<ApiResponse<TeamWorkLocationsData>> {
  const response = await api.get<ApiResponse<TeamWorkLocationsData>>(
    "/work-locations/team",
    {
      params: { date },
    },
  );
  return response.data;
}

/**
 * Override work location for a user (team lead/admin)
 * PUT /work-locations/override
 */
export async function overrideUserWorkLocation(
  payload: OverrideWorkLocationPayload,
): Promise<ApiResponse<null>> {
  const response = await api.put<ApiResponse<null>>(
    "/work-locations/override",
    payload,
  );
  return response.data;
}

// Global WFH Types

export interface GlobalWFHPolicy {
  id: string;
  start_date: string;
  end_date: string;
  location: WorkLocation;
  is_active: boolean;
  reason: string;
  declared_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGlobalWFHPayload {
  start_date: string;
  end_date: string;
  reason: string;
}

export interface GetGlobalWFHPoliciesParams {
  start_date?: string;
  end_date?: string;
}

/**
 * Create Global WFH Policy (Admin/Logistics)
 * POST /work-locations/global-wfh
 */
export async function createGlobalWFHPolicy(
  payload: CreateGlobalWFHPayload,
): Promise<ApiResponse<GlobalWFHPolicy>> {
  const response = await api.post<ApiResponse<GlobalWFHPolicy>>(
    "/work-locations/global-wfh",
    payload,
  );
  return response.data;
}

/**
 * List Global WFH Policies
 * GET /work-locations/global-wfh
 */
export async function getGlobalWFHPolicies(
  params?: GetGlobalWFHPoliciesParams,
): Promise<ApiResponse<GlobalWFHPolicy[]>> {
  const response = await api.get<ApiResponse<GlobalWFHPolicy[]>>(
    "/work-locations/global-wfh",
    { params },
  );
  return response.data;
}

/**
 * Deactivate Global WFH Policy (Admin/Logistics)
 * DELETE /work-locations/global-wfh/:id
 */
export async function deleteGlobalWFHPolicy(
  id: string,
): Promise<ApiResponse<null>> {
  const response = await api.delete<ApiResponse<null>>(
    `/work-locations/global-wfh/${id}`,
  );
  return response.data;
}
