import React, { useState, useEffect } from "react";
import { format, addDays, differenceInCalendarDays } from "date-fns";
import type { MealType } from "../../types";

interface BulkOptOutModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (startDate: string, endDate: string, reason: string, mealTypes: string[]) => Promise<void>;
  selectedCount: number;
}

const ALL_MEAL_TYPES: { id: MealType; label: string; icon: string }[] = [
  { id: "lunch", label: "Lunch", icon: "lunch_dining" },
  { id: "snacks", label: "Snacks", icon: "cookie" },
  { id: "iftar", label: "Iftar", icon: "dark_mode" },
  { id: "event_dinner", label: "Event Dinner", icon: "celebration" },
  { id: "optional_dinner", label: "Opt. Dinner", icon: "dinner_dining" },
];

export const BulkOptOutModal: React.FC<BulkOptOutModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  selectedCount,
}) => {
  const [startDate, setStartDate] = useState(format(new Date(), "yyyy-MM-dd"));
  const [endDate, setEndDate] = useState(format(addDays(new Date(), 2), "yyyy-MM-dd"));
  const [reason, setReason] = useState("");
  const [selectedMealTypes, setSelectedMealTypes] = useState<Set<MealType>>(
    new Set(["lunch", "snacks"])
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setStartDate(format(new Date(), "yyyy-MM-dd"));
      setEndDate(format(addDays(new Date(), 2), "yyyy-MM-dd"));
      setReason("");
      setSelectedMealTypes(new Set(["lunch", "snacks"]));
    }
  }, [isOpen]);

  const durationDays = differenceInCalendarDays(new Date(endDate), new Date(startDate)) + 1;
  const isAllSelected = ALL_MEAL_TYPES.every((t) => selectedMealTypes.has(t.id));

  const handleToggleMealType = (type: MealType) => {
    const newSet = new Set(selectedMealTypes);
    if (newSet.has(type)) {
      newSet.delete(type);
    } else {
      newSet.add(type);
    }
    setSelectedMealTypes(newSet);
  };

  const handleSelectAll = () => {
    if (isAllSelected) {
      setSelectedMealTypes(new Set());
    } else {
      setSelectedMealTypes(new Set(ALL_MEAL_TYPES.map((t) => t.id)));
    }
  };

  const handleSubmit = async () => {
    if (!startDate || !endDate || !reason.trim() || selectedMealTypes.size === 0) return;

    try {
      setIsSubmitting(true);
      const finalMealTypes = isAllSelected ? ['all'] : Array.from(selectedMealTypes);
      await onConfirm(startDate, endDate, reason, finalMealTypes);
      onClose();
    } catch (error) {
      console.error("Failed to submit bulk opt-out", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div 
        className="absolute inset-0 bg-[var(--color-background-light)]/50 backdrop-blur-md transition-opacity" 
        onClick={onClose}
      ></div>
      
      <div 
        className="relative w-full max-w-2xl bg-[#FFFDF5] rounded-3xl shadow-[var(--shadow-clay-modal)] border border-white/60 p-8 transform transition-all scale-100 overflow-hidden animate-in fade-in zoom-in duration-200"
        role="dialog"
        aria-modal="true"
      >
        {/* Decorative Background Element */}
        <div className="absolute top-0 right-0 w-48 h-48 bg-gradient-to-bl from-[var(--color-primary)]/5 to-transparent rounded-bl-[120px] pointer-events-none"></div>

        {/* Header */}
        <div className="flex items-center justify-between mb-6 relative">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-[#FF7E5F] to-[#FEB47B] text-white flex items-center justify-center shadow-[var(--shadow-clay-sm)]">
              <span className="material-symbols-outlined text-2xl">check_circle</span>
            </div>
            <div>
              <h2 className="text-2xl font-black text-[#23170f] tracking-tight">
                Confirm Bulk Opt-out
              </h2>
              <p className="text-sm text-[var(--color-text-sub)] font-medium">
                Please review the details before executing.
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="w-10 h-10 rounded-xl bg-transparent hover:bg-black/5 flex items-center justify-center text-[var(--color-text-sub)] transition-colors"
          >
            <span className="material-symbols-outlined">close</span>
          </button>
        </div>

        <div className="space-y-6">
          {/* Date Range Selection - Styled like "Opt-out Period" card */}
          <div className="bg-white/40 rounded-2xl p-5 shadow-[inset_6px_6px_12px_#e6dccf,inset_-6px_-6px_12px_#ffffff] border border-white/50">
            <div className="flex items-center justify-between mb-4">
              <span className="text-xs font-bold text-[var(--color-text-sub)] uppercase tracking-wider">
                Opt-out Period
              </span>
              <span className="text-xs font-bold text-[var(--color-primary)] bg-[var(--color-primary)]/10 px-2 py-1 rounded-lg">
                {durationDays > 0 ? `${durationDays} Days` : "Invalid Range"}
              </span>
            </div>
            
            <div className="flex flex-col md:flex-row items-center gap-4 text-[#23170f]">
              <div className="flex-1 w-full">
                 <label htmlFor="startDate" className="text-xs text-[var(--color-text-sub)] block mb-1">Start Date</label>
                 <input
                    type="date"
                    id="startDate"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="w-full bg-transparent font-bold text-lg focus:outline-none border-b border-transparent focus:border-[var(--color-primary)] transition-colors"
                 />
              </div>
              <span className="material-symbols-outlined text-[var(--color-text-sub)]/50 hidden md:block">arrow_forward</span>
              <div className="flex-1 w-full">
                 <label htmlFor="endDate" className="text-xs text-[var(--color-text-sub)] block mb-1">End Date</label>
                 <input
                    type="date"
                    id="endDate"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    min={startDate}
                    className="w-full bg-transparent font-bold text-lg focus:outline-none border-b border-transparent focus:border-[var(--color-primary)] transition-colors"
                 />
              </div>
            </div>
          </div>

          {/* Meal Selection */}
          <div>
            <div className="flex justify-between items-center mb-3">
              <label className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider">
                Meal Selection
              </label>
              <div 
                className="flex items-center gap-2 cursor-pointer"
                onClick={handleSelectAll}
              >
                <span className="text-xs font-bold text-[var(--color-text-sub)]">Select All</span>
                 {/* Custom Toggle Checkbox Style */}
                <div className={`
                    w-11 h-6 rounded-full relative transition-all duration-300 shadow-[inset_3px_3px_6px_#d1c0b0,inset_-3px_-3px_6px_#ffffff]
                    ${isAllSelected ? "bg-[#FFFDF5]" : "bg-[#FFFDF5]"}
                `}>
                    <div className={`
                        absolute top-0.5 left-0.5 w-5 h-5 rounded-full shadow-sm transition-all duration-300
                        ${isAllSelected ? "translate-x-5 bg-[var(--color-primary)]" : "translate-x-0 bg-gray-300"}
                    `}></div>
                </div>
              </div>
            </div>
            
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
              {ALL_MEAL_TYPES.map((meal) => {
                const isSelected = selectedMealTypes.has(meal.id);
                return (
                  <label key={meal.id} className="cursor-pointer group relative">
                    <input 
                        type="checkbox" 
                        className="peer sr-only"
                        checked={isSelected}
                        onChange={() => handleToggleMealType(meal.id)}
                    />
                    <div className={`
                        p-3 rounded-xl border transition-all flex items-center gap-2
                        ${isSelected 
                            ? "bg-orange-50 border-orange-200 shadow-none" 
                            : "bg-[#FFFDF5] border-transparent shadow-[var(--shadow-clay-sm)] hover:shadow-[var(--shadow-clay-md)]"
                        }
                    `}>
                      <span className={`material-symbols-outlined transition-colors ${isSelected ? "text-[var(--color-primary)]" : "text-[var(--color-text-sub)]"}`}>
                        {meal.icon}
                      </span>
                      <span className={`text-sm font-bold transition-colors ${isSelected ? "text-[#e57a36]" : "text-[var(--color-text-main)]"}`}>
                        {meal.label}
                      </span>
                    </div>
                    {/* Inner shadow effect when checked - simulated via inset box-shadow details in CSS if needed, or simple background change */}
                     {isSelected && (
                        <div className="absolute inset-0 rounded-xl pointer-events-none shadow-[inset_2px_2px_4px_rgba(0,0,0,0.05)]"></div>
                     )}
                  </label>
                );
              })}
            </div>
          </div>

          {/* Reason & Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-[#FFFDF5] p-4 rounded-2xl shadow-[inset_6px_6px_12px_#e6dccf,inset_-6px_-6px_12px_#ffffff]">
              <span className="block text-xs font-bold text-[var(--color-text-sub)] mb-2 uppercase">
                Reason Required
              </span>
              <textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="e.g., Team offsite event..."
                rows={2}
                className="w-full bg-transparent text-sm text-[#23170f] font-medium leading-relaxed italic focus:outline-none resize-none placeholder:text-gray-400 placeholder:not-italic"
              />
            </div>

            <div className="flex flex-col justify-center items-center bg-white/40 p-4 rounded-2xl border border-white shadow-sm">
              <span className="text-xs font-bold text-[var(--color-text-sub)] mb-1 uppercase">
                Total Users Affected
              </span>
              <div className="flex items-baseline gap-1">
                <span className="text-4xl font-black text-[#23170f]">
                  {selectedCount}
                </span>
                <span className="text-sm font-bold text-[var(--color-text-sub)]">Users</span>
              </div>
               {/* Optional: Add user avatars preview if data available, for now omitting based on props */}
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="mt-8 flex gap-4 pt-6 border-t border-[#e6dccf]/60">
            <button
              onClick={onClose}
              className="flex-1 px-6 py-3.5 rounded-2xl bg-[var(--color-background-light)] text-[var(--color-text-main)] font-bold shadow-[var(--shadow-clay-button)] hover:text-red-500 active:shadow-[var(--shadow-clay-button-active)] hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={isSubmitting || !startDate || !endDate || !reason || selectedMealTypes.size === 0}
              className="flex-[2] px-8 py-3.5 rounded-2xl bg-gradient-to-r from-[#FF7E5F] to-[#FEB47B] text-white font-bold shadow-[var(--shadow-clay-button)] hover:scale-[1.02] active:scale-[0.98] active:shadow-[var(--shadow-clay-button-active)] transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
               {isSubmitting ? (
                 <>
                   <span className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                   Processing...
                 </>
               ) : (
                 <>
                   <span className="material-symbols-outlined">bolt</span>
                   Confirm & Execute
                 </>
               )}
            </button>
        </div>

      </div>
    </div>
  );
};
