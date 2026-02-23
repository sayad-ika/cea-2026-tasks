import React, { useState, useEffect, useMemo, useCallback } from "react";
import { Header, Navbar, Footer, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import { getAllTeamsParticipation, getTeamParticipation, createBatchBulkOptOut } from "../services/mealService";
import toast from "react-hot-toast";
import type { TeamParticipationGroup } from "../types/team.types";
import { MEAL_TYPES } from "../utils/constants";
import { MEAL_OPTIONS } from "../types/schedule.types";
import type { MealType } from "../types";

type PanelStep = "pick" | "form" | null;

export const TeamParticipation: React.FC = () => {
  const { user } = useAuth();
  const isPrivileged = user?.role === "admin" || user?.role === "logistics" || user?.role === "team_lead";

  const [teams, setTeams] = useState<TeamParticipationGroup[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const [panelStep, setPanelStep] = useState<PanelStep>(null);
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const today = new Date().toISOString().split("T")[0];
  const [startDate, setStartDate] = useState(today);
  const [endDate, setEndDate] = useState(today);
  const [mealTypes, setMealTypes] = useState<string[]>(["lunch"]);
  const [reason, setReason] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState("");

  const fetchTeamParticipation = useCallback(async () => {
    try {
      const response =
        user?.role === "admin" || user?.role === "logistics"
          ? await getAllTeamsParticipation()
          : await getTeamParticipation();
      if (response.success && response.data) {
        setTeams(response.data.teams);
      }
    } catch {
      toast.error("Failed to load team participation.");
    } finally {
      setIsLoading(false);
    }
  }, [user?.role]);

  useEffect(() => {
    fetchTeamParticipation();
  }, [fetchTeamParticipation]);

  const allMembers = useMemo(() => teams.flatMap((t) => t.members), [teams]);

  const mealColumns = useMemo(() => {
    const firstMember = teams[0]?.members[0];
    if (!firstMember) return [];
    return (Object.keys(MEAL_TYPES) as MealType[]).filter((k) => k in firstMember.meals);
  }, [teams]);

  const toggleUser = (id: string) =>
    setSelectedUserIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );

  const toggleTeam = (teamMembers: { user_id: string }[]) => {
    const ids = teamMembers.map((m) => m.user_id);
    const allIn = ids.every((id) => selectedUserIds.includes(id));
    if (allIn) {
      setSelectedUserIds((prev) => prev.filter((id) => !ids.includes(id)));
    } else {
      setSelectedUserIds((prev) => [...new Set([...prev, ...ids])]);
    }
  };

  const toggleAll = () => {
    if (selectedUserIds.length === allMembers.length) {
      setSelectedUserIds([]);
    } else {
      setSelectedUserIds(allMembers.map((m) => m.user_id));
    }
  };

  const allSelected = allMembers.length > 0 && selectedUserIds.length === allMembers.length;

  const toggleMeal = (meal: string) =>
    setMealTypes((prev) =>
      prev.includes(meal) ? prev.filter((m) => m !== meal) : [...prev, meal]
    );

  const openPicker = () => {
    setSelectedUserIds([]);
    setStartDate(today);
    setEndDate(today);
    setMealTypes(["lunch"]);
    setReason("");
    setFormError("");
    setPanelStep("pick");
  };

  const closePanel = () => {
    setPanelStep(null);
    setSelectedUserIds([]);
    setFormError("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError("");

    if (mealTypes.length === 0) {
      setFormError("Select at least one meal type.");
      return;
    }
    if (endDate < startDate) {
      setFormError("End date must be on or after start date.");
      return;
    }

    try {
      setIsSubmitting(true);
      await createBatchBulkOptOut({
        user_ids: selectedUserIds,
        start_date: startDate,
        end_date: endDate,
        meal_types: mealTypes,
        reason: reason || undefined,
      });
      toast.success(`Bulk opt-out created for ${selectedUserIds.length} user(s).`);
      closePanel();
      fetchTeamParticipation();
    } catch (err: any) {
      setFormError(err?.error?.message || err?.message || "Failed to create bulk opt-out.");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) return <LoadingSpinner message="Loading participation data..." />;

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name} userRole={user?.role} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col gap-6">
        {/* Page header */}
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div>
            <h2 className="text-4xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
              Team Participation
            </h2>
            <p className="text-lg text-[var(--color-text-sub)] font-medium">
              Today's meal participation status for your team.
            </p>
          </div>

          {isPrivileged && panelStep === null && (
            <button
              type="button"
              onClick={openPicker}
              className="px-5 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 text-sm self-start md:self-auto"
            >
              <span className="material-symbols-outlined text-[18px]">block</span>
              Bulk Opt-Out
            </button>
          )}
        </div>

        {!panelStep && mealColumns.length > 0 && (
          <div className="bg-[#FFFDF5] rounded-[2rem] p-6 md:p-8 border border-[#e6dccf] shadow-[var(--shadow-clay-card)]">
            <div className="flex items-center justify-between mb-5">
              <div>
                <h3 className="text-xl font-black text-left tracking-tight text-[var(--color-background-dark)]">
                  Participation Roster
              </h3>
              <p className="text-sm text-[var(--color-text-sub)] font-medium mt-0.5">
                Showing{" "}
                <span className="font-bold text-[var(--color-primary)]">{allMembers.length}</span>{" "}
                members across {teams.length} team{teams.length !== 1 ? "s" : ""}
              </p>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-separate border-spacing-y-3">
              <thead>
                <tr className="text-[var(--color-text-sub)] text-sm">
                  <th className="font-bold py-2 px-4">All Members</th>
                  {mealColumns.map((meal) => (
                    <th key={meal} className="font-bold py-2 px-4 text-center">
                      {MEAL_TYPES[meal] ?? meal}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {teams.length === 0 ? (
                  <tr>
                    <td
                      colSpan={1 + mealColumns.length}
                      className="py-12 text-center text-[var(--color-text-sub)] opacity-60"
                    >
                      <span className="material-symbols-outlined text-5xl mb-2 block">group_off</span>
                      <span className="font-medium text-lg">No members found</span>
                    </td>
                  </tr>
                ) : (
                  teams.map((team) => (
                    <React.Fragment key={team.team_id}>
                      {/* Team header */}
                      <tr>
                        <td colSpan={1 + mealColumns.length} className="pt-6 pb-2 px-4">
                          <div className="flex items-center gap-2">
                            <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                            <span className="px-4 py-1.5 rounded-full bg-orange-50 text-orange-700 text-xs font-black uppercase tracking-wider border border-orange-100 shadow-sm">
                              {team.team_name}
                            </span>
                            <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                          </div>
                        </td>
                      </tr>

                      {/* Members */}
                      {team.members.map((member) => (
                        <tr
                          key={member.user_id}
                          className="rounded-xl shadow-sm bg-white/40"
                        >
                          <td className="py-4 px-4 rounded-l-xl">
                            <div className="flex items-center gap-3">
                              <div className="w-10 h-10 rounded-full bg-orange-100 flex items-center justify-center text-orange-700 font-bold text-sm shadow-sm shrink-0">
                                {member.name.charAt(0)}
                              </div>
                            <div>
                              <p className="font-bold text-[var(--color-background-dark)]">{member.name}</p>
                              {member.user_id === team.team_lead_user_id && (
                                <span className="text-xs text-orange-500 font-semibold">Team Lead</span>
                              )}
                              {member.is_over_wfh_limit && (
                                <span className="inline-flex items-center gap-1 text-[10px] font-bold text-red-600 bg-red-50 border border-red-200 rounded-lg px-2 py-0.5 mt-0.5">
                                  <span className="material-symbols-outlined text-[12px]">warning</span>
                                  WFH over limit
                                </span>
                              )}
                            </div>
                            </div>
                          </td>

                          {mealColumns.map((meal, i) => (
                            <td
                              key={meal}
                              className={`py-4 px-4 text-center ${i === mealColumns.length - 1 ? "rounded-r-xl" : ""}`}
                            >
                              <ParticipationBadge isParticipating={member.meals[meal]} />
                            </td>
                          ))}
                        </tr>
                      ))}
                    </React.Fragment>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>)
        }

        {!panelStep && mealColumns.length === 0 && (
          <div className="bg-[#FFFDF5] rounded-[2rem] p-8 border border-[#e6dccf] shadow-[var(--shadow-clay-card)] flex flex-col items-center justify-center text-center gap-3 py-16">
            <span className="material-symbols-outlined text-5xl text-[var(--color-text-sub)] opacity-40">
              no_meals
            </span>
            <p className="text-lg font-bold text-[var(--color-text-sub)]">No meals scheduled for today</p>
            <p className="text-sm text-[var(--color-text-sub)] opacity-70">
              Check back later or contact your administrator to set up today's schedule.
            </p>
          </div>
        )}

        {panelStep === "pick" && (
          <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] shadow-[var(--shadow-clay-card)] p-6 md:p-8">
            <div className="flex items-center justify-between mb-5">
              <div>
                <h3 className="text-xl font-black tracking-tight text-[var(--color-background-dark)]">
                  Select Members
                </h3>
                <p className="text-sm text-[var(--color-text-sub)] font-medium mt-0.5">
                  Choose who to opt out, then click Continue.
                </p>
              </div>
              <button
                type="button"
                onClick={closePanel}
                className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
              >
                <span className="material-symbols-outlined text-[20px]">close</span>
              </button>
            </div>

            {/* Select all */}
            <label className="flex items-center gap-3 px-4 py-3 rounded-xl bg-[var(--color-background-light)] border border-[#e6dccf] shadow-[var(--shadow-clay-button)] cursor-pointer mb-4 hover:border-[var(--color-primary)] transition-all">
              <input
                type="checkbox"
                checked={allSelected}
                onChange={toggleAll}
                className="w-4 h-4 rounded accent-orange-500 cursor-pointer"
              />
              <span className="font-bold text-sm text-[var(--color-background-dark)]">
                Select all members
              </span>
              {selectedUserIds.length > 0 && (
                <span className="ml-auto text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] bg-[var(--color-background-light)] border border-[#e6dccf] rounded-xl px-2.5 py-1">
                  {selectedUserIds.length} selected
                </span>
              )}
            </label>

            {/* Teams + members */}
            <div className="flex flex-col gap-4">
              {teams.map((team) => {
                const teamAllSelected = team.members.every((m) =>
                  selectedUserIds.includes(m.user_id)
                );
                return (
                  <div key={team.team_id}>
                    {/* Team header row */}
                    <div className="flex items-center gap-3 mb-2">
                      <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                      <label className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={teamAllSelected}
                          onChange={() => toggleTeam(team.members)}
                          className="w-3.5 h-3.5 rounded accent-orange-500 cursor-pointer"
                        />
                        <span className="px-3 py-1 rounded-full bg-orange-50 text-orange-700 text-xs font-black uppercase tracking-wider border border-orange-100">
                          {team.team_name}
                        </span>
                      </label>
                      <div className="h-px flex-grow bg-gradient-to-r from-transparent via-[#fa8c47]/30 to-transparent" />
                    </div>

                    {/* Member rows */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      {team.members.map((member) => {
                        const checked = selectedUserIds.includes(member.user_id);
                        return (
                          <label
                            key={member.user_id}
                            className={`flex items-center gap-3 px-4 py-2.5 rounded-xl border cursor-pointer transition-all ${
                              checked
                                ? "bg-orange-50 border-orange-200"
                                : "bg-white border-[#e6dccf] hover:border-[var(--color-primary)]/40"
                            }`}
                          >
                            <input
                              type="checkbox"
                              checked={checked}
                              onChange={() => toggleUser(member.user_id)}
                              className="w-4 h-4 rounded accent-orange-500 cursor-pointer shrink-0"
                            />
                            <div
                              className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold shrink-0 ${
                                checked ? "bg-orange-200 text-orange-800" : "bg-orange-100 text-orange-700"
                              }`}
                            >
                              {member.name.charAt(0)}
                            </div>
                            <div className="min-w-0">
                              <p className="font-bold text-sm text-[var(--color-background-dark)] truncate">
                                {member.name}
                              </p>
                              {member.user_id === team.team_lead_user_id && (
                                <p className="text-[10px] font-semibold text-orange-500 uppercase tracking-wider">
                                  Team Lead
                                </p>
                              )}
                            </div>
                          </label>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Footer */}
            <div className="border-t border-[#e6dccf] mt-6 pt-5 flex items-center justify-between gap-3">
              <button
                type="button"
                onClick={closePanel}
                className="px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-sm font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={selectedUserIds.length === 0}
                onClick={() => { setFormError(""); setPanelStep("form"); }}
                className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
              >
                Continue
                <span className="material-symbols-outlined text-[16px]">arrow_forward</span>
              </button>
            </div>
          </div>
        )}

        {panelStep === "form" && (
          <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] shadow-[var(--shadow-clay-card)] p-6 md:p-8">
            <div className="flex items-center justify-between mb-5">
              <div>
                <h3 className="text-xl font-black tracking-tight text-[var(--color-background-dark)]">
                  Bulk Opt-Out
                </h3>
                <p className="text-sm text-[var(--color-text-sub)] font-medium mt-0.5">
                  {selectedUserIds.length} member{selectedUserIds.length !== 1 ? "s" : ""} selected
                </p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setPanelStep("pick")}
                  className="px-3 py-1.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-xs font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] transition-all flex items-center gap-1"
                >
                  <span className="material-symbols-outlined text-[14px]">arrow_back</span>
                  Back
                </button>
                <button
                  type="button"
                  onClick={closePanel}
                  className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
                >
                  <span className="material-symbols-outlined text-[20px]">close</span>
                </button>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="flex flex-col gap-5">
              {formError && (
                <div className="flex items-center gap-3 p-4 rounded-xl bg-red-50 border border-red-200">
                  <p className="text-sm font-semibold text-red-600">{formError}</p>
                </div>
              )}

              {/* Date range */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                    Start Date
                  </label>
                  <input
                    type="date"
                    required
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                    End Date
                  </label>
                  <input
                    type="date"
                    required
                    value={endDate}
                    min={startDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                  />
                </div>
              </div>

              {/* Meal types */}
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                  Meal Types
                </label>
                <div className="flex flex-wrap gap-2">
                  {MEAL_OPTIONS.map((opt) => (
                    <button
                      key={opt.value}
                      type="button"
                      onClick={() => toggleMeal(opt.value)}
                      className={`px-3 py-1.5 rounded-xl border text-xs font-semibold transition-all ${
                        mealTypes.includes(opt.value)
                          ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white"
                          : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                      }`}
                    >
                      {opt.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Reason */}
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                  Reason{" "}
                  <span className="normal-case tracking-normal font-normal opacity-60">(optional)</span>
                </label>
                <input
                  type="text"
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="e.g. Team offsite, annual leave…"
                  className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all placeholder:text-[var(--color-text-sub)]/50"
                />
              </div>

              <div className="border-t border-[#e6dccf]" />

              <div className="flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={closePanel}
                  className="px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-sm font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || mealTypes.length === 0}
                  className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
                >
                  {isSubmitting ? (
                    <>
                      <span className="material-symbols-outlined text-[16px] animate-spin">progress_activity</span>
                      Saving…
                    </>
                  ) : (
                    <>
                      <span className="material-symbols-outlined text-[16px]">block</span>
                      Confirm Opt-Out
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        )}
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

const ParticipationBadge: React.FC<{ isParticipating: boolean | undefined }> = ({
  isParticipating,
}) => {
  if (isParticipating === undefined)
    return <span className="text-xs text-[var(--color-text-sub)]">—</span>;

  return (
    <div
      className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg border ${
        isParticipating
          ? "bg-green-50 text-green-700 border-green-200"
          : "bg-red-50 text-red-700 border-red-200"
      }`}
    >
      <span className="material-symbols-outlined text-[18px]">
        {isParticipating ? "check_circle" : "cancel"}
      </span>
      <span className="font-bold text-xs">{isParticipating ? "Yes" : "No"}</span>
    </div>
  );
};
