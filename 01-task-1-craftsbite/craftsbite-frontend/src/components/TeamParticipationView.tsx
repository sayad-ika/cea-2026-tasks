import React, { useState, useEffect } from "react";
import { format, addDays, subDays } from "date-fns";
import { Header, Navbar, Footer, LoadingSpinner } from ".";
import { TeamCard } from "./cards/TeamCard";
import type { TeamData, TeamParticipationResponse, ApiResponse } from "../types";
import { useAuth } from "../contexts/AuthContext";
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
  emptyStateIcon,
  fetchData,
}) => {
  const { user } = useAuth();
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [teamData, setTeamData] = useState<TeamData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

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
            <h2 className="text-4xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
              {title}
            </h2>
            <p className="text-lg text-[var(--color-text-sub)] font-medium">
              {subtitle}
            </p>
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

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm">
            {error}
          </div>
        )}

        {/* Team Data */}
        <div className="space-y-8">
          {teamData.length > 0 ? (
            teamData.map((team) => <TeamCard key={team.team_id} team={team} />)
          ) : (
            !isLoading && (
              <div className="text-center py-20 bg-white rounded-3xl border border-dashed border-gray-300">
                <span className="material-symbols-outlined text-6xl text-gray-300 mb-4 block">
                  {emptyStateIcon}
                </span>
                <p className="text-gray-500 text-lg font-medium">
                  {emptyStateMessage}
                </p>
              </div>
            )
          )}
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
