import React, { useState, useEffect } from "react";
import type { DayStatus, MealType, ScheduleEntry } from "../../types";
import toast from "react-hot-toast";

interface ScheduleModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
    /** If provided, modal acts as edit; otherwise it's create */
    editingSchedule?: ScheduleEntry | null;
}

interface ScheduleFormData {
    date: string;
    dayStatus: DayStatus;
    reason: string;
    availableMeals: MealType[];
}

const defaultFormData: ScheduleFormData = {
    date: "",
    dayStatus: "normal",
    reason: "",
    availableMeals: [],
};

export const ScheduleModal: React.FC<ScheduleModalProps> = ({
    isOpen,
    onClose,
    onSuccess,
    editingSchedule,
}) => {
    const isEditMode = !!editingSchedule;

    const [formData, setFormData] = useState<ScheduleFormData>(defaultFormData);
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Populate form when editing
    useEffect(() => {
        if (isOpen && editingSchedule) {
            const meals = editingSchedule.available_meals
                ? (editingSchedule.available_meals.split(",").map((m) => m.trim()) as MealType[])
                : [];
            setFormData({
                date: editingSchedule.date.substring(0, 10),
                dayStatus: editingSchedule.day_status,
                reason: editingSchedule.reason || "",
                availableMeals: meals,
            });
        } else if (isOpen && !editingSchedule) {
            setFormData(defaultFormData);
        }
    }, [isOpen, editingSchedule]);

    if (!isOpen) return null;

    const handleMealToggle = (meal: MealType) => {
        setFormData((prev) => ({
            ...prev,
            availableMeals: prev.availableMeals.includes(meal)
                ? prev.availableMeals.filter((m) => m !== meal)
                : [...prev.availableMeals, meal],
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!formData.date) {
            toast.error("Please select a date");
            return;
        }

        if (
            formData.dayStatus !== "office_closed" &&
            formData.dayStatus !== "govt_holiday" &&
            formData.availableMeals.length === 0
        ) {
            toast.error("Please select at least one available meal");
            return;
        }

        setIsSubmitting(true);

        try {
            const scheduleService = await import("../../services/scheduleService");
            const payload = {
                date: formData.date,
                day_status: formData.dayStatus,
                reason: formData.reason || undefined,
                available_meals: formData.availableMeals,
            };

            if (isEditMode) {
                await scheduleService.updateSchedule(formData.date, payload);
                toast.success("Schedule updated successfully!");
            } else {
                await scheduleService.createSchedule(payload);
                toast.success("Schedule created successfully!");
            }

            onSuccess();
            handleClose();
        } catch (error: any) {
            console.error(`Error ${isEditMode ? "updating" : "creating"} schedule:`, error);
            const msg =
                error?.response?.data?.error?.message ||
                `Failed to ${isEditMode ? "update" : "create"} schedule`;
            toast.error(msg);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleClose = () => {
        setFormData(defaultFormData);
        onClose();
    };

    const mealOptions: Array<{ value: MealType; label: string; icon: string; color: string }> = [
        { value: "lunch", label: "Lunch", icon: "restaurant", color: "orange" },
        { value: "snacks", label: "Snacks", icon: "cookie", color: "yellow" },
        { value: "iftar", label: "Iftar", icon: "nightlight", color: "indigo" },
        { value: "event_dinner", label: "Event Dinner", icon: "celebration", color: "purple" },
        { value: "optional_dinner", label: "Optional Dinner", icon: "dinner_dining", color: "pink" },
    ];

    const dayStatusOptions: Array<{ value: DayStatus; label: string }> = [
        { value: "normal", label: "Normal" },
        { value: "office_closed", label: "Office Closed" },
        { value: "govt_holiday", label: "Government Holiday" },
        { value: "celebration", label: "Celebration" },
    ];

    return (
        <div
            className="modal-overlay fixed inset-0 z-[100] flex items-center justify-center bg-[#23170f]/30 backdrop-blur-sm p-4"
            onClick={(e) => {
                if (e.target === e.currentTarget) handleClose();
            }}
        >
            <div className="bg-[#FFFDF5] rounded-[2rem] shadow-[var(--shadow-clay)] max-w-lg w-full p-8 border border-white/60 relative overflow-hidden max-h-[90vh] overflow-y-auto">
                {/* Header */}
                <div className="flex justify-between items-center mb-8">
                    <h2 className="text-2xl font-black text-[#23170f]">
                        {isEditMode ? "Update Schedule" : "Create Schedule"}
                    </h2>
                    <button
                        type="button"
                        onClick={handleClose}
                        className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] flex items-center justify-center text-[var(--color-text-sub)] hover:text-[var(--color-primary)] hover:shadow-[var(--shadow-clay-button-hover)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                    >
                        <span className="material-symbols-outlined">close</span>
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="space-y-6">
                    {/* Date Input */}
                    <div className="space-y-2">
                        <label className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">
                            Select Date
                        </label>
                        <div className="relative group">
                            <input
                                type="date"
                                value={formData.date}
                                onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                                className="w-full bg-[var(--color-background-light)] text-[#23170f] rounded-2xl border-none shadow-[var(--shadow-clay-inset)] px-4 py-3 focus:ring-2 focus:ring-[var(--color-primary)]/50 transition-all cursor-pointer font-medium disabled:opacity-60 disabled:cursor-not-allowed"
                                required
                                disabled={isEditMode}
                            />
                            <span className="absolute right-4 top-1/2 -translate-y-1/2 text-[var(--color-text-sub)] pointer-events-none group-hover:text-[var(--color-primary)] transition-colors">
                                <span className="material-symbols-outlined">calendar_month</span>
                            </span>
                        </div>
                    </div>

                    {/* Day Status Dropdown */}
                    <div className="space-y-2">
                        <label className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">
                            Day Status
                        </label>
                        <div className="relative">
                            <select
                                value={formData.dayStatus}
                                onChange={(e) =>
                                    setFormData({ ...formData, dayStatus: e.target.value as DayStatus })
                                }
                                className="w-full bg-[var(--color-background-light)] text-[#23170f] rounded-2xl border-none shadow-[var(--shadow-clay-button)] px-4 py-3 pr-10 focus:ring-0 cursor-pointer font-medium appearance-none"
                            >
                                {dayStatusOptions.map((option) => (
                                    <option key={option.value} value={option.value}>
                                        {option.label}
                                    </option>
                                ))}
                            </select>
                            <span className="absolute right-4 top-1/2 -translate-y-1/2 text-[var(--color-text-sub)] pointer-events-none">
                                <span className="material-symbols-outlined">expand_more</span>
                            </span>
                        </div>
                    </div>

                    {/* Reason Input */}
                    <div className="space-y-2">
                        <label className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">
                            Reason / Notes
                            <span className="text-xs normal-case text-[var(--color-text-sub)]/60 ml-1">
                                (Optional for normal days)
                            </span>
                        </label>
                        <textarea
                            value={formData.reason}
                            onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
                            placeholder="Add notes about this schedule..."
                            rows={3}
                            className="w-full bg-[var(--color-background-light)] text-[#23170f] rounded-2xl border-none shadow-[var(--shadow-clay-inset)] px-4 py-3 focus:ring-2 focus:ring-[var(--color-primary)]/50 resize-none placeholder:text-[var(--color-text-sub)]/40 font-medium"
                        />
                    </div>

                    {/* Available Meals */}
                    <div className="space-y-3 pt-2">
                        <label className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">
                            Available Meals
                        </label>
                        <div className="space-y-2">
                            {mealOptions.map((meal) => (
                                <label
                                    key={meal.value}
                                    className="flex items-center justify-between bg-[var(--color-background-light)] p-3 rounded-2xl shadow-[var(--shadow-clay-inset)] cursor-pointer hover:shadow-[var(--shadow-clay-button)] transition-all"
                                >
                                    <div className="flex items-center gap-3">
                                        <div
                                            className={`w-8 h-8 rounded-lg bg-${meal.color}-100 flex items-center justify-center text-${meal.color}-600`}
                                        >
                                            <span className="material-symbols-outlined text-[18px]">
                                                {meal.icon}
                                            </span>
                                        </div>
                                        <span className="font-bold text-[#23170f]">{meal.label}</span>
                                    </div>
                                    <input
                                        type="checkbox"
                                        checked={formData.availableMeals.includes(meal.value)}
                                        onChange={() => handleMealToggle(meal.value)}
                                        className="w-5 h-5 rounded border-2 border-[#e6dccf] text-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer transition-all"
                                    />
                                </label>
                            ))}
                        </div>
                    </div>

                    {/* Action Buttons */}
                    <div className="flex gap-4 pt-4 mt-6 border-t border-[#e6dccf]">
                        <button
                            type="button"
                            onClick={handleClose}
                            disabled={isSubmitting}
                            className="flex-1 py-3 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-bold shadow-[var(--shadow-clay-button)] hover:text-[#23170f] active:shadow-[var(--shadow-clay-button-active)] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="flex-1 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                        >
                            {isSubmitting
                                ? isEditMode
                                    ? "Updating..."
                                    : "Creating..."
                                : isEditMode
                                    ? "Update Schedule"
                                    : "Create Schedule"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
