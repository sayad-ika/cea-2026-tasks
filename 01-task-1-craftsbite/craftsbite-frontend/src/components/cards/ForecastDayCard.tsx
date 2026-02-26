import React from "react";
import { parseISO, isTomorrow } from "date-fns";
import type { HeadcountData, TeamHeadcount, DayStatus } from "../../types";
import { DAY_STATUSES, MEAL_TYPES } from "../../utils/constants";
import { isToday, formatDate } from "../../utils/dateUtils";

export const CLOSED_STATUSES: DayStatus[] = [
    "weekend",
    "office_closed",
    "govt_holiday",
];

export function formatDayLabel(dateStr: string): string {
    if (isToday(dateStr)) return "Today";
    if (isTomorrow(parseISO(dateStr))) return "Tomorrow";
    return formatDate(dateStr, "MMM d");
}

export function formatDayShort(dateStr: string): string {
    return formatDate(dateStr, "MMM d, EEE");
}

export function totalMealCount(day: HeadcountData): number {
    return Object.values(day.meals ?? {}).reduce(
        (sum, m) => sum + m.participating,
        0,
    );
}

export function ringPercent(day: HeadcountData): number {
    if (!day.total_active_users) return 0;
    return Math.min(
        Math.round((totalMealCount(day) / day.total_active_users) * 100),
        100,
    );
}

export const STATUS_BADGE: Record<
    DayStatus,
    { label: string; className: string; dotClass: string }
> = {
    normal: {
        label: "Normal",
        className:
            "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300 border border-green-300 dark:border-green-800",
        dotClass: "bg-green-500",
    },
    event_day: {
        label: "Event Day",
        className:
            "bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-300 border border-blue-300 dark:border-blue-800",
        dotClass: "bg-blue-500",
    },
    celebration: {
        label: "Celebration",
        className:
            "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300 border border-purple-300 dark:border-purple-800",
        dotClass: "bg-purple-500",
    },
    weekend: {
        label: "Weekend",
        className:
            "bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-300 border border-gray-400 dark:border-gray-600",
        dotClass: "bg-gray-500",
    },
    office_closed: {
        label: "Office Closed",
        className:
            "bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-300 border border-red-300 dark:border-red-800",
        dotClass: "bg-red-500",
    },
    govt_holiday: {
        label: "Holiday",
        className:
            "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300 border border-amber-300 dark:border-amber-800",
        dotClass: "bg-amber-500",
    },
};

export const RING_COLOR: Record<DayStatus, string> = {
    normal: "text-[var(--color-primary)]",
    event_day: "text-blue-400",
    celebration: "text-purple-400",
    weekend: "text-[var(--color-clay-shadow-dark)]",
    office_closed: "text-red-400",
    govt_holiday: "text-amber-400",
};

export const MEAL_DOT_COLORS = [
    "bg-[var(--color-primary)]",
    "bg-blue-400",
    "bg-purple-400",
    "bg-emerald-400",
    "bg-rose-400",
];

// ─── internal sub-components ──────────────────────────────────────────────────

interface DonutRingProps {
    percent: number;
    mealCount: number;
    status: DayStatus;
}

const DonutRing: React.FC<DonutRingProps> = ({
    percent,
    mealCount,
    status,
}) => {
    const colorClass = RING_COLOR[status] ?? "text-[#EF8543]";
    const dashArray = `${Math.max(percent, 0)}, 100`;

    return (
        <div className="relative w-24 h-24 flex-shrink-0">
            <svg
                className="w-full h-full transform -rotate-90"
                viewBox="0 0 36 36"
            >
                {/* track */}
                <path
                    className="text-[var(--color-clay-shadow)]"
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="3.8"
                    strokeLinecap="round"
                />
                {/* fill */}
                <path
                    className={`${colorClass} drop-shadow-sm`}
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="currentColor"
                    strokeDasharray={dashArray}
                    strokeDashoffset="0"
                    strokeWidth="3.8"
                    strokeLinecap="round"
                />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
                <span className="text-xl font-bold text-[var(--color-background-dark)] leading-none">
                    {mealCount}
                </span>
                <span className="text-[10px] text-[var(--color-text-sub)] font-medium mt-0.5">
                    Meals
                </span>
            </div>
        </div>
    );
};

interface TeamAvatarsProps {
    teams: TeamHeadcount[];
}

const TeamAvatars: React.FC<TeamAvatarsProps> = ({ teams }) => {
    const visible = teams.slice(0, 2);
    const extra = teams.length - visible.length;

    return (
        <span className="flex -space-x-2">
            {visible.map((_, i) => (
                <div
                    key={i}
                    className="w-6 h-6 rounded-full bg-[var(--color-clay-shadow)] border-2 border-[var(--color-clay-light)]"
                />
            ))}
            {extra > 0 && (
                <div className="w-6 h-6 rounded-full bg-[var(--color-text-sub)] border-2 border-[var(--color-clay-light)] text-[8px] flex items-center justify-center text-white font-bold">
                    +{extra}
                </div>
            )}
        </span>
    );
};

// ─── public component ─────────────────────────────────────────────────────────

export interface ForecastDayCardProps {
    day: HeadcountData;
    onViewTeams: (day: HeadcountData) => void;
}

