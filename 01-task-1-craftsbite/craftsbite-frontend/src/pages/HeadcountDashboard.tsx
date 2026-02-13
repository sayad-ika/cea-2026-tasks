import React, { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import type {
  HeadcountReportDay,
  MealType as MealTypeEnum,
  DailyAnnouncementResponse,
} from "../types";
import { MEAL_TYPES, DAY_STATUSES } from "../utils/constants";
import * as headcountService from "../services/headcountService";
import toast from "react-hot-toast";

/* ─── Status helpers ───────────────────────────────────── */

const STATUS_META: Record<string, { icon: string; gradient: string; bg: string }> = {
  celebration: {
    icon: "celebration",
    gradient: "from-amber-400 to-orange-500",
    bg: "bg-amber-50",
  },
  weekend: {
    icon: "weekend",
    gradient: "from-blue-400 to-indigo-500",
    bg: "bg-blue-50",
  },
  govt_holiday: {
    icon: "flag",
    gradient: "from-emerald-400 to-teal-500",
    bg: "bg-emerald-50",
  },
  office_closed: {
    icon: "lock",
    gradient: "from-slate-400 to-slate-600",
    bg: "bg-slate-50",
  },
  normal: {
    icon: "sunny",
    gradient: "from-[var(--color-primary)] to-[var(--color-primary-dark)]",
    bg: "bg-orange-50",
  },
};

const getStatusMeta = (status: string) =>
  STATUS_META[status] ?? STATUS_META.normal;

/* ─── Component ────────────────────────────────────────── */

export const HeadcountDashboard: React.FC = () => {
  const { user } = useAuth();
  const [reportData, setReportData] = useState<HeadcountReportDay[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedDateIndex, setSelectedDateIndex] = useState(0);
  const [expandedTeams, setExpandedTeams] = useState<Record<string, boolean>>(
    {},
  );

  // Announcement state
  const [announceData, setAnnounceData] =
    useState<DailyAnnouncementResponse | null>(null);
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
  const [isGeneratingAnnouncement, setIsGeneratingAnnouncement] =
    useState(false);

  /* ── Fetch ──────────────────────────────────────── */

  useEffect(() => {
    const fetchReport = async () => {
      try {
        setIsLoading(true);
        const res = await headcountService.getHeadcountReport();
        if (res.success && res.data) {
          setReportData(res.data);
          setSelectedDateIndex(0);
        }
      } catch (err: any) {
        setError(err?.error?.message || "Failed to load headcount report.");
      } finally {
        setIsLoading(false);
      }
    };
    fetchReport();
  }, []);

  /* ── Announcement modal lock ────────────────────── */

  useEffect(() => {
    if (!showAnnouncementModal) return;

    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setShowAnnouncementModal(false);
    };

    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleEscape);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleEscape);
    };
  }, [showAnnouncementModal]);

  /* ── Helpers ────────────────────────────────────── */

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr + "T00:00:00");
    return date.toLocaleDateString("en-US", {
      weekday: "long",
      month: "short",
      day: "numeric",
    });
  };

  const getDateLabel = (dateStr: string, index: number) => {
    const date = new Date(dateStr + "T00:00:00");
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date.getTime() === today.getTime()
      ? "Today"
      : index === 1
        ? "Tomorrow"
        : formatDate(dateStr);
  };

  const toggleTeam = (teamId: string) =>
    setExpandedTeams((prev) => ({ ...prev, [teamId]: !prev[teamId] }));

  const closeAnnouncementModal = () => setShowAnnouncementModal(false);

  const handleGenerateAnnouncement = async () => {
    if (!currentDay) return;
    setIsGeneratingAnnouncement(true);
    try {
      const res = await headcountService.generateDailyAnnouncement(
        currentDay.date,
      );
      if (res.success && res.data) {
        setAnnounceData(res.data);
        setShowAnnouncementModal(true);
        toast.success("Announcement generated!");
      } else {
        toast.error(res.message || "Failed to generate announcement");
      }
    } catch (err: any) {
      toast.error(err?.error?.message || "Failed to generate announcement");
    } finally {
      setIsGeneratingAnnouncement(false);
    }
  };

  const handleCopyToClipboard = async () => {
    if (!announceData?.message) return;
    try {
      await navigator.clipboard.writeText(announceData.message);
      toast.success("Copied to clipboard!");
    } catch {
      toast.error("Unable to copy automatically. Please copy manually.");
    }
  };

  const currentDay = reportData[selectedDateIndex];

  if (isLoading) {
    return <LoadingSpinner message="Loading headcount report..." />;
  }

  const statusMeta = currentDay ? getStatusMeta(currentDay.day_status) : null;

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header
        userName={user?.name || "User"}
        userRole={user?.role || "admin"}
      />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        {/* Page Title */}
        <div className="mb-10 text-center md:text-left">
          <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
            Headcount Dashboard
          </h2>
          <p className="text-lg text-[var(--color-text-sub)] font-medium">
            Team-wise meal participation & headcount report
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0">
            {error}
          </div>
        )}

        {/* Date Selector Tabs */}
        {reportData.length > 0 && (
          <div className="mb-8 flex gap-3 flex-wrap">
            {reportData.map((day, index) => (
              <button
                key={day.date}
                onClick={() => setSelectedDateIndex(index)}
                className={`px-6 py-3 rounded-2xl font-bold text-sm transition-all duration-200 ${
                  selectedDateIndex === index
                    ? "bg-[var(--color-primary)] text-white"
                    : "bg-[var(--color-background-light)] text-[var(--color-text-main)]"
                }`}
                style={{
                  boxShadow:
                    selectedDateIndex === index
                      ? "6px 6px 12px rgba(var(--color-primary-rgb), 0.3)"
                      : "var(--shadow-clay-button)",
                }}
              >
                {getDateLabel(day.date, index)}
              </button>
            ))}
          </div>
        )}

        {currentDay && statusMeta && (
          <>
            {/* ────── Day Status Banner ────── */}
            <div
              className={`mb-8 rounded-3xl overflow-hidden border border-white/40`}
              style={{ boxShadow: "var(--shadow-clay)" }}
            >
              <div
                className={`bg-gradient-to-r ${statusMeta.gradient} px-7 py-5 flex items-center gap-4`}
              >
                <div className="flex items-center justify-center w-12 h-12 bg-white/25 rounded-2xl backdrop-blur-sm">
                  <span className="material-symbols-outlined text-white text-[26px]">
                    {statusMeta.icon}
                  </span>
                </div>
                <div>
                  <h3 className="text-xl font-black text-white">
                    {DAY_STATUSES[currentDay.day_status] ||
                      currentDay.day_status}{" "}
                    &middot; {formatDate(currentDay.date)}
                  </h3>
                  {currentDay.special_day_note && (
                    <p className="text-sm text-white/85 font-medium mt-0.5">
                      {currentDay.special_day_note}
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* ────── Overview Stats Row ────── */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-5 mb-10">
              {/* Overall Total */}
              <StatCard
                icon="groups"
                label="Total People"
                value={currentDay.overall_total}
                color="var(--color-primary)"
              />
              {/* Office */}
              <StatCard
                icon="apartment"
                label="In Office"
                value={currentDay.office_wfh_split.office}
                color="#3b82f6"
              />
              {/* WFH */}
              <StatCard
                icon="home"
                label="Working from Home"
                value={currentDay.office_wfh_split.wfh}
                color="#8b5cf6"
              />
              {/* Unassigned */}
              <StatCard
                icon="person_off"
                label="Unassigned Users"
                value={currentDay.unassigned_users}
                color="#ef4444"
              />
            </div>

            {/* ────── Generate Announcement ────── */}
            <div className="mb-8">
              <button
                onClick={handleGenerateAnnouncement}
                disabled={isGeneratingAnnouncement}
                className="flex items-center gap-3 px-6 py-3 bg-[var(--color-primary)] text-white font-bold rounded-2xl transition-all duration-200 hover:shadow-[6px_6px_12px_rgba(var(--color-primary-rgb),_0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ boxShadow: "var(--shadow-clay-button)" }}
              >
                <span className="material-symbols-outlined text-[20px]">
                  {isGeneratingAnnouncement ? "hourglass_top" : "mail"}
                </span>
                {isGeneratingAnnouncement
                  ? "Generating..."
                  : "Generate Announcement"}
              </button>
            </div>

            {/* ────── Meal Type Totals ────── */}
            <div className="mb-12">
              <h3 className="text-2xl font-black text-[var(--color-background-dark)] mb-6">
                Meal Participation
              </h3>

              {Object.keys(currentDay.meal_type_totals).length > 0 ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                  {Object.entries(currentDay.meal_type_totals).map(
                    ([mealType, counts]) => {
                      const total = counts.participating + counts.opted_out;
                      const rate =
                        total > 0
                          ? Math.round((counts.participating / total) * 100)
                          : 0;

                      return (
                        <div
                          key={mealType}
                          className="bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col transition-transform hover:-translate-y-1 duration-300"
                          style={{ boxShadow: "var(--shadow-clay)" }}
                        >
                          <h4 className="text-xl font-bold text-[var(--color-background-dark)] mb-6 capitalize">
                            {MEAL_TYPES[mealType as MealTypeEnum] || mealType}
                          </h4>

                          <div className="flex justify-between items-end mb-6">
                            <div className="text-center">
                              <p className="text-4xl font-black text-green-600">
                                {counts.participating}
                              </p>
                              <p className="text-sm text-[var(--color-text-sub)] font-medium mt-1">
                                Eating
                              </p>
                            </div>
                            <div className="text-center">
                              <p className="text-4xl font-black text-red-500">
                                {counts.opted_out}
                              </p>
                              <p className="text-sm text-[var(--color-text-sub)] font-medium mt-1">
                                Opted Out
                              </p>
                            </div>
                          </div>

                          <div className="w-full h-px bg-gradient-to-r from-transparent via-[var(--color-clay-shadow)] to-transparent mb-4" />

                          <div className="flex items-center justify-between">
                            <span className="text-sm font-bold text-[var(--color-text-sub)]">
                              Participation
                            </span>
                            <span className="text-sm font-black text-[var(--color-primary)]">
                              {rate}%
                            </span>
                          </div>

                          <div
                            className="w-full h-2 rounded-full mt-2"
                            style={{
                              backgroundColor: "var(--color-clay-shadow)",
                            }}
                          >
                            <div
                              className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                              style={{ width: `${rate}%` }}
                            />
                          </div>
                        </div>
                      );
                    },
                  )}
                </div>
              ) : (
                <div className="text-center py-12">
                  <span className="material-symbols-outlined text-6xl text-[var(--color-text-sub)] mb-4 block">
                    no_meals
                  </span>
                  <p className="text-[var(--color-text-sub)] text-lg">
                    No meal data available for this day.
                  </p>
                </div>
              )}
            </div>

            {/* ────── Team Breakdown ────── */}
            {currentDay.team_totals.length > 0 && (
              <div className="mb-12">
                <h3 className="text-2xl font-black text-[var(--color-background-dark)] mb-6">
                  Team Breakdown
                </h3>

                <div className="space-y-5">
                  {currentDay.team_totals.map((team) => {
                    const isExpanded = expandedTeams[team.team_id] ?? false;
                    const officeTotal =
                      team.office_wfh_split.office +
                      team.office_wfh_split.wfh;
                    const officePct =
                      officeTotal > 0
                        ? Math.round(
                            (team.office_wfh_split.office / officeTotal) * 100,
                          )
                        : 0;

                    return (
                      <div
                        key={team.team_id}
                        className="rounded-3xl bg-[var(--color-background-light)] border border-white/40 overflow-hidden transition-all duration-300"
                        style={{ boxShadow: "var(--shadow-clay)" }}
                      >
                        {/* Team Header (clickable) */}
                        <button
                          onClick={() => toggleTeam(team.team_id)}
                          className="w-full flex items-center justify-between px-7 py-5 text-left group transition-colors duration-200 hover:bg-white/30"
                        >
                          <div className="flex items-center gap-4">
                            <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-[var(--color-primary)]/10 group-hover:bg-[var(--color-primary)]/15 transition-colors">
                              <span className="material-symbols-outlined text-[var(--color-primary)] text-[22px]">
                                group
                              </span>
                            </div>
                            <div>
                              <h4 className="text-lg font-black text-[var(--color-background-dark)]">
                                {team.team_name}
                              </h4>
                              <p className="text-sm text-[var(--color-text-sub)] font-medium">
                                {team.total_members} member
                                {team.total_members !== 1 ? "s" : ""}
                              </p>
                            </div>
                          </div>

                          <div className="flex items-center gap-5">
                            {/* Mini office/wfh pills */}
                            <div className="hidden sm:flex items-center gap-2">
                              <span className="inline-flex items-center gap-1 text-xs font-bold bg-blue-50 text-blue-600 px-3 py-1.5 rounded-xl">
                                <span className="material-symbols-outlined text-[14px]">
                                  apartment
                                </span>
                                {team.office_wfh_split.office}
                              </span>
                              <span className="inline-flex items-center gap-1 text-xs font-bold bg-purple-50 text-purple-600 px-3 py-1.5 rounded-xl">
                                <span className="material-symbols-outlined text-[14px]">
                                  home
                                </span>
                                {team.office_wfh_split.wfh}
                              </span>
                            </div>

                            <span
                              className={`material-symbols-outlined text-[var(--color-text-sub)] text-[20px] transition-transform duration-300 ${
                                isExpanded ? "rotate-180" : ""
                              }`}
                            >
                              expand_more
                            </span>
                          </div>
                        </button>

                        {/* Expanded Content */}
                        {isExpanded && (
                          <div className="px-7 pb-6 pt-1 border-t border-black/[0.04]">
                            {/* Office / WFH Bar */}
                            <div className="mb-5">
                              <div className="flex items-center justify-between mb-2">
                                <span className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-sub)]">
                                  Office vs WFH
                                </span>
                                <span className="text-xs font-bold text-[var(--color-text-sub)]">
                                  {team.office_wfh_split.office} office ·{" "}
                                  {team.office_wfh_split.wfh} wfh
                                </span>
                              </div>
                              <div className="flex h-3 rounded-full overflow-hidden bg-[var(--color-clay-shadow)]">
                                <div
                                  className="bg-gradient-to-r from-blue-400 to-blue-500 transition-all duration-500"
                                  style={{ width: `${officePct}%` }}
                                />
                                <div
                                  className="bg-gradient-to-r from-purple-400 to-purple-500 transition-all duration-500"
                                  style={{
                                    width: `${100 - officePct}%`,
                                  }}
                                />
                              </div>
                            </div>

                            {/* Team Meal Totals */}
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                              {Object.entries(team.meal_type_totals).map(
                                ([mealType, counts]) => {
                                  const total =
                                    counts.participating + counts.opted_out;
                                  const rate =
                                    total > 0
                                      ? Math.round(
                                          (counts.participating / total) * 100,
                                        )
                                      : 0;

                                  return (
                                    <div
                                      key={mealType}
                                      className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-4"
                                      style={{
                                        boxShadow:
                                          "var(--shadow-clay-button)",
                                      }}
                                    >
                                      <div className="flex items-start justify-between mb-3">
                                        <p className="text-sm font-black text-[var(--color-background-dark)] capitalize">
                                          {MEAL_TYPES[
                                            mealType as MealTypeEnum
                                          ] || mealType}
                                        </p>
                                        <p className="text-xs font-semibold text-[var(--color-text-sub)]">
                                          {rate}%
                                        </p>
                                      </div>

                                      <div className="grid grid-cols-2 gap-3 mb-3">
                                        <div className="rounded-xl bg-green-50 px-3 py-2 text-center">
                                          <p className="text-lg font-black text-green-600">
                                            {counts.participating}
                                          </p>
                                          <p className="text-xs text-[var(--color-text-sub)]">
                                            Eating
                                          </p>
                                        </div>
                                        <div className="rounded-xl bg-red-50 px-3 py-2 text-center">
                                          <p className="text-lg font-black text-red-500">
                                            {counts.opted_out}
                                          </p>
                                          <p className="text-xs text-[var(--color-text-sub)]">
                                            Opted Out
                                          </p>
                                        </div>
                                      </div>

                                      <div
                                        className="h-2 w-full rounded-full"
                                        style={{
                                          backgroundColor:
                                            "var(--color-clay-shadow)",
                                        }}
                                      >
                                        <div
                                          className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                                          style={{ width: `${rate}%` }}
                                        />
                                      </div>
                                    </div>
                                  );
                                },
                              )}
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </>
        )}

        {/* No data fallback */}
        {!currentDay && !error && (
          <div className="text-center py-16">
            <span className="material-symbols-outlined text-7xl text-[var(--color-text-sub)] mb-4 block">
              query_stats
            </span>
            <p className="text-[var(--color-text-sub)] text-lg">
              No headcount report data available.
            </p>
          </div>
        )}

        {/* ────── Announcement Modal ────── */}
        {showAnnouncementModal && announceData && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
            <div
              className="absolute inset-0 bg-[var(--color-background-dark)]/45 backdrop-blur-sm"
              onClick={closeAnnouncementModal}
              aria-hidden="true"
            />

            <div
              role="dialog"
              aria-modal="true"
              aria-labelledby="announcement-modal-title"
              className="relative w-full max-w-3xl max-h-[92vh] overflow-hidden rounded-[2rem] border border-white/40 bg-[var(--color-background-light)] p-1"
              style={{ boxShadow: "var(--shadow-clay-modal)" }}
              onClick={(event) => event.stopPropagation()}
            >
              <div className="max-h-[calc(92vh-8px)] overflow-y-auto rounded-[1.75rem] bg-gradient-to-b from-white/70 to-[var(--color-background-light)]">
                {/* Modal Header */}
                <div className="sticky top-0 z-10 border-b border-black/[0.05] bg-[var(--color-background-light)]/95 px-5 py-5 backdrop-blur-sm sm:px-8 sm:py-6">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex items-start gap-3">
                      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--color-primary)]/15">
                        <span className="material-symbols-outlined text-[24px] text-[var(--color-primary)]">
                          campaign
                        </span>
                      </div>
                      <div>
                        <h3
                          id="announcement-modal-title"
                          className="text-2xl font-black text-[var(--color-background-dark)]"
                        >
                          Daily Announcement
                        </h3>
                        <p className="mt-1 text-sm font-medium text-[var(--color-text-sub)]">
                          {formatDate(announceData.date)}
                        </p>
                      </div>
                    </div>

                    <button
                      onClick={closeAnnouncementModal}
                      className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] transition-all duration-200 hover:text-[var(--color-primary)]"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                      aria-label="Close announcement modal"
                    >
                      <span className="material-symbols-outlined text-[20px]">
                        close
                      </span>
                    </button>
                  </div>
                </div>

                {/* Modal Content */}
                <div className="space-y-7 px-5 pb-6 pt-6 sm:px-8 sm:pb-8">
                  {/* Summary cards */}
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div
                      className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-5"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <p className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-sub)]">
                        Day Status
                      </p>
                      <p className="mt-2 text-lg font-black text-[var(--color-background-dark)]">
                        {DAY_STATUSES[
                          announceData.day_status as keyof typeof DAY_STATUSES
                        ] || announceData.day_status}
                      </p>
                    </div>

                    <div
                      className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-5"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <p className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-sub)]">
                        Total Active Users
                      </p>
                      <p className="mt-2 text-3xl font-black text-[var(--color-primary)] leading-none">
                        {announceData.total_active_users.toLocaleString()}
                      </p>
                    </div>
                  </div>

                  {announceData.special_day_note && (
                    <div
                      className="rounded-2xl border border-black/[0.05] bg-white/70 p-5"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-sub)]">
                        Special Note
                      </p>
                      <p className="text-sm leading-relaxed text-[var(--color-text-main)]">
                        {announceData.special_day_note}
                      </p>
                    </div>
                  )}

                  {/* Meal Totals */}
                  <div>
                    <h4 className="mb-4 text-lg font-black text-[var(--color-background-dark)]">
                      Meal Participation
                    </h4>
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      {announceData.meal_totals.map((meal) => {
                        const totalResponses =
                          meal.participating + meal.opted_out;
                        const participationRate =
                          totalResponses > 0
                            ? Math.round(
                                (meal.participating / totalResponses) * 100,
                              )
                            : 0;

                        return (
                          <div
                            key={meal.meal_type}
                            className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-4"
                            style={{
                              boxShadow: "var(--shadow-clay-button)",
                            }}
                          >
                            <div className="mb-3 flex items-start justify-between gap-3">
                              <p className="text-sm font-black text-[var(--color-background-dark)] capitalize">
                                {MEAL_TYPES[
                                  meal.meal_type as MealTypeEnum
                                ] || meal.meal_type}
                              </p>
                              <p className="text-xs font-semibold text-[var(--color-text-sub)]">
                                {participationRate}% participating
                              </p>
                            </div>

                            <div className="mb-3 grid grid-cols-2 gap-3">
                              <div className="rounded-xl bg-green-50 px-3 py-2 text-center">
                                <p className="text-lg font-black text-green-600">
                                  {meal.participating}
                                </p>
                                <p className="text-xs text-[var(--color-text-sub)]">
                                  Eating
                                </p>
                              </div>
                              <div className="rounded-xl bg-red-50 px-3 py-2 text-center">
                                <p className="text-lg font-black text-red-500">
                                  {meal.opted_out}
                                </p>
                                <p className="text-xs text-[var(--color-text-sub)]">
                                  Opted Out
                                </p>
                              </div>
                            </div>

                            <div
                              className="h-2 w-full rounded-full"
                              style={{
                                backgroundColor: "var(--color-clay-shadow)",
                              }}
                            >
                              <div
                                className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                                style={{
                                  width: `${participationRate}%`,
                                }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Copy-paste message */}
                  <div
                    className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-5 sm:p-6"
                    style={{ boxShadow: "var(--shadow-clay-button)" }}
                  >
                    <h4 className="mb-3 text-lg font-black text-[var(--color-background-dark)]">
                      Copy-Paste Message
                    </h4>
                    <p className="break-words whitespace-pre-wrap rounded-xl bg-white/60 p-4 font-mono text-sm leading-relaxed text-[var(--color-text-main)]">
                      {announceData.message}
                    </p>
                  </div>

                  {/* Footer buttons */}
                  <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                    <button
                      onClick={closeAnnouncementModal}
                      className="px-6 py-3 rounded-2xl bg-[var(--color-background-light)] text-[var(--color-background-dark)] font-bold transition-all duration-200 hover:text-[var(--color-primary)]"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      Close
                    </button>
                    <button
                      onClick={handleCopyToClipboard}
                      className="flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-[var(--color-primary)] text-white font-bold transition-all duration-200 hover:shadow-[6px_6px_12px_rgba(var(--color-primary-rgb),_0.3)]"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <span className="material-symbols-outlined text-[18px]">
                        content_copy
                      </span>
                      Copy to Clipboard
                    </button>
                  </div>
                </div>
              </div>
            </div>
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

/* ─── StatCard sub-component ─────────────────────────── */

interface StatCardProps {
  icon: string;
  label: string;
  value: number;
  color: string;
}

const StatCard: React.FC<StatCardProps> = ({ icon, label, value, color }) => (
  <div
    className="bg-gradient-to-br from-[var(--color-background-light)] to-white rounded-2xl px-6 py-5 border border-black/[0.04] transition-all duration-200 hover:shadow-[var(--shadow-clay-button-hover)] group"
    style={{ boxShadow: "var(--shadow-clay-button)" }}
  >
    <div
      className="flex items-center justify-center w-10 h-10 rounded-xl mb-3 transition-colors duration-200"
      style={{ backgroundColor: `${color}15` }}
    >
      <span
        className="material-symbols-outlined text-[22px] font-medium"
        style={{ color }}
      >
        {icon}
      </span>
    </div>
    <p className="text-3xl font-black text-[var(--color-background-dark)] leading-none tracking-tight">
      {value.toLocaleString()}
    </p>
    <p className="text-[12px] text-[var(--color-text-sub)] font-semibold tracking-wide mt-1.5">
      {label}
    </p>
  </div>
);
