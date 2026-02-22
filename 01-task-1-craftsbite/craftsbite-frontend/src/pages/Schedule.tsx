import React, { useState, useEffect, useCallback } from "react";
import {
    format,
    startOfMonth,
    endOfMonth,
    addMonths,
    subMonths,
    isWeekend,
} from "date-fns";
import { Header, Footer, Navbar, LoadingSpinner, ScheduleCalendar } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as scheduleService from "../services/scheduleService";
import toast from "react-hot-toast";
import type {
    CreateScheduleRequest,
    ScheduleEntry,
} from "../types/schedule.types";
import { CreateScheduleModal } from "../components/modals/CreateScheduleModal";
import { STATUS_CONFIG, MEAL_LABELS, STATUS_COLORS_FOR_SCHEDULE } from "../types/schedule.types";


export const Schedule: React.FC = () => {
    const { user } = useAuth();
    const today = new Date();

    const [currentMonth, setCurrentMonth] = useState(
        startOfMonth(today),
    );
    const [scheduleMap, setScheduleMap] = useState<
        Record<string, ScheduleEntry>
    >({});
    const [isLoading, setIsLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedDate, setSelectedDate] = useState<string>(
        format(today, "yyyy-MM-dd"),
    );

    const fetchMonthSchedules = useCallback(async (monthStart: Date) => {
        try {
            setIsLoading(true);
            const start = format(monthStart, "yyyy-MM-dd");
            const end = format(endOfMonth(monthStart), "yyyy-MM-dd");
            const response = await scheduleService.getScheduleRange(start, end);
            if (response.success && Array.isArray(response.data)) {
                const map: Record<string, ScheduleEntry> = {};
                for (const entry of response.data) {
                    const key = format(new Date(entry.date), "yyyy-MM-dd");
                    map[key] = entry;
                }
                setScheduleMap(map);
            } else {
                setScheduleMap({});
            }
        } catch {
            toast.error("Failed to load schedules for this month.");
            setScheduleMap({});
        } finally {
            setIsLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchMonthSchedules(currentMonth);
    }, [currentMonth, fetchMonthSchedules]);

    const handleScheduleSubmit = async (payload: CreateScheduleRequest) => {
        try {
            await scheduleService.createSchedule(payload);
            await fetchMonthSchedules(currentMonth);
            toast.success("Schedule created successfully!");
        } catch (err: any) {
            const msg =
                err?.response?.data?.error?.message ||
                "Failed to create schedule.";
            toast.error(msg);
            throw err;
        }
    };

    if (isLoading) {
        return <LoadingSpinner message="Loading schedule…" />;
    }

    const selectedEntry = scheduleMap[selectedDate] ?? null;
    const isSelectedWeekend = isWeekend(new Date(selectedDate + "T12:00:00"));
    const effectiveStatus =
        selectedEntry?.day_status ?? (isSelectedWeekend ? "weekend" : "normal");
    const config = STATUS_CONFIG[effectiveStatus];
    const meals: string[] = selectedEntry?.available_meals
        ? selectedEntry.available_meals
              .split(",")
              .map((m) => m.trim())
              .filter(Boolean)
        : [];

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header userName={user?.name} userRole={user?.role} />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 max-w-5xl flex flex-col gap-6">
                {/* Page header */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
                    <div>
                        <h2 className="text-4xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
                            Schedule
                        </h2>
                        <p className="text-lg text-[var(--color-text-sub)] font-medium">
                            Monthly meal schedule overview
                        </p>
                    </div>

                    <button
                        onClick={() => {
                            setSelectedDate(format(today, "yyyy-MM-dd"));
                            setIsModalOpen(true);
                        }}
                        className="px-5 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 text-sm self-start md:self-auto"
                    >
                        <span className="material-symbols-outlined text-[18px]">
                            add
                        </span>
                        Create Schedule
                    </button>
                </div>

                {/* Calendar */}
                <ScheduleCalendar
                    currentMonth={currentMonth}
                    scheduleMap={scheduleMap}
                    selectedDate={selectedDate}
                    onSelectDate={setSelectedDate}
                    onPrevMonth={() => setCurrentMonth((m) => subMonths(m, 1))}
                    onNextMonth={() => setCurrentMonth((m) => addMonths(m, 1))}
                />

                {/* Selected day detail */}
                <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 shadow-[var(--shadow-clay-card)]">
                    <div className="flex items-center gap-4 mb-6">
                        <div className="w-14 h-14 rounded-2xl flex items-center justify-center shadow-[var(--shadow-clay-button)] bg-white">
                            <span className="material-symbols-outlined text-[28px] text-[var(--color-text-sub)]">
                                {config.icon}
                            </span>
                        </div>
                        <div>
                            <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-0.5">
                                {format(
                                    new Date(selectedDate + "T12:00:00"),
                                    "EEEE, MMMM d, yyyy",
                                )}
                            </p>
                            <h3 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)]">
                                {config.label}
                            </h3>
                        </div>

                        {/* Status badge */}
                        {selectedEntry && (
                            <span
                                className={`ml-auto px-3 py-1 rounded-xl text-xs font-semibold border ${STATUS_COLORS_FOR_SCHEDULE[selectedEntry.day_status]}`}
                            >
                                {STATUS_CONFIG[selectedEntry.day_status]?.label ?? selectedEntry.day_status}
                            </span>
                        )}
                    </div>

                    <div className="border-t border-[#e6dccf] pt-6 flex flex-col gap-5">
                        {selectedEntry?.reason && (
                            <div>
                                <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-1.5">
                                    Note
                                </p>
                                <p className="text-[var(--color-text-main)] font-medium">
                                    {selectedEntry.reason}
                                </p>
                            </div>
                        )}

                        {meals.length > 0 && (
                            <div>
                                <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-2">
                                    Available Meals
                                </p>
                                <div className="flex flex-wrap gap-2">
                                    {meals.map((m) => (
                                        <span
                                            key={m}
                                            className="inline-flex items-center px-3 py-1.5 rounded-xl text-xs font-semibold bg-[var(--color-background-light)] border border-[#e6dccf] text-[var(--color-text-sub)]"
                                        >
                                            {MEAL_LABELS[m] ?? m}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}

                        {!selectedEntry && !isSelectedWeekend && (
                            <div className="bg-[var(--color-background-light)] border border-[#e6dccf] px-5 py-3 rounded-2xl text-sm text-[var(--color-text-sub)] flex items-center gap-3">
                                <span className="material-symbols-outlined text-[20px]">
                                    info
                                </span>
                                No schedule set for this day.
                            </div>
                        )}

                        {!selectedEntry && isSelectedWeekend && (
                            <div className="bg-[var(--color-background-light)] border border-[#e6dccf] px-5 py-3 rounded-2xl text-sm text-[var(--color-text-sub)] flex items-center gap-3">
                                <span className="material-symbols-outlined text-[20px]">
                                    info
                                </span>
                                It's a weekend — no schedule needed.
                            </div>
                        )}

                        <button
                            onClick={() => setIsModalOpen(true)}
                            className="self-start mt-2 px-4 py-2 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all flex items-center gap-2 text-sm"
                        >
                            <span className="material-symbols-outlined text-[18px]">
                                add
                            </span>
                            {selectedEntry ? "Create Another" : "Create Schedule"}
                        </button>
                    </div>
                </div>
            </main>

            <Footer
                links={[
                    { label: "Privacy", href: "#" },
                    { label: "Terms", href: "#" },
                    { label: "Support", href: "#" },
                ]}
            />

            <CreateScheduleModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSubmit={handleScheduleSubmit}
                defaultDate={selectedDate}
            />
        </div>
    );
};
