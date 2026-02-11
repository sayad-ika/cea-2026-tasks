import React, { useState, useEffect, useMemo } from "react";
import {
    format,
    startOfMonth,
    endOfMonth,
    addMonths,
    subMonths,
    getDay,
    getDaysInMonth,
    isToday,
    isWeekend,
} from "date-fns";
import { Header, Footer, Navbar, LoadingSpinner, CreateScheduleModal } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as scheduleService from "../services/scheduleService";
import type { ScheduleEntry, DayStatus } from "../types";
import toast from "react-hot-toast";

/* ------------------------------------------------------------------ */
/*  DayCell – a single day in the calendar grid                        */
/* ------------------------------------------------------------------ */
interface DayCellProps {
    dayNumber: number;
    date: Date;
    schedule?: ScheduleEntry;
    isCurrentMonth: boolean;
}

const DayCell: React.FC<DayCellProps> = ({ dayNumber, date, schedule, isCurrentMonth }) => {
    const today = isToday(date);
    const weekend = isWeekend(date);

    // Determine day status – API data takes precedence, otherwise derive from weekend
    const status: DayStatus | "weekend" = schedule?.day_status ?? (weekend ? "weekend" as any : "normal");

    // Base wrapper classes
    const baseClasses =
        "group p-3 rounded-2xl transition-all flex flex-col items-start justify-start min-h-[90px] md:min-h-[110px] relative select-none";

    if (!isCurrentMonth) {
        return <div className="p-2 rounded-xl bg-transparent opacity-20" />;
    }

    /* ---------- office_closed / weekend ---------- */
    if (status === "office_closed" || status === "weekend") {
        return (
            <div
                className={`${baseClasses} bg-gray-100 shadow-inner opacity-80 border border-gray-200 cursor-not-allowed`}
            >
                <span className="text-sm font-bold text-gray-400">{dayNumber}</span>
                <div className="flex flex-col items-center w-full mt-2">
                    <span className="material-symbols-outlined text-gray-300 text-[24px]">lock</span>
                    <span className="text-[9px] uppercase font-bold text-gray-400 mt-1 tracking-wider">
                        {status === "office_closed" ? "Closed" : "Weekend"}
                    </span>
                </div>
                {schedule?.reason && (
                    <span className="text-[8px] text-gray-400 mt-auto truncate w-full text-center" title={schedule.reason}>
                        {schedule.reason}
                    </span>
                )}
            </div>
        );
    }

    /* ---------- govt_holiday ---------- */
    if (status === "govt_holiday") {
        return (
            <div
                className={`${baseClasses} bg-red-50/50 shadow-[var(--shadow-clay-button)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] border border-red-100`}
            >
                <span className="text-sm font-bold text-red-400">{dayNumber}</span>
                <div className="flex flex-col items-center w-full mt-2 opacity-60">
                    <span className="material-symbols-outlined text-red-300 text-[24px]">flag</span>
                    <span className="text-[9px] uppercase font-bold text-red-300 mt-1 tracking-wider">Holiday</span>
                </div>
                <div className="absolute top-2 right-2 w-1.5 h-1.5 rounded-full bg-red-300" />
                {schedule?.reason && (
                    <span className="text-[8px] text-red-400 mt-auto truncate w-full text-center" title={schedule.reason}>
                        {schedule.reason}
                    </span>
                )}
            </div>
        );
    }

    /* ---------- celebration ---------- */
    if (status === "celebration") {
        return (
            <div
                className={`${baseClasses} bg-yellow-50 shadow-[0_0_15px_rgba(250,140,71,0.15)] ring-1 ring-yellow-200/50 hover:shadow-[var(--shadow-clay-button-hover)] overflow-hidden`}
            >
                <div className="absolute inset-0 bg-gradient-to-br from-yellow-100/30 to-transparent pointer-events-none" />
                <span className="text-sm font-bold text-yellow-600 relative z-10">{dayNumber}</span>
                <div className="flex flex-col items-center w-full mt-1 relative z-10">
                    <span className="material-symbols-outlined text-yellow-500 text-[24px]">celebration</span>
                    <span className="text-[9px] uppercase font-bold text-yellow-600 mt-1 tracking-wider">Event</span>
                </div>
                {schedule?.reason && (
                    <span
                        className="self-end mt-auto bg-yellow-100/80 px-2 py-0.5 rounded-md text-[8px] font-bold text-yellow-700 shadow-sm relative z-10 truncate max-w-full text-center"
                        title={schedule.reason}
                    >
                        {schedule.reason}
                    </span>
                )}
            </div>
        );
    }

    /* ---------- normal ---------- */
    const meals = schedule?.available_meals
        ? schedule.available_meals.split(",").map((m) => m.trim())
        : [];

    const mealDotColor: Record<string, string> = {
        lunch: "bg-orange-400",
        snacks: "bg-blue-400",
        iftar: "bg-indigo-400",
        event_dinner: "bg-yellow-400",
        optional_dinner: "bg-purple-400",
    };

    return (
        <div
            className={`${baseClasses} bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] border ${today
                ? "ring-2 ring-[var(--color-primary)]/30 bg-white shadow-[inset_0_0_0_1px_var(--color-primary)]/10"
                : "border-transparent hover:border-[var(--color-primary)]/20"
                }`}
        >
            <span
                className={`text-sm font-bold ${today ? "text-[var(--color-primary)]" : "text-[var(--color-text-sub)] group-hover:text-[var(--color-primary)]"
                    } transition-colors`}
            >
                {dayNumber}
            </span>

            {today && (
                <span className="absolute top-2 right-2 flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[var(--color-primary)] opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-[var(--color-primary)]" />
                </span>
            )}

            {meals.length > 0 && (
                <div className="flex gap-1 mt-1">
                    {meals.map((m) => (
                        <span
                            key={m}
                            className={`w-1.5 h-1.5 rounded-full ${mealDotColor[m] ?? "bg-gray-400"}`}
                            title={m.replace("_", " ")}
                        />
                    ))}
                </div>
            )}

            {schedule?.reason && (
                <span
                    className="text-[8px] text-[var(--color-text-sub)] mt-auto truncate w-full"
                    title={schedule.reason}
                >
                    {schedule.reason}
                </span>
            )}
        </div>
    );
};

