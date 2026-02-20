import React, { useState, useEffect } from "react";
import {
    DAY_STATUS_OPTIONS,
    MEAL_OPTIONS,
    type CreateScheduleRequest,
    type DayStatus,
    type MealType,
    type ScheduleEntry,
} from "../../types";

interface CreateScheduleModalProps {
    isOpen: boolean;
    defaultDate?: string;
    existingSchedule?: ScheduleEntry | null;
    onClose: () => void;
    onSubmit: (payload: CreateScheduleRequest) => Promise<void>;
}

export const CreateScheduleModal: React.FC<CreateScheduleModalProps> = ({
    isOpen,
    defaultDate = "",
    onClose,
    onSubmit,
}) => {
    const [date, setDate] = useState<string>(defaultDate);
    const [dayStatus, setDayStatus] = useState<DayStatus>("normal");
    const [reason, setReason] = useState<string>("");
    const [meals, setMeals] = useState<MealType[]>([]);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string>("");

    useEffect(() => {
        if (isOpen) {
            setDate(defaultDate);
            setDayStatus("normal");
            setReason("");
            setMeals([]);
        }
    }, [isOpen, defaultDate]);

    const toggleMeal = (meal: MealType) => {
        setMeals((prev) =>
            prev.includes(meal)
                ? prev.filter((m) => m !== meal)
                : [...prev, meal],
        );
    };

    const showMeals = dayStatus !== "office_closed";

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!date) return;

        const payload: CreateScheduleRequest = {
            date,
            day_status: dayStatus,
            ...(reason.trim() && { reason: reason.trim() }),
            available_meals: showMeals ? meals : [],
        };

        try {
            setIsSubmitting(true);
            await onSubmit(payload);
            onClose();
        } catch (err: any) {
            const errorMessage =
                err?.message || "Failed to create schedule. Please try again.";
            setError(errorMessage);
        } finally {
            setIsSubmitting(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/30 backdrop-blur-sm"
            onClick={(e) => e.target === e.currentTarget && onClose()}
        >
            <div className="w-full max-w-md bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] shadow-[var(--shadow-clay-card)] p-8 flex flex-col gap-6">
                <div className="flex items-center justify-between">
                    <h2 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)]">
                        Create Schedule
                    </h2>
                    <button
                        type="button"
                        onClick={onClose}
                        className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
                    >
                        <span className="material-symbols-outlined text-[20px]">
                            close
                        </span>
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                    {error && (
                        <div className="flex items-center gap-3 p-4 rounded-xl bg-red-50 border border-red-200">
                            <div>
                                <p className="text-sm font-semibold text-red-600">
                                    Error
                                </p>
                                <p className="text-sm text-red-600 mt-1">
                                    {error}
                                </p>
                            </div>
                        </div>
                    )}

                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Date
                        </label>
                        <input
                            type="date"
                            required
                            value={date}
                            onChange={(e) => setDate(e.target.value)}
                            className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                        />
                    </div>

                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Day Status
                        </label>
                        <div className="grid grid-cols-2 gap-2">
                            {DAY_STATUS_OPTIONS.map((opt) => (
                                <button
                                    key={opt.value}
                                    type="button"
                                    onClick={() => setDayStatus(opt.value)}
                                    className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all ${
                                        dayStatus === opt.value
                                            ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white shadow-[var(--shadow-clay-button)]"
                                            : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                    }`}
                                >
                                    {opt.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Reason{" "}
                            <span className="normal-case tracking-normal font-normal opacity-60">
                                (optional)
                            </span>
                        </label>
                        <input
                            type="text"
                            placeholder="e.g. National Day, Team event…"
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all placeholder:text-[var(--color-text-sub)]/50"
                        />
                    </div>

                    {showMeals && (
                        <div className="flex flex-col gap-1.5">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                                Available Meals{" "}
                                <span className="normal-case tracking-normal font-normal opacity-60">
                                    (optional)
                                </span>
                            </label>
                            <div className="flex flex-wrap gap-2">
                                {MEAL_OPTIONS.map((opt) => (
                                    <button
                                        key={opt.value}
                                        type="button"
                                        onClick={() => toggleMeal(opt.value)}
                                        className={`px-3 py-1.5 rounded-xl border text-xs font-semibold transition-all ${
                                            meals.includes(opt.value)
                                                ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white"
                                                : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                        }`}
                                    >
                                        {opt.label}
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="border-t border-[#e6dccf]" />

                    <div className="flex items-center justify-end gap-3">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-sub)] text-sm font-semibold shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={!date || isSubmitting}
                            className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
                        >
                            {isSubmitting ? (
                                <>
                                    <span className="material-symbols-outlined text-[16px] animate-spin">
                                        progress_activity
                                    </span>{" "}
                                    Saving…
                                </>
                            ) : (
                                <>
                                    <span className="material-symbols-outlined text-[16px]">
                                        add
                                    </span>{" "}
                                    Create
                                </>
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
