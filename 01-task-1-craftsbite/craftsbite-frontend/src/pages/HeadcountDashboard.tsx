import React, { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import type { HeadcountData, MealType as MealTypeEnum } from "../types";
import { MEAL_TYPES } from "../utils/constants";
import * as headcountService from "../services/headcountService";
import { formatDate, getTodayString } from "../utils";
import { HeadcountDateModal, AnnouncementModal } from "../components/modals";

export const HeadcountDashboard: React.FC = () => {
  const { user } = useAuth();

  const [headcountData, setHeadcountData]     = useState<HeadcountData | null>(null);
  const [viewedDate, setViewedDate]           = useState<string>(getTodayString());
  const [isLoading, setIsLoading]             = useState(true);
  const [isFetching, setIsFetching]           = useState(false);
  const [error, setError]                     = useState("");
  const [expandedTeamId, setExpandedTeamId]   = useState<string | null>(null);
  const [isDateModalOpen, setIsDateModalOpen] = useState(false);

  const [announcement, setAnnouncement]               = useState<string>("");
  const [announcementDate, setAnnouncementDate]       = useState<string>("");
  const [isAnnounceModalOpen, setIsAnnounceModalOpen] = useState(false);
  const [isAnnouncePickerOpen, setIsAnnouncePickerOpen] = useState(false);
  const [isFetchingAnnouncement, setIsFetchingAnnouncement] = useState(false);
  const [announceError, setAnnounceError]             = useState("");

  const fetchByDate = async (date: string, isInitial = false) => {
    try {
      isInitial ? setIsLoading(true) : setIsFetching(true);
      setError("");
      const res = await headcountService.getHeadcountByDate(date);
      if (res.success && res.data) {
        setHeadcountData(res.data);
        setViewedDate(date);
        setExpandedTeamId(null);
        setIsDateModalOpen(false);
      } else {
        setError("No data returned for the selected date.");
      }
    } catch (err: any) {
      setError(err?.error?.message || "Failed to load headcount data.");
    } finally {
      isInitial ? setIsLoading(false) : setIsFetching(false);
    }
  };

  const fetchAnnouncement = async (date: string) => {
    try {
      setIsFetchingAnnouncement(true);
      setAnnounceError("");
      const res = await headcountService.getHeadcountAnnouncement(date);
      if (res.success && res.data) {
        setAnnouncement(res.data.message);
        setAnnouncementDate(date);
        setIsAnnouncePickerOpen(false);
        setIsAnnounceModalOpen(true);
      } else {
        setAnnounceError("No announcement data returned.");
      }
    } catch (err: any) {
      setAnnounceError(err?.error?.message || "Failed to generate announcement.");
    } finally {
      setIsFetchingAnnouncement(false);
    }
  };

  useEffect(() => {
    fetchByDate(getTodayString(), true);
  }, []);

  const getMealLabel = (mealType: string): string =>
    MEAL_TYPES[mealType as MealTypeEnum] || mealType;

  const displayLabel =
    viewedDate === getTodayString() ? "Today" : formatDate(viewedDate);

  if (isLoading) {
    return <LoadingSpinner message="Loading headcount data..." />;
  }

  const currentHeadcount = headcountData;

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name || "User"} userRole={user?.role || "admin"} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">

        {/* Page Title */}
        <div className="mb-10 flex flex-col md:flex-row md:items-end md:justify-between gap-4">
          <div>
            <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
              Headcount Dashboard
            </h2>
            <p className="text-lg text-[var(--color-text-sub)] font-medium">
              Meal participation summary ·{" "}
              <span className="font-bold text-[var(--color-primary)]">{displayLabel}</span>
            </p>
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-2 self-start md:self-auto flex-wrap">
            {/* Pick Date */}
            <button
              type="button"
              onClick={() => setIsDateModalOpen(true)}
              className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-xs font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
            >
              <span className="material-symbols-outlined text-[15px]">calendar_month</span>
              Pick Date
            </button>

            {/* Generate Announcement */}
            <button
              type="button"
              onClick={() => setIsAnnouncePickerOpen(true)}
              className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-xs font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
            >
              <span className="material-symbols-outlined text-[15px]">campaign</span>
              Generate Announcement
            </button>
          </div>
        </div>

        {/* Headcount error */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0">
            {error}
          </div>
        )}

        {/* Announcement error (inline, dismissible) */}
        {announceError && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0 flex items-center justify-between gap-4">
            <span>{announceError}</span>
            <button
              type="button"
              onClick={() => setAnnounceError("")}
              className="text-red-400 hover:text-red-600 transition-colors"
            >
              <span className="material-symbols-outlined text-[18px]">close</span>
            </button>
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
                {displayLabel}
              </p>
            </div>
          </div>
        )}

        {currentHeadcount?.location_split &&
          (() => {
            const loc = currentHeadcount.location_split;
            const total = currentHeadcount.total_active_users || 1;
            const officePercent = Math.round((loc.office / total) * 100);
            const wfhPercent    = Math.round((loc.wfh    / total) * 100);
            const notSetPercent = 100 - officePercent - wfhPercent;
            return (
              <div
                className="mb-8 bg-[var(--color-background-light)] rounded-3xl p-6 md:p-8"
                style={{ boxShadow: "var(--shadow-clay)" }}
              >
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-[var(--color-primary)]/10">
                    <span className="material-symbols-outlined text-[20px] text-[var(--color-primary)]">
                      location_on
                    </span>
                  </div>
                  <h3 className="text-lg font-black tracking-tight text-[var(--color-background-dark)]">
                    Work Location Overview
                  </h3>
                </div>

                <div className="flex flex-wrap gap-4 mb-5">
                  {[
                    { emoji: "🏢", label: "Office",  count: loc.office },
                    { emoji: "🏠", label: "WFH",     count: loc.wfh },
                    { emoji: "❓", label: "Not Set", count: loc.not_set },
                  ].map(({ emoji, label, count }) => (
                    <div
                      key={label}
                      className="flex items-center gap-2 bg-white rounded-2xl px-4 py-2.5 border border-[#e6dccf]"
                      style={{ boxShadow: "var(--shadow-clay-button)" }}
                    >
                      <span className="text-lg">{emoji}</span>
                      <div>
                        <p className="text-xl font-black text-[var(--color-background-dark)] leading-none">
                          {count}
                        </p>
                        <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mt-0.5">
                          {label}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>

                <div
                  className="w-full h-3 rounded-full flex overflow-hidden gap-0.5"
                  style={{ backgroundColor: "var(--color-clay-shadow)" }}
                >
                  {officePercent > 0 && (
                    <div
                      className="h-full rounded-l-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                      style={{ width: `${officePercent}%` }}
                    />
                  )}
                  {wfhPercent > 0 && (
                    <div
                      className="h-full bg-blue-400 transition-all duration-500"
                      style={{ width: `${wfhPercent}%` }}
                    />
                  )}
                  {notSetPercent > 0 && (
                    <div
                      className="h-full rounded-r-full bg-gray-300 transition-all duration-500"
                      style={{ width: `${notSetPercent}%` }}
                    />
                  )}
                </div>

                <div className="flex gap-4 mt-3 text-xs text-[var(--color-text-sub)] font-semibold flex-wrap">
                  <span>
                    <span className="inline-block w-2 h-2 rounded-full bg-[var(--color-primary)] mr-1" />
                    {officePercent}% Office
                  </span>
                  <span>
                    <span className="inline-block w-2 h-2 rounded-full bg-blue-400 mr-1" />
                    {wfhPercent}% WFH
                  </span>
                  <span>
                    <span className="inline-block w-2 h-2 rounded-full bg-gray-300 mr-1" />
                    {notSetPercent}% Not Set
                  </span>
                </div>
              </div>
            );
          })()}

        {/* Headcount Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
          {currentHeadcount?.meals &&
          Object.keys(currentHeadcount.meals).length > 0 ? (
            Object.entries(currentHeadcount.meals).map(([mealType, counts]) => {
              const total = counts.participating + counts.opted_out;
              const participationRate =
                total > 0 ? Math.round((counts.participating / total) * 100) : 0;
              return (
                <div
                  key={mealType}
                  className="bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col transition-transform hover:-translate-y-1 duration-300"
                  style={{ boxShadow: "var(--shadow-clay)" }}
                >
                  <h3 className="text-xl font-bold text-[var(--color-background-dark)] mb-6 capitalize">
                    {MEAL_TYPES[mealType as MealTypeEnum] || mealType}
                  </h3>
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
                      {participationRate}%
                    </span>
                  </div>
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
                No headcount data available for {displayLabel}.
              </p>
            </div>
          )}
        </div>

        {/* Team Breakdown */}
        {currentHeadcount?.teams && currentHeadcount.teams.length > 0 && (
          <div className="mb-12">
            <h3 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)] mb-5">
              Team Breakdown
            </h3>
            <div className="flex flex-col gap-4">
              {currentHeadcount.teams.map((team) => {
                const isExpanded = expandedTeamId === team.team_id;
                const teamTotal  = team.total_members || 1;
                const offPct = Math.round((team.location_split.office / teamTotal) * 100);
                const wfhPct = Math.round((team.location_split.wfh    / teamTotal) * 100);
                return (
                  <div
                    key={team.team_id}
                    className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] overflow-hidden"
                    style={{ boxShadow: "var(--shadow-clay-card)" }}
                  >
                    <button
                      type="button"
                      onClick={() =>
                        setExpandedTeamId(isExpanded ? null : team.team_id)
                      }
                      className="w-full flex items-center justify-between px-6 py-5 text-left hover:bg-[var(--color-background-light)] transition-colors"
                    >
                      <div>
                        <h4 className="text-lg font-black tracking-tight text-[var(--color-background-dark)]">
                          {team.team_name}
                        </h4>
                        <p className="text-sm text-[var(--color-text-sub)] mt-0.5">
                          {team.total_members} members · 🏢{" "}
                          {team.location_split.office} · 🏠{" "}
                          {team.location_split.wfh} · ❓{" "}
                          {team.location_split.not_set}
                        </p>
                      </div>
                      <span
                        className="material-symbols-outlined text-[20px] text-[var(--color-text-sub)] transition-transform duration-200"
                        style={{
                          transform: isExpanded
                            ? "rotate(180deg)"
                            : "rotate(0deg)",
                        }}
                      >
                        expand_more
                      </span>
                    </button>

                    {isExpanded && (
                      <div className="border-t border-[#e6dccf] px-6 py-5 flex flex-col gap-4">
                        <div>
                          <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-2">
                            Location Split
                          </p>
                          <div
                            className="w-full h-2 rounded-full flex overflow-hidden gap-0.5"
                            style={{ backgroundColor: "var(--color-clay-shadow)" }}
                          >
                            {offPct > 0 && (
                              <div
                                className="h-full rounded-l-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)]"
                                style={{ width: `${offPct}%` }}
                              />
                            )}
                            {wfhPct > 0 && (
                              <div
                                className="h-full bg-blue-400"
                                style={{ width: `${wfhPct}%` }}
                              />
                            )}
                            {100 - offPct - wfhPct > 0 && (
                              <div
                                className="h-full rounded-r-full bg-gray-300"
                                style={{ width: `${100 - offPct - wfhPct}%` }}
                              />
                            )}
                          </div>
                        </div>

                        {Object.entries(team.meals).map(([mealType, counts]) => {
                          const pct =
                            team.total_members > 0
                              ? Math.round(
                                  (counts.participating / team.total_members) * 100
                                )
                              : 0;
                          return (
                            <div key={mealType}>
                              <div className="flex items-center justify-between mb-1.5">
                                <p className="text-sm font-semibold text-[var(--color-text-main)] capitalize">
                                  {getMealLabel(mealType)}
                                </p>
                                <p className="text-sm font-black text-[var(--color-primary)]">
                                  {counts.participating}/{team.total_members}
                                </p>
                              </div>
                              <div
                                className="w-full h-2 rounded-full"
                                style={{ backgroundColor: "var(--color-clay-shadow)" }}
                              >
                                <div
                                  className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                                  style={{ width: `${pct}%` }}
                                />
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </main>

      <Footer
        links={[
          { label: "Privacy", href: "#" },
          { label: "Terms",   href: "#" },
          { label: "Support", href: "#" },
        ]}
      />

      <HeadcountDateModal
        isOpen={isDateModalOpen}
        title="View Headcount by Date"
        submitLabel="Fetch"
        submitIcon="search"
        isSubmitting={isFetching}
        onClose={() => setIsDateModalOpen(false)}
        onSubmit={(date) => fetchByDate(date)}
      />

      <HeadcountDateModal
        isOpen={isAnnouncePickerOpen}
        title="Generate Announcement"
        submitLabel="Generate"
        submitIcon="campaign"
        isSubmitting={isFetchingAnnouncement}
        onClose={() => {
          setIsAnnouncePickerOpen(false);
          setAnnounceError("");
        }}
        onSubmit={(date) => fetchAnnouncement(date)}
      />

      <AnnouncementModal
        isOpen={isAnnounceModalOpen}
        date={announcementDate}
        message={announcement}
        onClose={() => setIsAnnounceModalOpen(false)}
      />
    </div>
  );
};
