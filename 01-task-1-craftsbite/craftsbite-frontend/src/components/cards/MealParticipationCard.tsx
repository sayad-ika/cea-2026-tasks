import React, { useState, useEffect } from "react";
import { MEAL_OPTIONS } from "../../types/schedule.types";
import * as mealService from "../../services/mealService";
import toast from "react-hot-toast";
import type { MealType } from "../../types";
import { FORWARD_WINDOW_DAYS } from "../../utils";


function getTodayDateStr(): string {
    return new Date().toISOString().split("T")[0];
}

function getMaxDateStr(): string {
    const d = new Date();
    d.setDate(d.getDate() + FORWARD_WINDOW_DAYS);
    return d.toISOString().split("T")[0];
}

function getTomorrowDateStr(): string {
    const d = new Date();
    d.setDate(d.getDate() + 1);
    return d.toISOString().split("T")[0];
}

// ─── Modal ───────────────────────────────────────────────────────────────────

interface SetMealParticipationModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

const SetMealParticipationModal: React.FC<SetMealParticipationModalProps> = ({
    isOpen,
    onClose,
    onSuccess,
}) => {
    const [date, setDate] = useState(getTomorrowDateStr());
    const [selectedMeals, setSelectedMeals] = useState<MealType[]>([]);
    const [participating, setParticipating] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        if (isOpen) {
            setDate(getTomorrowDateStr());
            setSelectedMeals([]);
            setParticipating(true);
            setError("");
        }
    }, [isOpen]);

    const allSelected = MEAL_OPTIONS.every((o) => selectedMeals.includes(o.value));

    const toggleAll = () => {
        setSelectedMeals(allSelected ? [] : MEAL_OPTIONS.map((o) => o.value));
    };

    const toggleMeal = (meal: MealType) => {
        setSelectedMeals((prev) =>
            prev.includes(meal) ? prev.filter((m) => m !== meal) : [...prev, meal],
        );
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!date || selectedMeals.length === 0) return;

        try {
            setIsSubmitting(true);
            setError("");

            await Promise.all(
                selectedMeals.map((meal) =>
                    mealService.setMealParticipation({
                        date,
                        meal_type: meal,
                        participating,
                    }),
                ),
            );

            toast.success(`Participation ${participating ? "confirmed" : "cancelled"}!`);
            onSuccess();
            onClose();
        } catch (err: any) {
            const msg =
                err?.response?.data?.error?.message ||
                "Failed to update meal participation.";
            setError(msg);
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
                {/* Header */}
                <div className="flex items-center justify-between">
                    <h2 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)]">
                        Meal Participation
                    </h2>
                    <button
                        type="button"
                        onClick={onClose}
                        className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
                    >
                        <span className="material-symbols-outlined text-[20px]">close</span>
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                    {/* Error */}
                    {error && (
                        <div className="flex items-center gap-3 p-4 rounded-xl bg-red-50 border border-red-200">
                            <div>
                                <p className="text-sm font-semibold text-red-600">Error</p>
                                <p className="text-sm text-red-600 mt-1">{error}</p>
                            </div>
                        </div>
                    )}

                    {/* Date */}
                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Date
                        </label>
                        <input
                            type="date"
                            required
                            value={date}
                            min={getTodayDateStr()}
                            max={getMaxDateStr()}
                            onChange={(e) => setDate(e.target.value)}
                            className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                        />
                        <p className="text-[11px] text-[var(--color-text-sub)] opacity-70 ml-1">
                            You can update participation up to {FORWARD_WINDOW_DAYS} days ahead.
                        </p>
                    </div>

                    {/* Participating toggle */}
                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Participation
                        </label>
                        <div className="grid grid-cols-2 gap-2">
                            <button
                                type="button"
                                onClick={() => setParticipating(true)}
                                className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${
                                    participating
                                        ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white shadow-[var(--shadow-clay-button)]"
                                        : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                }`}
                            >
                                <span className="material-symbols-outlined text-[16px]">check_circle</span>
                                Opting In
                            </button>
                            <button
                                type="button"
                                onClick={() => setParticipating(false)}
                                className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${
                                    !participating
                                        ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white shadow-[var(--shadow-clay-button)]"
                                        : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                }`}
                            >
                                <span className="material-symbols-outlined text-[16px]">cancel</span>
                                Opting Out
                            </button>
                        </div>
                    </div>

                    {/* Meals */}
                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Meals
                        </label>
                        <div className="flex flex-wrap gap-2">
                            <button
                                type="button"
                                onClick={toggleAll}
                                className={`px-3 py-1.5 rounded-xl border text-xs font-semibold transition-all ${
                                    allSelected
                                        ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white"
                                        : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                }`}
                            >
                                All
                            </button>

                            {MEAL_OPTIONS.map((opt) => (
                                <button
                                    key={opt.value}
                                    type="button"
                                    onClick={() => toggleMeal(opt.value)}
                                    className={`px-3 py-1.5 rounded-xl border text-xs font-semibold transition-all ${
                                        selectedMeals.includes(opt.value)
                                            ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white"
                                            : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                                    }`}
                                >
                                    {opt.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="border-t border-[#e6dccf]" />

                    {/* Actions */}
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
                            disabled={!date || selectedMeals.length === 0 || isSubmitting}
                            className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
                        >
                            {isSubmitting ? (
                                <>
                                    <span className="material-symbols-outlined text-[16px] animate-spin">
                                        progress_activity
                                    </span>
                                    Saving…
                                </>
                            ) : (
                                <>
                                    <span className="material-symbols-outlined text-[16px]">
                                        check
                                    </span>
                                    Confirm
                                </>
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// ─── Card ────────────────────────────────────────────────────────────────────

export const MealParticipationCard: React.FC = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);

    return (
        <>
            <div className="mb-6 bg-[var(--color-background-light)] rounded-3xl p-6 md:p-8 flex flex-col md:flex-row items-center justify-between gap-6 border-2 border-white/50 shadow-[var(--shadow-clay)]">
                <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-full bg-[var(--color-primary)]/10 flex items-center justify-center text-[var(--color-primary)] shadow-[var(--shadow-clay-inset)]">
                        <span className="material-symbols-outlined text-2xl">restaurant</span>
                    </div>
                    <div>
                        <h3 className="text-xl font-bold text-left text-[var(--color-background-dark)]">
                            Meal Participation
                        </h3>
                        <p className="text-sm text-[var(--color-text-sub)]">
                            Opt in or out of meals for any upcoming day
                        </p>
                    </div>
                </div>

                <button
                    type="button"
                    onClick={() => setIsModalOpen(true)}
                    className="flex items-center gap-2 px-6 py-3 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-sm hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
                    style={{ boxShadow: "var(--shadow-clay-button)" }}
                >
                    <span className="material-symbols-outlined text-lg">edit_note</span>
                    Set Participation
                </button>
            </div>

            <SetMealParticipationModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSuccess={() => {}}
            />
        </>
    );
};
