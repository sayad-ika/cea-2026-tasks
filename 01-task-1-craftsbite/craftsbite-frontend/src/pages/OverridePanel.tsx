import React, { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import type { MealType as MealTypeEnum } from "../types";
import { MEAL_TYPES } from "../utils/constants";
import * as userService from "../services/userService";
import * as mealService from "../services/mealService";
import * as workLocationService from "../services/workLocationService";
import toast from "react-hot-toast";

interface SelectableUser {
  id: string;
  name: string;
  email: string;
}

export const OverridePanel: React.FC = () => {
  const { user } = useAuth();
  const [selectableUsers, setSelectableUsers] = useState<SelectableUser[]>([]);
  const [teamName, setTeamName] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [wlSelectedUserId, setWlSelectedUserId] = useState("");
  const [wlSelectedDate, setWlSelectedDate] = useState(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return tomorrow.toISOString().split("T")[0];
  });
  const [wlLocation, setWlLocation] = useState<"office" | "wfh">("office");
  const [wlIsSubmitting, setWlIsSubmitting] = useState(false);
  const [wReason, setWReason] = useState("");
  const [wlError, setWlError] = useState("");
  const [tab, setTab] = useState<"meal" | "work-location">("meal");

  // Form state
  const [selectedUserId, setSelectedUserId] = useState("");
  const [selectedDate, setSelectedDate] = useState(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return tomorrow.toISOString().split("T")[0];
  });
  const [selectedMealType, setSelectedMealType] = useState<MealTypeEnum>("lunch");
  const [participating, setParticipating] = useState(false);
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  const isTeamLead = user?.role === "team_lead";

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        if (isTeamLead) {
          const res = await userService.getTeamMembers();
          if (res.success && res.data) {
            setSelectableUsers(
              res.data.members.map((m) => ({
                id: m.id,
                name: m.name,
                email: m.email,
              })),
            );
            setTeamName(res.data.members[0]?.team_name || null);
          }
        } else {
          const res = await userService.getUsers();
          if (res.success && res.data) {
            const list = Array.isArray(res.data) ? res.data : [];
            setSelectableUsers(
              list.map((u) => ({ id: u.id, name: u.name, email: u.email })),
            );
          }
        }
      } catch (err: any) {
        setError(err?.error?.message || "Failed to load users.");
      } finally {
        setIsLoading(false);
      }
    };
    fetchUsers();
  }, [isTeamLead]);

  const handleWlOverride = async (e: React.FormEvent) => {
    e.preventDefault();
    setWlError("");

    if (!wlSelectedUserId) {
      setWlError("Please select a user.");
      return;
    }
    if (!wReason.trim()) {
      setWlError("Please provide a reason for the override.");
      return;
    }

    try {
      setWlIsSubmitting(true);
      await workLocationService.overrideWorkLocation({
        user_id: wlSelectedUserId,
        date: wlSelectedDate,
        location: wlLocation,
        reason: wReason.trim(),
      });
      toast.success("Work location corrected successfully");
      setWReason("");
    } catch (err: any) {
      const msg = err?.error?.message || "Failed to override work location.";
      setWlError(msg);
      toast.error(msg);
    } finally {
      setWlIsSubmitting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!selectedUserId) {
      setError("Please select a user.");
      return;
    }
    if (!reason.trim()) {
      setError("Please provide a reason for the override.");
      return;
    }

    try {
      setIsSubmitting(true);
      await mealService.overrideParticipation({
        user_id: selectedUserId,
        date: selectedDate,
        meal_type: selectedMealType,
        participating,
        reason: reason.trim(),
      });
      toast.success("Participation overridden successfully");
      setReason("");
    } catch (err: any) {
      const msg = err?.error?.message || "Failed to override participation.";
      setError(msg);
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const mealTypeOptions = Object.entries(MEAL_TYPES).map(([value, label]) => ({
    value,
    label,
  }));

  if (isLoading) {
    return <LoadingSpinner message="Loading users..." />;
  }

  if (isTeamLead && selectableUsers.length === 0) {
    return (
      <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
        <Header userName={user?.name || "User"} userRole={user?.role || "admin"} />
        <Navbar />
        <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col items-center justify-center">
          <div
            className="bg-[var(--color-background-light)] rounded-3xl p-10 md:p-14 max-w-2xl w-full text-center"
            style={{ boxShadow: "var(--shadow-clay)" }}
          >
            <div className="mb-4 flex justify-center">
              <span className="material-symbols-outlined text-6xl text-[var(--color-primary)]/60">
                groups_2
              </span>
            </div>
            <h3 className="text-3xl font-black text-[var(--color-background-dark)] mb-3">
              No Team Members Yet
            </h3>
            <p className="text-lg text-[var(--color-text-sub)] mb-2">
              You don't have any team members assigned yet.
            </p>
            <p className="text-sm text-[var(--color-text-sub)]/70">
              Once your team members are set up, you'll be able to override their meal
              participation here.
            </p>
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
  }

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name || "User"} userRole={user?.role || "admin"} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        {/* Page Title */}
        <div className="mb-4 text-center md:text-left">
          <h2 className="text-2xl md:text-4xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
            Override Participation & Work Location
          </h2>
          <p className="text-lg text-[var(--color-text-sub)] font-medium">
            {isTeamLead
              ? "Override meal participation for your team members."
              : "Mark meal participation on behalf of an employee."}
          </p>
        </div>

        {isTeamLead && teamName && (
          <div
            className="mt-3 inline-flex items-center gap-2 bg-[var(--color-background-light)] rounded-2xl px-4 py-2 text-sm font-bold text-[var(--color-primary)]"
            style={{ boxShadow: "var(--shadow-clay-button)" }}
          >
            <span className="material-symbols-outlined text-base">groups</span>
            Team: {teamName}
          </div>
        )}

        {/* Tabs */}
        <div className="flex gap-3 mb-6 mt-4">
          <button
            type="button"
            onClick={() => setTab("meal")}
            className={`flex-1 py-3 px-4 rounded-2xl font-bold text-sm transition-all duration-200 flex items-center justify-center gap-2 ${
              tab === "meal"
                ? "bg-[var(--color-primary)] text-white"
                : "bg-[var(--color-background-light)] text-[var(--color-text-sub)] opacity-50 hover:opacity-75"
            }`}
            style={{
              boxShadow:
                tab === "meal"
                  ? "6px 6px 12px #c26629, -6px -6px 12px #ffb275"
                  : "var(--shadow-clay-button)",
            }}
          >
            <span className="material-symbols-outlined text-lg">restaurant</span>
            Meal
          </button>
          <button
            type="button"
            onClick={() => setTab("work-location")}
            className={`flex-1 py-3 px-4 rounded-2xl font-bold text-sm transition-all duration-200 flex items-center justify-center gap-2 ${
              tab === "work-location"
                ? "bg-[var(--color-primary)] text-white"
                : "bg-[var(--color-background-light)] text-[var(--color-text-sub)] opacity-50 hover:opacity-75"
            }`}
            style={{
              boxShadow:
                tab === "work-location"
                  ? "6px 6px 12px #c26629, -6px -6px 12px #ffb275"
                  : "var(--shadow-clay-button)",
            }}
          >
            <span className="material-symbols-outlined text-lg">location_on</span>
            Work Location
          </button>
        </div>

        {/* Meal Override Form */}
        {tab === "meal" && (
          <div
            className="bg-[var(--color-background-light)] rounded-3xl p-8 md:p-10 max-w-2xl w-full mx-auto md:mx-0"
            style={{ boxShadow: "var(--shadow-clay)" }}
          >
            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                  {error}
                </div>
              )}

              <div className="space-y-2">
                <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                  {isTeamLead ? "Team Member" : "Employee"}
                </label>
                <select
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value)}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                >
                  <option value="">
                    {isTeamLead ? "Select a team member..." : "Select an employee..."}
                  </option>
                  {selectableUsers.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name} ({u.email})
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
                  onChange={(e) => setSelectedDate(e.target.value)}
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
                  onChange={(e) => setSelectedMealType(e.target.value as MealTypeEnum)}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                >
                  {mealTypeOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
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
                    ✓ Opt In
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
                    ✗ Opt Out
                  </button>
                </div>
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                  Reason <span className="text-red-400">*</span>
                </label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="e.g., User on leave, requested via email..."
                  rows={3}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] placeholder-[var(--color-text-sub)]/50 outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 resize-none"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-4 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-lg hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ boxShadow: "var(--shadow-clay-button)" }}
              >
                {isSubmitting ? (
                  <>
                    <span className="animate-spin material-symbols-outlined">progress_activity</span>
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
        )}

        {/* Work Location Override Form */}
        {tab === "work-location" && (
          <div
            className="bg-[var(--color-background-light)] rounded-3xl p-8 md:p-10 max-w-2xl w-full mx-auto md:mx-0"
            style={{ boxShadow: "var(--shadow-clay)" }}
          >
            <form onSubmit={handleWlOverride} className="space-y-6">
              {wlError && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                  {wlError}
                </div>
              )}

              <div className="space-y-2">
                <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                  {isTeamLead ? "Team Member" : "Employee"}
                </label>
                <select
                  value={wlSelectedUserId}
                  onChange={(e) => setWlSelectedUserId(e.target.value)}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                >
                  <option value="">
                    {isTeamLead ? "Select a team member..." : "Select an employee..."}
                  </option>
                  {selectableUsers.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name} ({u.email})
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
                  value={wlSelectedDate}
                  onChange={(e) => setWlSelectedDate(e.target.value)}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                  Work Location
                </label>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => setWlLocation("office")}
                    className={`flex-1 py-2.5 rounded-xl font-bold text-sm transition-all duration-200 flex items-center justify-center gap-1.5 ${
                      wlLocation === "office"
                        ? "bg-[var(--color-primary)] text-white"
                        : "bg-[var(--color-background-light)] text-[var(--color-text-sub)]"
                    }`}
                    style={{
                      boxShadow:
                        wlLocation === "office"
                          ? "4px 4px 8px #c26629, -4px -4px 8px #ffb275"
                          : "var(--shadow-clay-button)",
                    }}
                  >
                    <span className="material-symbols-outlined text-base">apartment</span>
                    Office
                  </button>
                  <button
                    type="button"
                    onClick={() => setWlLocation("wfh")}
                    className={`flex-1 py-2.5 rounded-xl font-bold text-sm transition-all duration-200 flex items-center justify-center gap-1.5 ${
                      wlLocation === "wfh"
                        ? "bg-[var(--color-primary)] text-white"
                        : "bg-[var(--color-background-light)] text-[var(--color-text-sub)]"
                    }`}
                    style={{
                      boxShadow:
                        wlLocation === "wfh"
                          ? "4px 4px 8px #c26629, -4px -4px 8px #ffb275"
                          : "var(--shadow-clay-button)",
                    }}
                  >
                    <span className="material-symbols-outlined text-base">home</span>
                    WFH
                  </button>
                </div>
              </div>

              {/* Reason — new field */}
              <div className="space-y-2">
                <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                  Reason <span className="text-red-400">*</span>
                </label>
                <textarea
                  value={wReason}
                  onChange={(e) => setWReason(e.target.value)}
                  placeholder="e.g., Employee working from home due to illness..."
                  rows={3}
                  className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] placeholder-[var(--color-text-sub)]/50 outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 resize-none"
                  style={{ boxShadow: "var(--shadow-clay-inset)" }}
                />
              </div>

              <button
                type="submit"
                disabled={wlIsSubmitting}
                className="w-full py-4 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-lg hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ boxShadow: "var(--shadow-clay-button)" }}
              >
                {wlIsSubmitting ? (
                  <>
                    <span className="animate-spin material-symbols-outlined">progress_activity</span>
                    <span>Submitting...</span>
                  </>
                ) : (
                  <>
                    <span className="material-symbols-outlined">location_on</span>
                    <span>Override Work Location</span>
                  </>
                )}
              </button>
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
