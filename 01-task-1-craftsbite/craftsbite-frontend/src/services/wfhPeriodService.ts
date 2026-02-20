import api from "./api";
import type { ApiResponse } from "../types";
import type { WFHPeriod, CreateWFHPeriodRequest } from "../types/wfh-period.types";

// GET /wfh-periods
export async function listWFHPeriods(): Promise<ApiResponse<WFHPeriod[]>> {
  const response = await api.get<ApiResponse<WFHPeriod[]>>("/wfh-periods");
  return response.data;
}

// Create a new WFH period.
export async function createWFHPeriod(
  payload: CreateWFHPeriodRequest
): Promise<ApiResponse<WFHPeriod>> {
  const response = await api.post<ApiResponse<WFHPeriod>>(
    "/wfh-periods",
    payload
  );
  return response.data;
}

// Delete a WFH period by ID.
export async function deleteWFHPeriod(
  id: string
): Promise<ApiResponse<null>> {
  const response = await api.delete<ApiResponse<null>>(`/wfh-periods/${id}`);
  return response.data;
}