export const ForecastDayCard: React.FC<ForecastDayCardProps> = ({
    day,
    onViewTeams,
}) => {
    const badge = STATUS_BADGE[day.day_status] ?? STATUS_BADGE.normal;
    const isClosed = CLOSED_STATUSES.includes(day.day_status);
    const percent = ringPercent(day);
    const mealCount = totalMealCount(day);
    const mealEntries = Object.entries(day.meals ?? {});
    const teams = day.teams ?? [];

    return (
        <div
            className={`
                bg-[var(--color-clay-light)] rounded-3xl p-6
                shadow-[var(--shadow-clay-md)]
                hover:shadow-[var(--shadow-clay)]
                transition-all duration-300
                border border-[var(--color-clay-shadow)]
                flex flex-col
                ${isClosed ? "opacity-75" : ""}
            `}
        >
            {/* ── header ── */}
            <div className="flex justify-between items-start mb-6">
                <div>
                    <div className="text-xs font-bold text-[var(--color-text-sub)] uppercase tracking-wider mb-1">
                        {formatDayShort(day.date)}
                    </div>
                    <div className="text-xl font-bold text-[var(--color-background-dark)]">
                        {formatDayLabel(day.date)}
                    </div>
                </div>
                <span
                    className={`px-3 py-1 rounded-lg text-xs font-bold ${badge.className}`}
                >
                    {badge.label}
                </span>
            </div>

            {/* ── body ── */}
            {isClosed ? (
                <div className="flex flex-col items-center justify-center py-8 text-center flex-grow">
                    <div className="w-16 h-16 bg-[var(--color-background-light)] rounded-full flex items-center justify-center text-[var(--color-clay-shadow-dark)] mb-4">
                        <span className="material-symbols-outlined text-3xl">
                            {day.day_status === "weekend" ? "weekend" : "block"}
                        </span>
                    </div>
                    <h3 className="font-bold text-[var(--color-text-sub)] mb-1">
                        No Meals Scheduled
                    </h3>
                    <p className="text-xs text-[var(--color-text-sub)]">
                        {day.day_status === "weekend"
                            ? "Office kitchen closed"
                            : DAY_STATUSES[day.day_status]}
                    </p>
                </div>
            ) : (
                <div className="flex items-center gap-6 mb-6">
                    <DonutRing
                        percent={percent}
                        mealCount={mealCount}
                        status={day.day_status}
                    />

                    <div className="flex flex-col gap-2 w-full">
                        {/* location split */}
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center gap-2">
                                <span className="w-2 h-2 rounded-full bg-[var(--color-primary)]" />
                                <span className="text-[var(--color-text-main)] font-medium">
                                    Office
                                </span>
                            </div>
                            <span className="font-bold text-[var(--color-background-dark)]">
                                {day.location_split?.office ?? 0}
                            </span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center gap-2">
                                <span className="w-2 h-2 rounded-full bg-[var(--color-clay-shadow-dark)]" />
                                <span className="text-[var(--color-text-main)] font-medium">
                                    WFH
                                </span>
                            </div>
                            <span className="font-bold text-[var(--color-background-dark)]">
                                {day.location_split?.wfh ?? 0}
                            </span>
                        </div>

                        {/* meal breakdown */}
                        {mealEntries.length > 0 && (
                            <>
                                <div className="border-t border-[var(--color-clay-shadow)] my-1" />
                                {mealEntries.map(([mealKey, summary], idx) => (
                                    <div
                                        key={mealKey}
                                        className="flex items-center justify-between text-sm"
                                    >
                                        <div className="flex items-center gap-2">
                                            <span
                                                className={`w-2 h-2 rounded-full ${MEAL_DOT_COLORS[idx % MEAL_DOT_COLORS.length]}`}
                                            />
                                            <span className="text-[var(--color-text-main)] font-medium capitalize">
                                                {MEAL_TYPES[
                                                    mealKey as keyof typeof MEAL_TYPES
                                                ] ?? mealKey}
                                            </span>
                                        </div>
                                        <span className="font-bold text-[var(--color-background-dark)]">
                                            {summary.participating}
                                        </span>
                                    </div>
                                ))}
                            </>
                        )}
                    </div>
                </div>
            )}

            {/* ── footer ── */}
            <div
                className={`mt-auto pt-4 border-t border-[var(--color-clay-shadow)] ${isClosed ? "opacity-50 pointer-events-none" : ""}`}
            >
                <div className="flex justify-between items-center mb-4 text-xs text-[var(--color-text-sub)]">
                    {teams.length > 0 ? (
                        <>
                            <span>
                                {teams.length} Team
                                {teams.length !== 1 ? "s" : ""} Participating
                            </span>
                            <TeamAvatars teams={teams} />
                        </>
                    ) : (
                        <span>No team data</span>
                    )}
                </div>

                <button
                    onClick={() => onViewTeams(day)}
                    className="w-full py-2.5 rounded-xl bg-[var(--color-background-light)] text-[var(--color-primary)] font-semibold text-sm hover:bg-[var(--color-clay-shadow)] transition-colors flex items-center justify-center gap-2"
                    style={{ boxShadow: "var(--shadow-clay-button)" }}
                >
                    View Team Details
                    <span className="material-symbols-outlined text-base">
                        arrow_forward
                    </span>
                </button>
            </div>
        </div>
    );
};
