import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import type { MealType as MealTypeEnum } from "../types";
import { MEAL_TYPES } from "../utils/constants";
import * as userService from "../services/userService";
import * as mealService from "../services/mealService";
import * as workLocationService from "../services/workLocationService";
import type { WorkLocation } from "../services/workLocationService";
import toast from "react-hot-toast";

type OverrideTab = "meal" | "work_location";

interface SelectableUser {
  id: string;
  name: string;
  email: string;
  teamName?: string;
  currentLocation?: WorkLocation;
  locationSource?: string;
}

const getDefaultDate = () => {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  return tomorrow.toISOString().split("T")[0];
};

const getErrorMessage = (error: unknown, fallback: string) => {
  if (typeof error === "object" && error !== null) {
    const candidate = error as {
      error?: { message?: string };
      message?: string;
    };
    return candidate.error?.message || candidate.message || fallback;
  }
  return fallback;
};

const toReadable = (value: string) =>
  value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

const getInitials = (name: string) =>
  name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() || "")
    .join("");

const statusPillClass = (location?: WorkLocation) => {
  if (location === "office") {
    return "bg-gradient-to-br from-blue-100 to-blue-50 text-blue-700 shadow-[4px_4px_8px_#dbeafe,-4px_-4px_8px_#ffffff]";
  }
  if (location === "wfh") {
    return "bg-gradient-to-br from-green-100 to-green-50 text-green-700 shadow-[4px_4px_8px_#dcfce7,-4px_-4px_8px_#ffffff]";
  }
  return "bg-gradient-to-br from-gray-100 to-gray-50 text-gray-600 shadow-[4px_4px_8px_#e5e7eb,-4px_-4px_8px_#ffffff]";
};

