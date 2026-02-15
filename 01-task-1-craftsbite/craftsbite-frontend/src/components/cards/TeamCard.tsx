import React, { useState } from "react";
import type { TeamData, MealParticipationDetail } from "../../types";

interface TeamCardProps {
  team: TeamData;
}

export const TeamCard: React.FC<TeamCardProps> = ({ team }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="bg-[#E1D0B3] rounded-3xl overflow-hidden shadow-sm border border-[#e6dccf] transition-all duration-300">
      <div
        className="p-6 md:p-8 flex items-center gap-3 cursor-pointer hover:bg-[#d4c3a6] transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="w-10 h-10 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center">
          <span className="material-symbols-outlined">groups</span>
        </div>
        <h3 className="text-2xl font-bold text-[var(--color-background-dark)]">
          {team.team_name}
        </h3>
        <div className="ml-auto flex items-center gap-4">
          <span className="bg-gray-100 text-gray-600 px-3 py-1 rounded-full text-xs font-bold">
            {team.members.length} Members
          </span>
          <span
            className={`material-symbols-outlined transition-transform duration-300 ${
              isExpanded ? "rotate-180" : ""
            }`}
          >
            expand_more
          </span>
        </div>
      </div>

      <div
        className={`grid transition-all duration-300 ease-in-out ${
          isExpanded
            ? "grid-rows-[1fr] opacity-100 p-6 md:p-8 pt-0"
            : "grid-rows-[0fr] opacity-0 p-0"
        }`}
      >
        <div className="overflow-hidden">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {team.members.map((member) => (
              <div
                key={member.user_id}
                className="bg-[#F7F1DE] rounded-2xl p-5 border border-[#f0eadd] hover:border-orange-200 transition-colors"
              >
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-xs font-bold text-gray-600">
                    {member.name.charAt(0)}
                  </div>
                  <span className="font-bold text-[var(--color-background-dark)]">
                    {member.name}
                  </span>
                </div>

                <div className="space-y-3">
                  {member.participations.map((p: MealParticipationDetail) => (
                    <div
                      key={p.meal_type}
                      className="flex items-center justify-between text-sm"
                    >
                      <span className="capitalize text-[var(--color-text-sub)] font-medium">
                        {p.meal_type.replace("_", " ")}
                      </span>
                      <div
                        className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg border ${
                          p.is_participating
                            ? "bg-green-50 text-green-700 border-green-200"
                            : "bg-red-50 text-red-700 border-red-200"
                        }`}
                      >
                        <span className="material-symbols-outlined text-[16px]">
                          {p.is_participating ? "check_circle" : "cancel"}
                        </span>
                        <span className="font-bold text-xs">
                          {p.is_participating ? "Yes" : "No"}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
