import React from "react";
import { TeamParticipationView } from "../components/TeamParticipationView";
import * as mealService from "../services/mealService";

export const TeamParticipation: React.FC = () => {
  return (
    <TeamParticipationView
      title="Team Participation"
      subtitle="View daily meal status for your team members"
      emptyStateMessage="No team participation data found for this date."
      emptyStateIcon="groups_off"
      fetchData={mealService.getTeamParticipation}
    />
  );
};
