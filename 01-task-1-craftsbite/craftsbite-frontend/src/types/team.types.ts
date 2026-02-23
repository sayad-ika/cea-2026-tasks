export interface TeamMemberParticipation {
  user_id: string;
  name: string;
  email: string;
  meals: Record<string, boolean>;
  is_over_wfh_limit?: boolean;
}

export interface TeamParticipationGroup {
  team_id: string;
  team_name: string;
  team_lead_user_id: string;
  members: TeamMemberParticipation[];
}

export interface TeamParticipationResponse {
  date: string;
  teams: TeamParticipationGroup[];
}