/* ------------------------------------------------------------------ */
/*  Legend                                                              */
/* ------------------------------------------------------------------ */
const CalendarLegend: React.FC = () => (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-[var(--color-text-sub)] mt-6">
        <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded bg-[var(--color-background-light)] border border-[#e6dccf]" />
            <span>Normal</span>
        </div>
        <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded bg-red-50 border border-red-200" />
            <span>Govt Holiday</span>
        </div>
        <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded bg-yellow-50 border border-yellow-200" />
            <span>Celebration</span>
        </div>
        <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded bg-gray-100 border border-gray-200" />
            <span>Closed / Weekend</span>
        </div>
        <div className="border-l border-[#e6dccf] pl-4 flex items-center gap-3">
            <div className="flex items-center gap-1">
                <span className="w-1.5 h-1.5 rounded-full bg-orange-400" />
                <span>Lunch</span>
            </div>
            <div className="flex items-center gap-1">
                <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                <span>Snacks</span>
            </div>
            <div className="flex items-center gap-1">
                <span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />
                <span>Iftar</span>
            </div>
        </div>
    </div>
);

/* ------------------------------------------------------------------ */
/*  Schedule Page                                                      */
/* ------------------------------------------------------------------ */
export const Schedule: React.FC = () => {
    const { user } = useAuth();
    const [currentMonth, setCurrentMonth] = useState(new Date());
    const [schedules, setSchedules] = useState<ScheduleEntry[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState("");
    const [isModalOpen, setIsModalOpen] = useState(false);

    /* Fetch schedules for the visible month */
    const fetchSchedules = async (month: Date) => {
        try {
            setIsLoading(true);
            setError("");
            const start = format(startOfMonth(month), "yyyy-MM-dd");
            const end = format(endOfMonth(month), "yyyy-MM-dd");
            const response = await scheduleService.getScheduleRange(start, end);

            if (response.success && response.data) {
                setSchedules(response.data);
            } else {
                setSchedules([]);
            }
        } catch (err: any) {
            console.error("Error fetching schedules:", err);
            const msg = err?.response?.data?.error?.message || "Failed to load schedule data.";
            setError(msg);
            toast.error(msg);
            setSchedules([]);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchSchedules(currentMonth);
    }, [currentMonth]);

    const handlePrevMonth = () => setCurrentMonth((prev) => subMonths(prev, 1));
    const handleNextMonth = () => setCurrentMonth((prev) => addMonths(prev, 1));

    /* Build a lookup map: "YYYY-MM-DD" → ScheduleEntry */
    const scheduleMap = useMemo(() => {
        const map = new Map<string, ScheduleEntry>();
        schedules.forEach((s) => {
            // API may return "2026-02-06T00:00:00Z" — extract just YYYY-MM-DD
            const dateKey = s.date.substring(0, 10);
            map.set(dateKey, s);
        });
        return map;
    }, [schedules]);

    /* Build calendar grid cells */
    const calendarCells = useMemo(() => {
        const monthStart = startOfMonth(currentMonth);
        const startDowSunday = getDay(monthStart); // 0=Sun … 6=Sat
        // Convert to Monday-first: Mon=0, Tue=1, ..., Sun=6
        const startDow = startDowSunday === 0 ? 6 : startDowSunday - 1;
        const totalDays = getDaysInMonth(currentMonth);

        const cells: { date: Date; dayNumber: number; isCurrentMonth: boolean }[] = [];

        // Leading blanks (from previous month)
        for (let i = 0; i < startDow; i++) {
            cells.push({ date: new Date(0), dayNumber: 0, isCurrentMonth: false });
        }

        // Actual days
        for (let d = 1; d <= totalDays; d++) {
            const date = new Date(currentMonth.getFullYear(), currentMonth.getMonth(), d);
            cells.push({ date, dayNumber: d, isCurrentMonth: true });
        }

        // Trailing blanks (fill to complete last row)
        const remainder = cells.length % 7;
        if (remainder > 0) {
            for (let i = 0; i < 7 - remainder; i++) {
                cells.push({ date: new Date(0), dayNumber: 0, isCurrentMonth: false });
            }
        }

        return cells;
    }, [currentMonth]);

    const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

    if (isLoading && schedules.length === 0) {
        return <LoadingSpinner message="Loading schedule…" />;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header userName={user?.name} userRole={user?.role} />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
                {/* Page Title */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
                    <div>
                        <h2 className="text-4xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
                            Schedule
                        </h2>
                        <p className="text-lg text-[var(--color-text-sub)] font-medium">
                            Monthly day-status calendar for operations
                        </p>
                    </div>
                    <button
                        onClick={() => setIsModalOpen(true)}
                        className="px-5 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 text-sm"
                    >
                        <span className="material-symbols-outlined text-[18px]">add</span>
                        Create Schedule
                    </button>
                </div>

                {/* Calendar Card */}
                <div className="bg-[#FFFDF5] p-6 md:p-8 rounded-[2rem] shadow-[var(--shadow-clay-card)] border border-white/60 flex-1">
                    {/* Month Navigation */}
                    <div className="flex flex-col sm:flex-row justify-between items-center gap-4 mb-8">
                        <div className="flex items-center gap-4">
                            <button
                                onClick={handlePrevMonth}
                                className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] flex items-center justify-center hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                            >
                                <span className="material-symbols-outlined">chevron_left</span>
                            </button>
                            <h3 className="text-xl font-bold text-[var(--color-background-dark)] w-44 text-center">
                                {format(currentMonth, "MMMM yyyy")}
                            </h3>
                            <button
                                onClick={handleNextMonth}
                                className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] flex items-center justify-center hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                            >
                                <span className="material-symbols-outlined">chevron_right</span>
                            </button>
                        </div>

                        {/* Today button */}
                        <button
                            onClick={() => setCurrentMonth(new Date())}
                            className="px-4 py-2 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all flex items-center gap-2 text-sm"
                        >
                            <span className="material-symbols-outlined text-[18px]">today</span>
                            Today
                        </button>
                    </div>

                    {/* Weekday Headers */}
                    <div className="grid grid-cols-7 gap-3 mb-4 text-center">
                        {WEEKDAYS.map((d) => (
                            <div key={d} className="text-xs font-bold text-[var(--color-text-sub)] uppercase tracking-wider">
                                {d}
                            </div>
                        ))}
                    </div>

                    {/* Calendar Grid */}
                    <div className="grid grid-cols-7 gap-3">
                        {calendarCells.map((cell, idx) => {
                            const dateStr = cell.isCurrentMonth ? format(cell.date, "yyyy-MM-dd") : "";
                            const schedule = dateStr ? scheduleMap.get(dateStr) : undefined;

                            return (
                                <DayCell
                                    key={idx}
                                    dayNumber={cell.dayNumber}
                                    date={cell.date}
                                    schedule={schedule}
                                    isCurrentMonth={cell.isCurrentMonth}
                                />
                            );
                        })}
                    </div>

                    {/* Legend */}
                    <CalendarLegend />
                </div>

                {error && (
                    <div className="mt-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm">
                        {error}
                    </div>
                )}
            </main>

            <Footer links={[{ label: "Privacy", href: "#" }, { label: "Terms", href: "#" }, { label: "Support", href: "#" }]} />

            {/* Create Schedule Modal */}
            <CreateScheduleModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSuccess={() => {
                    fetchSchedules(currentMonth);
                }}
            />
        </div>
    );
};
