export interface WFHPeriod {
  id: string;
  start_date: string;
  end_date: string;
  reason?: string;
  created_by: string;
  active: boolean;
  created_at: string;
}

export interface CreateWFHPeriodRequest {
  start_date: string;
  end_date: string;
  reason?: string;
}
