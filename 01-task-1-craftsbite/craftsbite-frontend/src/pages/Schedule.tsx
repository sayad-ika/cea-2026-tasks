import React, { useState, useEffect } from "react";
import { format, isWeekend } from "date-fns";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as scheduleService from "../services/scheduleService";
import type { DayStatus } from "../types";
import toast from "react-hot-toast";
import type {
    CreateScheduleRequest,
    ScheduleEntry,
} from "../types/schedule.types";
import { CreateScheduleModal } from "../components/modals/CreateScheduleModal";
import { STATUS_CONFIG, MEAL_LABELS } from "../types/schedule.types";

export const Schedule: React.FC = () => {
    const { user } = useAuth();
    const today = new Date();
    const todayKey = format(today, "yyyy-MM-dd");
    const isWeekendDay = isWeekend(today);

    const [schedule, setSchedule] = useState<ScheduleEntry | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);

    const fetchTodaySchedule = async () => {
        try {
            setIsLoading(true);
            const response = await scheduleService.getScheduleByDate(todayKey);
            if (response.success && response.data) {
                setSchedule(response.data);
            } else {
                setSchedule(null);
            }
        } catch (err: any) {
            const msg =
                err?.response?.data?.error?.message ||
                "Failed to load today's schedule.";
            toast.error(msg);
            setSchedule(null);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchTodaySchedule();
    }, []);

    const handleScheduleSubmit = async (payload: CreateScheduleRequest) => {
        try {
            await scheduleService.createSchedule(payload);
            fetchTodaySchedule();
            toast.success("Schedule created successfully!");
        } catch (err: any) {
            const msg =
                err?.response?.data?.error?.message ||
                "Failed to create schedule.";
            toast.error(msg);
            throw err;
        }
    };

    const effectiveStatus: DayStatus | "weekend" =
        schedule?.day_status ?? (isWeekendDay ? "weekend" : "normal");

    const config = STATUS_CONFIG[effectiveStatus];

    const meals: string[] = schedule?.available_meals
        ? schedule.available_meals
              .split(",")
              .map((m) => m.trim())
              .filter(Boolean)
        : [];

    if (isLoading) {
        return <LoadingSpinner message="Loading today's schedule…" />;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header userName={user?.name} userRole={user?.role} />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 max-w-2xl flex flex-col">
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
                    <div>
                        <h2 className="text-4xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
                            Today
                        </h2>
                        <p className="text-lg text-[var(--color-text-sub)] font-medium">
                            {format(today, "EEEE, MMMM d, yyyy")}
                        </p>
                    </div>

                    {!schedule && !isWeekendDay && (
                        <button
                            onClick={() => setIsModalOpen(true)}
                            className="px-5 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 text-sm self-start md:self-auto"
                        >
                            <span className="material-symbols-outlined text-[18px]">
                                add
                            </span>
                            Create Schedule
                        </button>
                    )}
                </div>

                <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 shadow-[var(--shadow-clay-card)]">
                    <div className="flex items-center gap-4 mb-6">
                        <div className="w-14 h-14 rounded-2xl flex items-center justify-center shadow-[var(--shadow-clay-button)] bg-white">
                            <span className="material-symbols-outlined text-[28px] text-[var(--color-text-sub)]">
                                {config.icon}
                            </span>
                        </div>
                        <div>
                            <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-0.5">
                                Day Status
                            </p>
                            <h3 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)]">
                                {config.label}
                            </h3>
                        </div>
                    </div>

                    <div className="border-t border-[#e6dccf] pt-6 flex flex-col gap-5">
                        {schedule?.reason && (
                            <div>
                                <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-1.5">
                                    Note
                                </p>
                                <p className="text-[var(--color-text-main)] font-medium">
                                    {schedule.reason}
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

                        {schedule && (
                            <button
                                onClick={() => setIsModalOpen(true)}
                                className="self-start mt-2 px-4 py-2 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all flex items-center gap-2 text-sm"
                            >
                                <span className="material-symbols-outlined text-[18px]">
                                    add
                                </span>
                                Create Schedule
                            </button>
                        )}
                    </div>
                </div>

                {!schedule && isWeekendDay && (
                    <div className="mt-6 bg-[#FFFDF5] border border-[#e6dccf] text-[var(--color-text-sub)] px-6 py-4 rounded-2xl text-sm flex items-center gap-3">
                        <span className="material-symbols-outlined text-[20px]">
                            info
                        </span>
                        It's the weekend — no schedule needed.
                    </div>
                )}

                {!schedule && !isWeekendDay && (
                    <div className="mt-6 bg-[#FFFDF5] border border-[#e6dccf] px-6 py-4 rounded-2xl text-sm text-[var(--color-text-sub)] flex items-center gap-3">
                        <span className="material-symbols-outlined text-[20px]">
                            info
                        </span>
                        No schedule has been created for today yet. Use the
                        button above to add one.
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

            <CreateScheduleModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSubmit={handleScheduleSubmit}
                defaultDate={todayKey}
            />
        </div>
    );
};
