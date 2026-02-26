import React, { useEffect, useState } from "react";
import type { HeadcountData, DayStatus } from "../types";
import { getForecast } from "../services/headcountService";
import {
    ForecastDayCard,
    STATUS_BADGE,
} from "../components/cards/ForecastDayCard";
import { ForecastTeamModal } from "../components/modals/ForecastTeamModal";

interface FilterChipProps {
    status: DayStatus;
    active: boolean;
    onClick: () => void;
}

const FilterChip: React.FC<FilterChipProps> = ({ status, active, onClick }) => {
    const badge = STATUS_BADGE[status];
    return (
        <button
            onClick={onClick}
            className={`
        px-4 py-2.5 rounded-xl text-sm font-medium
        border transition-colors flex items-center gap-2
        ${
            active
                ? "bg-[var(--color-primary)] border-[var(--color-primary)] text-white"
                : "bg-[var(--color-clay-light)] border-[var(--color-clay-shadow)] text-[var(--color-text-main)] hover:border-[var(--color-primary)]"
        }
      `}
        >
            <span
                className={`w-2 h-2 rounded-full ${active ? "bg-white" : badge.dotClass}`}
            />
            {badge.label}
        </button>
    );
};

// ─── main page ───────────────────────────────────────────────────────────────

const FILTERABLE_STATUSES: DayStatus[] = ["normal", "event_day", "celebration"];

export const ForecastDashboard: React.FC = () => {
    const [days, setDays] = useState<HeadcountData[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState("");
    const [activeFilters, setActiveFilters] = useState<Set<DayStatus>>(
        new Set(),
    );
    const [modalDay, setModalDay] = useState<HeadcountData | null>(null);

    useEffect(() => {
        const load = async () => {
            try {
                setIsLoading(true);
                setError("");
                const res = await getForecast(7);
                if (res.success && res.data) {
                    setDays(res.data);
                } else {
                    setError("No forecast data found.");
                }
            } catch (err: any) {
                setError(
                    err?.error?.message ?? "Failed to load forecast data.",
                );
            } finally {
                setIsLoading(false);
            }
        };
        load();
    }, []);

    const toggleFilter = (status: DayStatus) => {
        setActiveFilters((prev) => {
            const next = new Set(prev);
            next.has(status) ? next.delete(status) : next.add(status);
            return next;
        });
    };

    const visibleDays =
        activeFilters.size === 0
            ? days
            : days.filter((d) => activeFilters.has(d.day_status));

    return (
        <div>
            {/* ── page title ── */}
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
                <div>
                    <h1 className="text-4xl text-left font-extrabold text-[var(--color-background-dark)] tracking-tight mb-2">
                        Meal Forecast
                    </h1>
                    <p className="text-[var(--color-text-sub)]">
                        Overview of upcoming meal requirements across teams.
                    </p>
                </div>
            </div>

            {/* ── filter chips ── */}
            <div className="flex flex-wrap gap-3 mb-8">
                {FILTERABLE_STATUSES.map((status) => (
                    <FilterChip
                        key={status}
                        status={status}
                        active={activeFilters.has(status)}
                        onClick={() => toggleFilter(status)}
                    />
                ))}
                {activeFilters.size > 0 && (
                    <button
                        onClick={() => setActiveFilters(new Set())}
                        className="px-4 py-2.5 rounded-xl text-sm font-medium border border-dashed border-[var(--color-clay-shadow-dark)] text-[var(--color-text-sub)] hover:border-[var(--color-danger)] hover:text-[var(--color-danger)] transition-colors"
                    >
                        Clear filters
                    </button>
                )}
            </div>

            {/* ── states ── */}
            {isLoading ? (
                <div className="flex flex-col items-center justify-center py-24 gap-4 text-[var(--color-text-sub)]">
                    <div className="w-10 h-10 border-4 border-[var(--color-primary)] border-t-transparent rounded-full animate-spin" />
                    <p className="text-sm font-medium">Loading forecast…</p>
                </div>
            ) : error ? (
                <div className="flex flex-col items-center justify-center py-24 gap-3 text-red-400">
                    <span className="material-symbols-outlined text-4xl">
                        error
                    </span>
                    <p className="text-sm font-medium">{error}</p>
                    <button
                        onClick={() => window.location.reload()}
                        className="mt-2 px-4 py-2 rounded-xl bg-red-50 dark:bg-red-900/20 text-red-500 text-sm font-semibold border border-red-100 dark:border-red-800 hover:bg-red-100 transition-colors"
                    >
                        Retry
                    </button>
                </div>
            ) : visibleDays.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-24 gap-3 text-[var(--color-text-sub)]">
                    <span className="material-symbols-outlined text-4xl">
                        calendar_today
                    </span>
                    <p className="text-sm font-medium">
                        No days match the selected filters.
                    </p>
                </div>
            ) : (
                /* ── grid ── */
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {visibleDays.map((day) => (
                        <ForecastDayCard
                            key={day.date}
                            day={day}
                            onViewTeams={setModalDay}
                        />
                    ))}
                </div>
            )}

            {/* ── team modal ── */}
            {modalDay && (
                <ForecastTeamModal
                    day={modalDay}
                    onClose={() => setModalDay(null)}
                />
            )}
        </div>
    );
};
