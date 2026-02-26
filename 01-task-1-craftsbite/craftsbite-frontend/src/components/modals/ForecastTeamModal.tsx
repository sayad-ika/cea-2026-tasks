import React from "react";
import type { HeadcountData } from "../../types";
import { MEAL_TYPES } from "../../utils/constants";
import { formatDayLabel } from "../cards/ForecastDayCard";
import { formatDate } from "../../utils/dateUtils";

export interface ForecastTeamModalProps {
    day: HeadcountData | null;
    onClose: () => void;
}

export const ForecastTeamModal: React.FC<ForecastTeamModalProps> = ({
    day,
    onClose,
}) => {
    if (!day) return null;

    const teams = day.teams ?? [];

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-sm"
            onClick={onClose}
        >
            <div
                className="bg-[var(--color-clay-light)] w-full max-w-md rounded-3xl p-6 m-4"
                style={{ boxShadow: "var(--shadow-clay-modal)" }}
                onClick={(e) => e.stopPropagation()}
            >
                {/* ── header ── */}
                <div className="flex justify-between items-center mb-6">
                    <div>
                        <h3 className="text-xl font-bold text-[var(--color-background-dark)]">
                            {formatDayLabel(day.date)} Breakdown
                        </h3>
                        <p className="text-sm text-[var(--color-text-sub)]">
                            {formatDate(day.date, "MMM d, yyyy")}
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="w-8 h-8 rounded-full bg-[var(--color-background-light)] flex items-center justify-center text-[var(--color-text-sub)] hover:bg-[var(--color-clay-shadow)] transition-colors"
                    >
                        <span className="material-symbols-outlined text-sm">
                            close
                        </span>
                    </button>
                </div>

                {/* ── teams list ── */}
                {teams.length === 0 ? (
                    <p className="text-center py-8 text-[var(--color-text-sub)] text-sm">
                        No team data available for this day.
                    </p>
                ) : (
                    <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-1">
                        {teams.map((team) => {
                            const mealEntries = Object.entries(
                                team.meals ?? {},
                            );
                            const barPercent =
                                day.total_active_users > 0
                                    ? Math.round(
                                          (team.total_members /
                                              day.total_active_users) *
                                              100,
                                      )
                                    : 0;

                            return (
                                <div
                                    key={team.team_id}
                                    className="p-4 rounded-xl bg-[var(--color-background-light)] border border-[var(--color-clay-shadow)]"
                                >
                                    {/* team name row */}
                                    <div className="flex justify-between items-center mb-2">
                                        <div className="flex items-center gap-2">
                                            <span className="w-2 h-2 rounded-full bg-[var(--color-primary)]" />
                                            <span className="font-bold text-[var(--color-background-dark)]">
                                                {team.team_name}
                                            </span>
                                        </div>
                                        <span className="text-sm font-semibold text-[var(--color-background-dark)]">
                                            {team.total_members} Members
                                        </span>
                                    </div>

                                    {/* progress bar */}
                                    <div className="w-full bg-[var(--color-clay-shadow)] h-2 rounded-full overflow-hidden">
                                        <div
                                            className="bg-[var(--color-primary)] h-full rounded-full transition-all duration-500"
                                            style={{ width: `${barPercent}%` }}
                                        />
                                    </div>

                                    {/* location split */}
                                    <div className="flex justify-between mt-2 text-xs text-[var(--color-text-sub)]">
                                        <span>
                                            {team.location_split?.office ?? 0}{" "}
                                            Office
                                        </span>
                                        <span>
                                            {team.location_split?.wfh ?? 0} WFH
                                        </span>
                                        <span>
                                            {team.location_split?.not_set ?? 0}{" "}
                                            Not Set
                                        </span>
                                    </div>

                                    {/* meal breakdown */}
                                    {mealEntries.length > 0 && (
                                        <div className="mt-3 pt-3 border-t border-[var(--color-clay-shadow)] flex flex-wrap gap-2">
                                            {mealEntries.map(
                                                ([mealKey, summary]) => (
                                                    <span
                                                        key={mealKey}
                                                        className="text-[11px] font-medium bg-white dark:bg-[#2C2A28] px-2 py-0.5 rounded-md border border-orange-100 dark:border-gray-700 text-gray-600 dark:text-gray-300"
                                                    >
                                                        {MEAL_TYPES[
                                                            mealKey as keyof typeof MEAL_TYPES
                                                        ] ?? mealKey}
                                                        :{" "}
                                                        <span className="font-bold text-gray-900 dark:text-white">
                                                            {
                                                                summary.participating
                                                            }
                                                        </span>
                                                    </span>
                                                ),
                                            )}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                )}

                {/* ── summary row ── */}
                <div className="mt-6 pt-4 border-t border-[var(--color-clay-shadow)] flex items-center justify-between text-sm text-[var(--color-text-sub)]">
                    <span>Total active users</span>
                    <span className="font-bold text-[var(--color-background-dark)]">
                        {day.total_active_users}
                    </span>
                </div>
            </div>
        </div>
    );
};
