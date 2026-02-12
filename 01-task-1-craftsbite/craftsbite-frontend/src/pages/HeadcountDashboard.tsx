import React, { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import type { HeadcountData, MealType as MealTypeEnum } from "../types";
import { MEAL_TYPES } from "../utils/constants";
import * as headcountService from "../services/headcountService";
import type { DailyAnnouncementResponse } from "../types";
import toast from "react-hot-toast";

export const HeadcountDashboard: React.FC = () => {
  const { user } = useAuth();
  const [headcountData, setHeadcountData] = useState<HeadcountData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedDateIndex, setSelectedDateIndex] = useState(0);
  const [announceData, setAnnounceData] =
    useState<DailyAnnouncementResponse | null>(null);
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
  const [isGeneratingAnnouncement, setIsGeneratingAnnouncement] =
    useState(false);

  useEffect(() => {
    const fetchHeadcount = async () => {
      try {
        setIsLoading(true);
        const res = await headcountService.getTodayHeadcount();
        if (res.success && res.data) {
          setHeadcountData(res.data);
          setSelectedDateIndex(0);
        }
      } catch (err: any) {
        setError(err?.error?.message || "Failed to load headcount data.");
      } finally {
        setIsLoading(false);
      }
    };
    fetchHeadcount();
  }, []);

  useEffect(() => {
    if (!showAnnouncementModal) return;

    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setShowAnnouncementModal(false);
      }
    };

    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleEscape);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleEscape);
    };
  }, [showAnnouncementModal]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      weekday: "long",
      month: "short",
      day: "numeric",
    });
  };

  const getDateLabel = (dateStr: string, index: number) => {
    const date = new Date(dateStr);
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const isToday = date.getTime() === today.getTime();

    return isToday ? "Today" : index === 1 ? "Tomorrow" : formatDate(dateStr);
  };

  const formatDayStatus = (status: string) =>
    status.replace(/[_-]/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());

  const closeAnnouncementModal = () => setShowAnnouncementModal(false);

  const handleGenerateAnnouncement = async () => {
    if (!currentHeadcount) return;

    setIsGeneratingAnnouncement(true);
    try {
      const res = await headcountService.generateDailyAnnouncement(
        currentHeadcount.date,
      );
      if (res.success && res.data) {
        setAnnounceData(res.data);
        setShowAnnouncementModal(true);
        toast.success("Announcement generated successfully!");
      } else {
        toast.error(res.message || "Failed to generate announcement");
      }
    } catch (err: any) {
      const errorMsg = err?.error?.message || "Failed to generate announcement";
      toast.error(errorMsg);
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

  const currentHeadcount = headcountData[selectedDateIndex];

  if (isLoading) {
    return <LoadingSpinner message="Loading headcount data..." />;
  }

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
            Meal participation summary
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0">
            {error}
          </div>
        )}

        {/* Date Selector Tabs */}
        {headcountData.length > 0 && (
          <div className="mb-8 flex gap-3 flex-wrap">
            {headcountData.map((_, index) => {
              const dateStr = headcountData[index].date;
              return (
                <button
                  key={index}
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
                  {getDateLabel(dateStr, index)}
                </button>
              );
            })}
          </div>
        )}

        {/* Total Active Users Badge */}
        {currentHeadcount && (
          <div
            className="mb-8 inline-flex items-center gap-4 bg-gradient-to-br from-[var(--color-background-light)] to-white rounded-2xl px-7 py-5 self-start border border-black/[0.04] transition-all duration-200 hover:shadow-[var(--shadow-clay-button-hover)] group"
            style={{ boxShadow: "var(--shadow-clay-button)" }}
          >
            <div className="flex items-center justify-center w-12 h-12 bg-[var(--color-primary)]/10 rounded-xl group-hover:bg-[var(--color-primary)]/15 transition-colors duration-200">
              <span className="material-symbols-outlined text-[var(--color-primary)] text-[26px] font-medium">
                groups
              </span>
            </div>
            <div className="flex flex-col gap-0.5">
              <p className="text-[28px] font-black text-[var(--color-background-dark)] leading-none tracking-tight">
                {currentHeadcount.total_active_users.toLocaleString()}
              </p>
              <p className="text-[13px] text-[var(--color-text-sub)] font-semibold tracking-wide mt-1">
                Active Users
              </p>
              <p className="text-[11px] text-[var(--color-text-sub)]/70 font-medium">
                {getDateLabel(currentHeadcount.date, selectedDateIndex)}
              </p>
            </div>
          </div>
        )}

        {/* Generate Announcement Button */}
        {currentHeadcount && (
          <div className="mb-8">
            <button
              onClick={handleGenerateAnnouncement}
              disabled={isGeneratingAnnouncement}
              className="flex items-center gap-3 px-6 py-3 bg-[var(--color-primary)] text-white font-bold rounded-2xl transition-all duration-200 hover:shadow-[6px_6px_12px_rgba(var(--color-primary-rgb),_0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                boxShadow: "var(--shadow-clay-button)",
              }}
            >
              <span className="material-symbols-outlined text-[20px]">
                {isGeneratingAnnouncement ? "hourglass_top" : "mail"}
              </span>
              {isGeneratingAnnouncement
                ? "Generating..."
                : "Generate Announcement"}
            </button>
          </div>
        )}

        {/* Headcount Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
          {currentHeadcount?.meals &&
          Object.keys(currentHeadcount.meals).length > 0 ? (
            Object.entries(currentHeadcount.meals).map(([mealType, counts]) => {
              const total = counts.participating + counts.opted_out;
              const participationRate =
                total > 0
                  ? Math.round((counts.participating / total) * 100)
                  : 0;

              return (
                <div
                  key={mealType}
                  className="bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col transition-transform hover:-translate-y-1 duration-300"
                  style={{ boxShadow: "var(--shadow-clay)" }}
                >
                  {/* Meal Name */}
                  <h3 className="text-xl font-bold text-[var(--color-background-dark)] mb-6 capitalize">
                    {MEAL_TYPES[mealType as MealTypeEnum] || mealType}
                  </h3>

                  {/* Counts */}
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

                  {/* Divider */}
                  <div className="w-full h-px bg-gradient-to-r from-transparent via-[var(--color-clay-shadow)] to-transparent mb-4" />

                  {/* Participation Rate */}
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-bold text-[var(--color-text-sub)]">
                      Participation
                    </span>
                    <span className="text-sm font-black text-[var(--color-primary)]">
                      {participationRate}%
                    </span>
                  </div>

                  {/* Progress Bar */}
                  <div
                    className="w-full h-2 rounded-full mt-2"
                    style={{ backgroundColor: "var(--color-clay-shadow)" }}
                  >
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                      style={{ width: `${participationRate}%` }}
                    />
                  </div>
                </div>
              );
            })
          ) : (
            <div className="col-span-full text-center py-12">
              <span className="material-symbols-outlined text-6xl text-[var(--color-text-sub)] mb-4 block">
                no_meals
              </span>
              <p className="text-[var(--color-text-sub)] text-lg">
                No headcount data available for{" "}
                {getDateLabel(currentHeadcount?.date || "", selectedDateIndex)}.
              </p>
            </div>
          )}
        </div>

        {/* Announcement Modal */}
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
                {/* Header */}
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

                {/* Content */}
                <div className="space-y-7 px-5 pb-6 pt-6 sm:px-8 sm:pb-8">
                  {/* Summary */}
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div
                      className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-5"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <p className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-sub)]">
                        Day Status
                      </p>
                      <p className="mt-2 text-lg font-black text-[var(--color-background-dark)]">
                        {formatDayStatus(announceData.day_status)}
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
                        const totalResponses = meal.participating + meal.opted_out;
                        const participationRate =
                          totalResponses > 0
                            ? Math.round((meal.participating / totalResponses) * 100)
                            : 0;

                        return (
                          <div
                            key={meal.meal_type}
                            className="rounded-2xl border border-black/[0.05] bg-[var(--color-background-light)] p-4"
                            style={{ boxShadow: "var(--shadow-clay-button)" }}
                          >
                            <div className="mb-3 flex items-start justify-between gap-3">
                              <p className="text-sm font-black text-[var(--color-background-dark)] capitalize">
                                {MEAL_TYPES[meal.meal_type as MealTypeEnum] ||
                                  meal.meal_type}
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
                              style={{ backgroundColor: "var(--color-clay-shadow)" }}
                            >
                              <div
                                className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                                style={{ width: `${participationRate}%` }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Message Section */}
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

                  {/* Footer */}
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
