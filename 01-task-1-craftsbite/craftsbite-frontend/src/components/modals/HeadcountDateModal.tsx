import { useState, useEffect } from "react";
import { getTodayString } from "../../utils";

export interface HeadcountDateModalProps {
  isOpen: boolean;
  title: string;
  submitLabel: string;
  submitIcon: string;
  isSubmitting: boolean;
  onClose: () => void;
  onSubmit: (date: string) => void;
}

export const HeadcountDateModal: React.FC<HeadcountDateModalProps> = ({
  isOpen,
  title,
  submitLabel,
  submitIcon,
  isSubmitting,
  onClose,
  onSubmit,
}) => {
  const [pickedDate, setPickedDate] = useState(getTodayString());

  useEffect(() => {
    if (isOpen) setPickedDate(getTodayString());
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/30 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-full max-w-sm bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 flex flex-col gap-6"
        style={{ boxShadow: "var(--shadow-clay-card)" }}
      >
        {/* Header */}
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-black tracking-tight text-[var(--color-background-dark)]">
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
          >
            <span className="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>

        {/* Date field */}
        <div className="flex flex-col gap-1.5">
          <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
            Date
          </label>
          <input
            type="date"
            value={pickedDate}
            onChange={(e) => setPickedDate(e.target.value)}
            className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
          />
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
            type="button"
            disabled={!pickedDate || isSubmitting}
            onClick={() => pickedDate && onSubmit(pickedDate)}
            className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
          >
            {isSubmitting ? (
              <>
                <span className="material-symbols-outlined text-[16px] animate-spin">
                  progress_activity
                </span>
                Loading…
              </>
            ) : (
              <>
                <span className="material-symbols-outlined text-[16px]">{submitIcon}</span>
                {submitLabel}
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
