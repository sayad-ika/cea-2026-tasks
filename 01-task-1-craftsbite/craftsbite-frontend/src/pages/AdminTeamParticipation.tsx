import React from "react";
import { TeamParticipationView } from "../components/TeamParticipationView";
import * as mealService from "../services/mealService";

export const AdminTeamParticipation: React.FC = () => {
  return (
    <TeamParticipationView
      title="All Teams Status"
      subtitle="Daily meal overview for the entire organization"
      emptyStateMessage="No participation data found for this date."
      emptyStateIcon="assignment_turned_in"
      fetchData={mealService.getAllTeamsParticipation}
    />
  );
};