export const OverridePanel: React.FC = () => {
  const { user } = useAuth();
  const isTeamLead = user?.role === "team_lead";

  const [activeTab, setActiveTab] = useState<OverrideTab>("meal");
  const [teamName, setTeamName] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingWorkUsers, setIsLoadingWorkUsers] = useState(false);
  const [isSubmittingMeal, setIsSubmittingMeal] = useState(false);
  const [isSubmittingWork, setIsSubmittingWork] = useState(false);

  const [mealUsers, setMealUsers] = useState<SelectableUser[]>([]);
  const [workUsers, setWorkUsers] = useState<SelectableUser[]>([]);

  const [mealError, setMealError] = useState("");
  const [workError, setWorkError] = useState("");

  // Meal override form state
  const [selectedUserId, setSelectedUserId] = useState("");
  const [selectedDate, setSelectedDate] = useState(getDefaultDate);
  const [selectedMealType, setSelectedMealType] =
    useState<MealTypeEnum>("lunch");
  const [participating, setParticipating] = useState(false);
  const [reason, setReason] = useState("");

  // Work location dashboard state
  const [selectedWorkDate, setSelectedWorkDate] = useState(getDefaultDate);
  const [workSearch, setWorkSearch] = useState("");
  const [overrideTargetId, setOverrideTargetId] = useState<string | null>(null);
  const [overrideLocation, setOverrideLocation] =
    useState<WorkLocation>("office");
  const [overrideReason, setOverrideReason] = useState("");

  const mealTypeOptions = Object.entries(MEAL_TYPES).map(([value, label]) => ({
    value,
    label,
  }));

  useEffect(() => {
    const fetchBaseUsers = async () => {
      try {
        if (isTeamLead) {
          const response = await userService.getTeamMembers();
          if (response.success && response.data) {
            const members = response.data.members.map((member) => ({
              id: member.id,
              name: member.name,
              email: member.email,
              teamName: member.team_name,
            }));
            setMealUsers(members);
            setWorkUsers(members);
            setTeamName(response.data.members[0]?.team_name || null);
          }
        } else {
          const response = await userService.getUsers();
          if (response.success && response.data) {
            const members = response.data.map((member) => ({
              id: member.id,
              name: member.name,
              email: member.email,
            }));
            setMealUsers(members);
            setWorkUsers(members);
          }
        }
      } catch (error: unknown) {
        const msg = getErrorMessage(error, "Failed to load users.");
        setMealError(msg);
      } finally {
        setIsLoading(false);
      }
    };

    fetchBaseUsers();
  }, [isTeamLead]);

  const fetchTeamWorkLocations = useCallback(
    async (date: string) => {
      if (!isTeamLead) return;

      try {
        setIsLoadingWorkUsers(true);
        setWorkError("");

        const response = await workLocationService.getTeamWorkLocations(date);
        if (response.success && response.data) {
          const flatMembers = response.data.teams.flatMap((team) =>
            team.members.map((member) => ({
              id: member.user_id,
              name: member.name,
              email: member.email,
              teamName: team.team_name,
              currentLocation: member.location,
              locationSource: member.source,
            })),
          );
          setWorkUsers(flatMembers);
          if (!teamName) {
            setTeamName(response.data.teams[0]?.team_name || null);
          }
        } else {
          setWorkUsers([]);
        }
      } catch (error: unknown) {
        const msg = getErrorMessage(error, "Failed to load team locations.");
        setWorkError(msg);
        toast.error(msg);
      } finally {
        setIsLoadingWorkUsers(false);
      }
    },
    [isTeamLead, teamName],
  );

  useEffect(() => {
    if (activeTab !== "work_location") return;
    if (isTeamLead) {
      void fetchTeamWorkLocations(selectedWorkDate);
      return;
    }
    setWorkUsers(mealUsers);
  }, [activeTab, isTeamLead, selectedWorkDate, fetchTeamWorkLocations, mealUsers]);

  const filteredWorkUsers = useMemo(() => {
    const query = workSearch.trim().toLowerCase();
    if (!query) return workUsers;
    return workUsers.filter(
      (member) =>
        member.name.toLowerCase().includes(query) ||
        member.email.toLowerCase().includes(query) ||
        (member.teamName || "").toLowerCase().includes(query),
    );
  }, [workUsers, workSearch]);

  const officeCount = useMemo(
    () =>
      filteredWorkUsers.filter((member) => member.currentLocation === "office")
        .length,
    [filteredWorkUsers],
  );
  const wfhCount = useMemo(
    () => filteredWorkUsers.filter((member) => member.currentLocation === "wfh").length,
    [filteredWorkUsers],
  );
  const overrideCount = useMemo(
    () =>
      filteredWorkUsers.filter(
        (member) => member.locationSource?.toLowerCase() === "explicit",
      ).length,
    [filteredWorkUsers],
  );

  const handleMealSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setMealError("");

    if (!selectedUserId) {
      setMealError("Please select a user.");
      return;
    }
    if (!reason.trim()) {
      setMealError("Please provide a reason for the override.");
      return;
    }

    try {
      setIsSubmittingMeal(true);
      await mealService.overrideParticipation({
        user_id: selectedUserId,
        date: selectedDate,
        meal_type: selectedMealType,
        participating,
        reason: reason.trim(),
      });
      toast.success("Meal participation overridden successfully.");
      setReason("");
    } catch (error: unknown) {
      const msg = getErrorMessage(error, "Failed to override participation.");
      setMealError(msg);
      toast.error(msg);
    } finally {
      setIsSubmittingMeal(false);
    }
  };

  const openOverridePopover = (member: SelectableUser) => {
    setOverrideTargetId(member.id);
    setOverrideLocation(member.currentLocation || "office");
    setOverrideReason("");
    setWorkError("");
  };

  const closeOverridePopover = () => {
    setOverrideTargetId(null);
    setOverrideReason("");
  };

  const handleWorkOverrideSubmit = async (userId: string) => {
    if (!overrideReason.trim()) {
      setWorkError("Please provide a reason for the override.");
      return;
    }

    try {
      setIsSubmittingWork(true);
      setWorkError("");

      await workLocationService.overrideUserWorkLocation({
        user_id: userId,
        date: selectedWorkDate,
        location: overrideLocation,
        reason: overrideReason.trim(),
      });

      toast.success("Work location overridden successfully.");
      closeOverridePopover();

      if (isTeamLead) {
        await fetchTeamWorkLocations(selectedWorkDate);
      } else {
        setWorkUsers((prev) =>
          prev.map((member) =>
            member.id === userId
              ? {
                  ...member,
                  currentLocation: overrideLocation,
                  locationSource: "explicit",
                }
              : member,
          ),
        );
      }
    } catch (error: unknown) {
      const msg = getErrorMessage(error, "Failed to override work location.");
      setWorkError(msg);
      toast.error(msg);
    } finally {
      setIsSubmittingWork(false);
    }
  };

  const handleExportWorkReport = () => {
    const headers = ["name", "email", "team", "location", "source"];
    const rows = filteredWorkUsers.map((member) => [
      member.name,
      member.email,
      member.teamName || "",
      member.currentLocation || "unknown",
      member.locationSource || "unknown",
    ]);

    const csv = [headers.join(","), ...rows.map((row) => row.join(","))].join("\n");
    const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `work-location-${selectedWorkDate}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading users..." />;
  }

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header
        userName={user?.name || "User"}
        userRole={user?.role || "admin"}
      />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        <div className="mb-8 flex items-center gap-3 flex-wrap">
          <button
            type="button"
            onClick={() => setActiveTab("meal")}
            className={`px-6 py-3 rounded-2xl font-bold text-sm transition-all ${
              activeTab === "meal"
                ? "bg-[var(--color-primary)] text-white"
                : "bg-[var(--color-background-light)] text-[var(--color-text-main)]"
            }`}
            style={{
              boxShadow:
                activeTab === "meal"
                  ? "6px 6px 12px rgba(250, 140, 71, 0.3)"
                  : "var(--shadow-clay-button)",
            }}
          >
            Meal Participation
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("work_location")}
            className={`px-6 py-3 rounded-2xl font-bold text-sm transition-all ${
              activeTab === "work_location"
                ? "bg-[var(--color-primary)] text-white"
                : "bg-[var(--color-background-light)] text-[var(--color-text-main)]"
            }`}
            style={{
              boxShadow:
                activeTab === "work_location"
                  ? "6px 6px 12px rgba(250, 140, 71, 0.3)"
                  : "var(--shadow-clay-button)",
            }}
          >
            Work Location Override
          </button>
        </div>

        {activeTab === "meal" ? (
          <>
            <div className="mb-10 text-center md:text-left">
              <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
                Override Participation
              </h2>
              <p className="text-lg text-[var(--color-text-sub)] font-medium">
                {isTeamLead
                  ? "Override meal participation for your team members."
                  : "Mark meal participation on behalf of any employee."}
              </p>
              {isTeamLead && teamName && (
                <div
                  className="mt-3 inline-flex items-center gap-2 bg-[var(--color-background-light)] rounded-2xl px-4 py-2 text-sm font-bold text-[var(--color-primary)]"
                  style={{ boxShadow: "var(--shadow-clay-button)" }}
                >
                  <span className="material-symbols-outlined text-base">groups</span>
                  Team: {teamName}
                </div>
              )}
            </div>

            <div
              className="bg-[var(--color-background-light)] rounded-3xl p-8 md:p-10 max-w-2xl w-full mx-auto md:mx-0"
              style={{ boxShadow: "var(--shadow-clay)" }}
            >
              <form onSubmit={handleMealSubmit} className="space-y-6">
                {mealError && (
                  <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                    {mealError}
                  </div>
                )}

                <div className="space-y-2">
                  <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                    {isTeamLead ? "Team Member" : "Employee"}
                  </label>
                  <select
                    value={selectedUserId}
                    onChange={(event) => setSelectedUserId(event.target.value)}
                    className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                    style={{ boxShadow: "var(--shadow-clay-inset)" }}
                  >
                    <option value="">
                      {isTeamLead
                        ? "Select a team member..."
                        : "Select an employee..."}
                    </option>
                    {mealUsers.map((member) => (
                      <option key={member.id} value={member.id}>
                        {member.name} ({member.email})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="space-y-2">
                  <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                    Date
                  </label>
                  <input
                    type="date"
                    value={selectedDate}
                    onChange={(event) => setSelectedDate(event.target.value)}
                    className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50"
                    style={{ boxShadow: "var(--shadow-clay-inset)" }}
                  />
                </div>

                <div className="space-y-2">
                  <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                    Meal Type
                  </label>
                  <select
                    value={selectedMealType}
                    onChange={(event) =>
                      setSelectedMealType(event.target.value as MealTypeEnum)
                    }
                    className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                    style={{ boxShadow: "var(--shadow-clay-inset)" }}
                  >
                    {mealTypeOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="space-y-2">
                  <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                    Participation Status
                  </label>
                  <div className="flex gap-4">
                    <button
                      type="button"
                      onClick={() => setParticipating(true)}
                      className={`flex-1 py-3.5 rounded-2xl font-bold text-sm transition-all duration-200 ${
                        participating
                          ? "bg-green-500 text-white"
                          : "bg-[var(--color-background-light)] text-[var(--color-text-sub)]"
                      }`}
                      style={{
                        boxShadow: participating
                          ? "6px 6px 12px #a3d9a5, -6px -6px 12px #ffffff"
                          : "var(--shadow-clay-button)",
                      }}
                    >
                      Opt In
                    </button>
                    <button
                      type="button"
                      onClick={() => setParticipating(false)}
                      className={`flex-1 py-3.5 rounded-2xl font-bold text-sm transition-all duration-200 ${
                        !participating
                          ? "bg-red-500 text-white"
                          : "bg-[var(--color-background-light)] text-[var(--color-text-sub)]"
                      }`}
                      style={{
                        boxShadow: !participating
                          ? "6px 6px 12px #fca5a5, -6px -6px 12px #ffffff"
                          : "var(--shadow-clay-button)",
                      }}
                    >
                      Opt Out
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                    Reason <span className="text-red-400">*</span>
                  </label>
                  <textarea
                    value={reason}
                    onChange={(event) => setReason(event.target.value)}
                    placeholder="e.g., User on leave, requested via email..."
                    rows={3}
                    className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] placeholder-[var(--color-text-sub)]/50 outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 resize-none"
                    style={{ boxShadow: "var(--shadow-clay-inset)" }}
                  />
                </div>

                <button
                  type="submit"
                  disabled={isSubmittingMeal}
                  className="w-full py-4 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-lg hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  style={{ boxShadow: "var(--shadow-clay-button)" }}
                >
                  {isSubmittingMeal ? (
                    <>
                      <span className="animate-spin material-symbols-outlined">
                        progress_activity
                      </span>
                      <span>Submitting...</span>
                    </>
                  ) : (
                    <>
                      <span className="material-symbols-outlined">swap_horiz</span>
                      <span>Override Participation</span>
                    </>
                  )}
                </button>
              </form>
            </div>
          </>
        ) : (
          <>
            <div className="mb-10 flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
              <div>
                <h2 className="text-3xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
                  Location Override Dashboard
                </h2>
                <p className="text-[var(--color-text-sub)] font-medium">
                  Manage team work locations and approve exceptions for{" "}
                  <span className="text-[var(--color-primary)] font-bold">
                    {selectedWorkDate}
                  </span>
                  .
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-3 w-full md:w-auto">
                <input
                  type="date"
                  value={selectedWorkDate}
                  onChange={(event) => setSelectedWorkDate(event.target.value)}
                  className="px-4 py-3 rounded-2xl bg-[#FFFDF5] border-none focus:ring-2 focus:ring-[var(--color-primary)]/40 outline-none"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                />
                <div className="relative group">
                  <input
                    type="text"
                    value={workSearch}
                    onChange={(event) => setWorkSearch(event.target.value)}
                    placeholder="Search employee..."
                    className="pl-12 pr-4 py-3 w-64 rounded-2xl bg-[#FFFDF5] border-none focus:ring-2 focus:ring-[var(--color-primary)]/40 outline-none"
                    style={{ boxShadow: "var(--shadow-clay-inset)" }}
                  />
                  <span className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-[var(--color-text-sub)]">
                    search
                  </span>
                </div>
                {/* <button
                  type="button"
                  onClick={handleExportWorkReport}
                  className="px-6 py-3 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold shadow-[var(--shadow-clay-button)] hover:scale-[1.02] active:scale-[0.98] active:shadow-[var(--shadow-clay-button-active)] transition-all duration-200 flex items-center gap-2"
                >
                  <span className="material-symbols-outlined">download</span>
                  <span className="hidden md:inline">Export Report</span>
                </button> */}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
              <div className="bg-[#FFFDF5] rounded-3xl p-6 flex items-center gap-4 border border-white/50 shadow-[var(--shadow-clay-md)]">
                <div className="w-12 h-12 rounded-2xl bg-blue-100 text-blue-600 flex items-center justify-center shadow-sm">
                  <span className="material-symbols-outlined">business</span>
                </div>
                <div>
                  <p className="text-[var(--color-text-sub)] text-sm font-semibold uppercase tracking-wider">
                    In Office
                  </p>
                  <p className="text-3xl font-black text-[var(--color-background-dark)]">
                    {officeCount}
                  </p>
                </div>
              </div>

              <div className="bg-[#FFFDF5] rounded-3xl p-6 flex items-center gap-4 border border-white/50 shadow-[var(--shadow-clay-md)]">
                <div className="w-12 h-12 rounded-2xl bg-green-100 text-green-600 flex items-center justify-center shadow-sm">
                  <span className="material-symbols-outlined">home_work</span>
                </div>
                <div>
                  <p className="text-[var(--color-text-sub)] text-sm font-semibold uppercase tracking-wider">
                    Working Remote
                  </p>
                  <p className="text-3xl font-black text-[var(--color-background-dark)]">
                    {wfhCount}
                  </p>
                </div>
              </div>

              <div className="bg-[#FFFDF5] rounded-3xl p-6 flex items-center gap-4 border border-white/50 shadow-[var(--shadow-clay-md)]">
                <div className="w-12 h-12 rounded-2xl bg-orange-100 text-orange-600 flex items-center justify-center shadow-sm">
                  <span className="material-symbols-outlined">edit_note</span>
                </div>
                <div>
                  <p className="text-[var(--color-text-sub)] text-sm font-semibold uppercase tracking-wider">
                    Overrides Active
                  </p>
                  <p className="text-3xl font-black text-[var(--color-background-dark)]">
                    {overrideCount}
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-[#FFFDF5] rounded-[1.5rem] p-8 border border-white/60 shadow-[var(--shadow-clay-md)]">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-bold text-[var(--color-background-dark)]">
                  Team Roster
                </h3>
                <button
                  type="button"
                  onClick={() =>
                    isTeamLead
                      ? void fetchTeamWorkLocations(selectedWorkDate)
                      : setWorkUsers(mealUsers)
                  }
                  className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] flex items-center justify-center shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                >
                  <span className="material-symbols-outlined">refresh</span>
                </button>
              </div>

              {workError && (
                <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                  {workError}
                </div>
              )}

              {isLoadingWorkUsers && (
                <div className="mb-4 text-sm text-[var(--color-text-sub)] flex items-center gap-2">
                  <span className="animate-spin material-symbols-outlined text-base">
                    progress_activity
                  </span>
                  Loading team locations...
                </div>
              )}

              <div className="overflow-x-auto overflow-y-visible">
                <table className="w-full text-left border-separate border-spacing-y-3">
                  <thead>
                    <tr className="text-[var(--color-text-sub)] text-sm">
                      <th className="font-bold py-2 px-4">Employee</th>
                      <th className="font-bold py-2 px-4">Team</th>
                      <th className="font-bold py-2 px-4">Current Status</th>
                      <th className="font-bold py-2 px-4">Source</th>
                      <th className="font-bold py-2 px-4 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="text-[var(--color-text-main)]">
                    {filteredWorkUsers.length === 0 ? (
                      <tr>
                        <td
                          colSpan={5}
                          className="py-10 text-center text-[var(--color-text-sub)]"
                        >
                          {isTeamLead
                            ? "No team members found for the selected date."
                            : "No users found for your search."}
                        </td>
                      </tr>
                    ) : (
                      filteredWorkUsers.map((member) => (
                        <tr
                          key={member.id}
                          className="bg-white/40 hover:bg-white/60 transition-colors rounded-xl group relative"
                        >
                          <td className="py-4 px-4 rounded-l-xl border-l border-y border-transparent">
                            <div className="flex items-center gap-3">
                              <div className="w-10 h-10 rounded-full bg-orange-100 flex items-center justify-center text-orange-700 font-bold text-sm shadow-sm">
                                {getInitials(member.name)}
                              </div>
                              <div>
                                <p className="font-bold text-[var(--color-background-dark)]">
                                  {member.name}
                                </p>
                                <p className="text-xs text-[var(--color-text-sub)]">
                                  {member.email}
                                </p>
                              </div>
                            </div>
                          </td>

                          <td className="py-4 px-4 border-y border-transparent text-sm">
                            {member.teamName || "Organization"}
                          </td>

                          <td className="py-4 px-4 border-y border-transparent">
                            <span
                              className={`px-3 py-1 rounded-lg text-xs font-bold inline-flex items-center gap-1 ${statusPillClass(member.currentLocation)}`}
                            >
                              <span className="w-2 h-2 rounded-full bg-current opacity-70" />
                              {member.currentLocation
                                ? member.currentLocation.toUpperCase()
                                : "UNKNOWN"}
                            </span>
                          </td>

                          <td className="py-4 px-4 border-y border-transparent text-sm text-[var(--color-text-sub)]">
                            {member.locationSource
                              ? toReadable(member.locationSource)
                              : "Not Available"}
                          </td>

                          <td className="py-4 px-4 rounded-r-xl border-r border-y border-transparent text-right relative">
                            <button
                              type="button"
                              onClick={() => openOverridePopover(member)}
                              className="px-4 py-2 rounded-xl bg-[var(--color-background-light)] text-[var(--color-primary)] font-bold text-sm shadow-[var(--shadow-clay-button)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                            >
                              Override
                            </button>

                            {overrideTargetId === member.id && (
                              <div className="absolute right-16 top-1/2 -translate-y-1/2 z-50">
                                <div
                                  className="w-80 p-5 rounded-2xl bg-[#FFFDF5] border border-white/60 flex flex-col gap-4 text-left"
                                  style={{
                                    boxShadow:
                                      "15px 15px 30px #e6dccf, -15px -15px 30px #ffffff",
                                  }}
                                >
                                  <div className="flex justify-between items-center border-b border-[#e6dccf] pb-2">
                                    <h4 className="font-bold text-[var(--color-background-dark)]">
                                      Override Location
                                    </h4>
                                    <button
                                      type="button"
                                      onClick={closeOverridePopover}
                                      className="text-[var(--color-text-sub)] hover:text-red-500"
                                    >
                                      <span className="material-symbols-outlined text-lg">
                                        close
                                      </span>
                                    </button>
                                  </div>

                                  <div>
                                    <label className="block text-xs font-bold text-[var(--color-text-sub)] mb-2 uppercase">
                                      New Status
                                    </label>
                                    <div className="flex bg-[#F0EBE0] p-1 rounded-xl shadow-inner">
                                      <button
                                        type="button"
                                        onClick={() => setOverrideLocation("office")}
                                        className={`flex-1 py-2 rounded-lg text-sm font-bold transition-colors ${
                                          overrideLocation === "office"
                                            ? "bg-white shadow-sm text-[var(--color-primary)]"
                                            : "text-[var(--color-text-sub)] hover:text-[var(--color-text-main)]"
                                        }`}
                                      >
                                        Office
                                      </button>
                                      <button
                                        type="button"
                                        onClick={() => setOverrideLocation("wfh")}
                                        className={`flex-1 py-2 rounded-lg text-sm font-bold transition-colors ${
                                          overrideLocation === "wfh"
                                            ? "bg-white shadow-sm text-green-700"
                                            : "text-[var(--color-text-sub)] hover:text-[var(--color-text-main)]"
                                        }`}
                                      >
                                        WFH
                                      </button>
                                    </div>
                                  </div>

                                  <div>
                                    <label className="block text-xs font-bold text-[var(--color-text-sub)] mb-2 uppercase">
                                      Reason
                                    </label>
                                    <textarea
                                      value={overrideReason}
                                      onChange={(event) =>
                                        setOverrideReason(event.target.value)
                                      }
                                      placeholder="Enter reason for override..."
                                      rows={3}
                                      className="w-full p-3 text-sm text-[var(--color-text-main)] rounded-xl resize-none outline-none focus:ring-1 focus:ring-[var(--color-primary)]/50"
                                      style={{ boxShadow: "var(--shadow-clay-inset)" }}
                                    />
                                  </div>

                                  <div className="flex gap-2 mt-1">
                                    <button
                                      type="button"
                                      onClick={closeOverridePopover}
                                      className="flex-1 py-2 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-bold shadow-[var(--shadow-clay-button)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                                    >
                                      Cancel
                                    </button>
                                    <button
                                      type="button"
                                      onClick={() => handleWorkOverrideSubmit(member.id)}
                                      disabled={isSubmittingWork}
                                      className="flex-1 py-2 rounded-xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white text-sm font-bold shadow-[var(--shadow-clay-button)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all disabled:opacity-60"
                                    >
                                      {isSubmittingWork ? "Saving..." : "Save"}
                                    </button>
                                  </div>
                                </div>
                                <div className="absolute top-1/2 -right-2 -mt-2 w-4 h-4 bg-[#FFFDF5] transform rotate-45 border-r border-t border-white/60" />
                              </div>
                            )}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="mt-8 flex justify-between items-center border-t border-[#e6dccf] pt-6">
                <span className="text-sm text-[var(--color-text-sub)]">
                  Showing{" "}
                  <span className="font-bold text-[var(--color-text-main)]">
                    {filteredWorkUsers.length}
                  </span>{" "}
                  employee{filteredWorkUsers.length === 1 ? "" : "s"}
                </span>
              </div>
            </div>
          </>
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
