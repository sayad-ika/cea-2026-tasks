import React, { useState, useEffect } from "react";
import { format, addDays, subDays } from "date-fns";
import { Header, Navbar, Footer, LoadingSpinner } from ".";
// import { BulkOptOutModal } from "./modals/BulkOptOutModal";
import type { TeamData, TeamParticipationResponse, ApiResponse } from "../types";
import { useAuth } from "../contexts/AuthContext";
// import * as mealService from "../services/mealService";
import toast from "react-hot-toast";

interface TeamParticipationViewProps {
  title: string;
  subtitle: string;
  emptyStateMessage: string;
  emptyStateIcon: string;
  fetchData: (date: string) => Promise<ApiResponse<TeamParticipationResponse>>;
}

export const TeamParticipationView: React.FC<TeamParticipationViewProps> = ({
  title,
  subtitle,
  emptyStateMessage,
  emptyStateIcon, // Kept to avoid breaking interface, though unused in table view currently
  fetchData,
}) => {
  const { user } = useAuth();
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [teamData, setTeamData] = useState<TeamData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  // Bulk Opt-Out State
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set());
  // const [isBulkModalOpen, setIsBulkModalOpen] = useState(false);

  const loadData = async (date: Date) => {
    try {
      setIsLoading(true);
      setError("");
      const dateStr = format(date, "yyyy-MM-dd");
      const response = await fetchData(dateStr);

      if (response.success && response.data) {
        setTeamData(response.data.teams);
      }
    } catch (err: any) {
      console.error("Error fetching team participation:", err);
      const msg =
        err?.response?.data?.error?.message ||
        "Failed to load team participation.";
      setError(msg);
      toast.error(msg);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData(selectedDate);
  }, [selectedDate, fetchData]);

  const handlePrevDay = () => setSelectedDate((prev) => subDays(prev, 1));
  const handleNextDay = () => setSelectedDate((prev) => addDays(prev, 1));

  // Selection Handlers
  const toggleSelectionMode = () => {
    setIsSelectionMode(!isSelectionMode);
    setSelectedUserIds(new Set()); // Clear selection when toggling
  };

  const handleToggleMemberSelection = (memberId: string) => {
    const newSelection = new Set(selectedUserIds);
    if (newSelection.has(memberId)) {
      newSelection.delete(memberId);
    } else {
      newSelection.add(memberId);
    }
    setSelectedUserIds(newSelection);
  };

  const handleToggleTeamSelection = (_teamId: string, memberIds: string[]) => {
    const newSelection = new Set(selectedUserIds);
    const allSelected = memberIds.every((id) => newSelection.has(id));

    if (allSelected) {
      // Deselect all
      memberIds.forEach((id) => newSelection.delete(id));
    } else {
      // Select all
      memberIds.forEach((id) => newSelection.add(id));
    }
    setSelectedUserIds(newSelection);
  };



  if (isLoading && teamData.length === 0) {
    return <LoadingSpinner message="Loading participation data..." />;
  }

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name} userRole={user?.role} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        {/* Page Title & Controls */}
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
          <div>
            <h2 className="text-left text-4xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
              {title}
            </h2>
            <p className="text-lg text-[var(--color-text-sub)] font-medium">
              {subtitle}
            </p>
          </div>

          <div className="flex flex-col md:flex-row items-center gap-4">
            {/* Bulk Actions */}
            <div className="flex items-center gap-2">
              {isSelectionMode ? (
                <>
                  <span className="text-sm font-bold text-[var(--color-text-sub)] mr-2">
                    {selectedUserIds.size} Selected
                  </span>
                  <button
                    onClick={() => setIsSelectionMode(false)}
                    className="px-4 py-2 rounded-xl text-sm font-bold text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                        toast("Feature coming soon", { icon: "🚧" });
                    }}
                    disabled={selectedUserIds.size === 0}
                    className="px-6 py-2 rounded-xl bg-[var(--color-primary)] text-white text-sm font-bold shadow-md hover:bg-[var(--color-primary-dark)] disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                  >
                    Proceed ({selectedUserIds.size})
                  </button>
                </>
              ) : (
                <button
                  onClick={toggleSelectionMode}
                  className="px-4 py-2 rounded-xl bg-white border border-[#e6dccf] text-[var(--color-text-sub)] text-sm font-bold shadow-sm hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] transition-all flex items-center gap-2"
                >
                  <span className="material-symbols-outlined text-lg">
                    checklist_rtl
                  </span>
                  Bulk Opt-Out
                </button>
              )}
            </div>

            {/* Date Navigation */}
            <div className="flex items-center gap-4 bg-[var(--color-background-light)] p-2 rounded-2xl border border-[#e6dccf] shadow-sm">
              <button
                onClick={handlePrevDay}
                className="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-orange-50 text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors"
              >
                <span className="material-symbols-outlined">chevron_left</span>
              </button>

              <div className="flex flex-col items-center min-w-[140px]">
                <span className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider">
                  {format(selectedDate, "EEEE")}
                </span>
                <span className="text-lg font-black text-[var(--color-background-dark)]">
                  {format(selectedDate, "MMM d, yyyy")}
                </span>
              </div>

              <button
                onClick={handleNextDay}
                className="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-orange-50 text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors"
              >
                <span className="material-symbols-outlined">chevron_right</span>
              </button>
            </div>
          </div>
        </div>

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm">
            {error}
          </div>
        )}

      {/* Team Data List */}
      <div className="bg-[#FFFDF5] rounded-[1.5rem] p-6 md:p-8 border border-white/60 shadow-[var(--shadow-clay-md)]">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-xl font-bold text-[var(--color-background-dark)]">
            Participation Roster
          </h3>
          <div className="text-sm text-[var(--color-text-sub)]">
            Showing <span className="font-bold text-[var(--color-primary)]">{teamData.reduce((acc, team) => acc + team.members.length, 0)}</span> members
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-separate border-spacing-y-3">
            <thead>
              <tr className="text-[var(--color-text-sub)] text-sm">
                <th className="font-bold py-2 px-4 pl-12">Employee</th>
                <th className="font-bold py-2 px-4 text-center">Lunch</th>
                <th className="font-bold py-2 px-4 text-center">Snacks</th>
              </tr>
            </thead>
            <tbody className="text-[var(--color-text-main)]">
              {teamData.length === 0 ? (
                <tr>
                  <td colSpan={3} className="py-12 text-center text-[var(--color-text-sub)]">
                    {!isLoading && (
                        <div className="flex flex-col items-center justify-center gap-3 opacity-60 hover:opacity-100 transition-opacity">
                             {emptyStateIcon && (
                               <span className="material-symbols-outlined text-5xl mb-2">{emptyStateIcon}</span>
                             )}
                             <span className="font-medium text-lg">{emptyStateMessage}</span>
                        </div>
                    )}
                  </td>
                </tr>
              ) : (
                teamData.map((team) => (
                  <React.Fragment key={team.team_id}>
                    {/* Team Header Row */}
                    <tr>
                      <td colSpan={3} className="pt-6 pb-2 px-4">
                        <div className="flex items-center gap-2">
                          <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent"></div>
                          <span className="px-4 py-1.5 rounded-full bg-orange-50 text-orange-700 text-xs font-black uppercase tracking-wider border border-orange-100 shadow-sm">
                            {team.team_name}
                          </span>
                          <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent"></div>
                          
                          {isSelectionMode && (
                             <button
                                onClick={() => handleToggleTeamSelection(team.team_id, team.members.map(m => m.user_id))}
                                className="ml-2 px-3 py-1 rounded-lg bg-white border border-orange-200 text-xs font-bold text-orange-600 hover:bg-orange-50 transition-colors shadow-sm"
                             >
                               Select Team
                             </button>
                          )}
                        </div>
                      </td>
                    </tr>

                    {/* Member Rows */}
                    {team.members.map((member) => {
                      const isSelected = selectedUserIds.has(member.user_id);
                      const lunchParticipation = member.participations.find(p => p.meal_type === 'lunch');
                      const snacksParticipation = member.participations.find(p => p.meal_type === 'snacks');

                      return (
                        <tr
                          key={member.user_id}
                          onClick={() => isSelectionMode && handleToggleMemberSelection(member.user_id)}
                          className={`
                            transition-all duration-200 rounded-xl group relative cursor-pointer
                            ${isSelectionMode && isSelected 
                              ? "bg-orange-50/80 shadow-[inset_0_0_0_2px_#fa8c47]" 
                              : "bg-white/40 hover:bg-white/60 shadow-sm hover:shadow-md"
                            }
                          `}
                        >
                          <td className="py-4 px-4 rounded-l-xl border-l border-y border-transparent relative">
                            <div className="flex items-center gap-4">
                              {/* Selection Checkbox */}
                              <div className={`
                                w-5 h-5 rounded-md border-2 flex items-center justify-center transition-all duration-200 flex-shrink-0
                                ${isSelectionMode 
                                  ? (isSelected ? "border-[#fa8c47] bg-[#fa8c47]" : "border-gray-300 bg-white") 
                                  : "opacity-0 w-0 border-0 p-0 overflow-hidden"
                                }
                              `}>
                                {isSelected && <span className="material-symbols-outlined text-white text-[14px] font-bold">check</span>}
                              </div>

                              <div className="flex items-center gap-3">
                                <div className="w-10 h-10 rounded-full bg-orange-100 flex items-center justify-center text-orange-700 font-bold text-sm shadow-sm">
                                  {member.name.charAt(0)}
                                </div>
                                <div>
                                  <p className="font-bold text-[var(--color-background-dark)]">
                                    {member.name}
                                  </p>
                                </div>
                              </div>
                            </div>
                          </td>

                          <td className="py-4 px-4 border-y border-transparent text-center">
                             {lunchParticipation && (
                               <div className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg border ${
                                 lunchParticipation.is_participating
                                   ? "bg-green-50 text-green-700 border-green-200"
                                   : "bg-red-50 text-red-700 border-red-200"
                               }`}>
                                 <span className="material-symbols-outlined text-[18px]">
                                   {lunchParticipation.is_participating ? "check_circle" : "cancel"}
                                 </span>
                                 <span className="font-bold text-xs">
                                   {lunchParticipation.is_participating ? "Yes" : "No"}
                                 </span>
                               </div>
                             )}
                          </td>

                          <td className="py-4 px-4 rounded-r-xl border-r border-y border-transparent text-center">
                            {snacksParticipation && (
                               <div className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg border ${
                                 snacksParticipation.is_participating
                                   ? "bg-green-50 text-green-700 border-green-200"
                                   : "bg-red-50 text-red-700 border-red-200"
                               }`}>
                                 <span className="material-symbols-outlined text-[18px]">
                                   {snacksParticipation.is_participating ? "check_circle" : "cancel"}
                                 </span>
                                 <span className="font-bold text-xs">
                                   {snacksParticipation.is_participating ? "Yes" : "No"}
                                 </span>
                               </div>
                             )}
                          </td>
                        </tr>
                      );
                    })}
                  </React.Fragment>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
      </main>

      <Footer
        links={[
          { label: "Privacy", href: "#" },
          { label: "Terms", href: "#" },
          { label: "Support", href: "#" },
        ]}
      />
    </div>
  );
};
