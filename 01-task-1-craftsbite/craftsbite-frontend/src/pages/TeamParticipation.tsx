import { useState, useEffect } from "react";
import { Header, Navbar, Footer, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import { getTeamParticipation } from "../services/mealService";
import toast from "react-hot-toast";
import type { TeamParticipationGroup } from "../types/team.types";

export const TeamParticipation: React.FC = () => {
  const { user } = useAuth();
  const [teams, setTeams] = useState<TeamParticipationGroup[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchTeamParticipation = async () => {
      try {
        const response = await getTeamParticipation();
        if (response.success && response.data) {
          setTeams(response.data.teams);
        }
      } catch (err: any) {
        toast.error("Failed to load team participation.");
      } finally {
        setIsLoading(false);
      }
    };
    fetchTeamParticipation();
  }, []);

  if (isLoading)
    return <LoadingSpinner message="Loading participation data..." />;

  const totalMembers = teams.reduce((acc, team) => acc + team.members.length, 0);

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name} userRole={user?.role} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        <div className="mb-10">
          <h2 className="text-4xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
            Team Participation
          </h2>
          <p className="text-lg text-[var(--color-text-sub)] font-medium">
            Today's meal participation status for your team.
          </p>
        </div>

        <div className="bg-[#FFFDF5] rounded-[1.5rem] p-6 md:p-8 border border-white/60 shadow-[var(--shadow-clay-md)]">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-xl font-bold text-[var(--color-background-dark)]">
              Participation Roster
            </h3>
            <span className="text-sm text-[var(--color-text-sub)]">
              Showing{" "}
              <span className="font-bold text-[var(--color-primary)]">
                {totalMembers}
              </span>{" "}
              members
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-separate border-spacing-y-3">
              <thead>
                <tr className="text-[var(--color-text-sub)] text-sm">
                  <th className="font-bold py-2 px-4">Employee</th>
                  <th className="font-bold py-2 px-4 text-center">Lunch</th>
                  <th className="font-bold py-2 px-4 text-center">Snacks</th>
                </tr>
              </thead>
              <tbody>
                {teams.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="py-12 text-center text-[var(--color-text-sub)] opacity-60">
                      <span className="material-symbols-outlined text-5xl mb-2 block">group_off</span>
                      <span className="font-medium text-lg">No members found</span>
                    </td>
                  </tr>
                ) : (
                  teams.map((team) => (
                    <>
                      {/* Team Header */}
                      <tr key={team.team_id}>
                        <td colSpan={3} className="pt-6 pb-2 px-4">
                          <div className="flex items-center gap-2">
                            <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                            <span className="px-4 py-1.5 rounded-full bg-orange-50 text-orange-700 text-xs font-black uppercase tracking-wider border border-orange-100 shadow-sm">
                              {team.team_name}
                            </span>
                            <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                          </div>
                        </td>
                      </tr>

                      {/* Team Members */}
                      {team.members.map((member) => (
                        <tr key={member.user_id} className="bg-white/40 rounded-xl shadow-sm">
                          <td className="py-4 px-4 rounded-l-xl border-l border-y border-transparent">
                            <div className="flex items-center gap-3">
                              <div className="w-10 h-10 rounded-full bg-orange-100 flex items-center justify-center text-orange-700 font-bold text-sm shadow-sm">
                                {member.name.charAt(0)}
                              </div>
                              <div>
                                <p className="font-bold text-[var(--color-background-dark)]">
                                  {member.name}
                                </p>
                                {member.user_id === team.team_lead_user_id && (
                                  <span className="text-xs text-orange-500 font-semibold">Team Lead</span>
                                )}
                              </div>
                            </div>
                          </td>
                          <td className="py-4 px-4 border-y border-transparent text-center">
                            <ParticipationBadge isParticipating={member.meals["lunch"]} />
                          </td>
                          <td className="py-4 px-4 rounded-r-xl border-r border-y border-transparent text-center">
                            <ParticipationBadge isParticipating={member.meals["snacks"]} />
                          </td>
                        </tr>
                      ))}
                    </>
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

const ParticipationBadge: React.FC<{ isParticipating: boolean | undefined }> = ({ isParticipating }) => {
  if (isParticipating === undefined)
    return <span className="text-xs text-[var(--color-text-sub)]">—</span>;

  return (
    <div className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg border ${
      isParticipating ? "bg-green-50 text-green-700 border-green-200" : "bg-red-50 text-red-700 border-red-200"
    }`}>
      <span className="material-symbols-outlined text-[18px]">
        {isParticipating ? "check_circle" : "cancel"}
      </span>
      <span className="font-bold text-xs">{isParticipating ? "Yes" : "No"}</span>
    </div>
  );
};