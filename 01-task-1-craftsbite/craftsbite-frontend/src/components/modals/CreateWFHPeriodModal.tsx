import React, { useState, useEffect } from "react";
import type { CreateWFHPeriodRequest } from "../../types/wfh-period.types";

export interface CreateWFHPeriodModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (payload: CreateWFHPeriodRequest) => Promise<void>;
}

export const CreateWFHPeriodModal: React.FC<CreateWFHPeriodModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
}) => {
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [reason, setReason] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isOpen) {
      setStartDate("");
      setEndDate("");
      setReason("");
      setError("");
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!startDate || !endDate) return;

    if (endDate < startDate) {
      setError("End date must be on or after start date.");
      return;
    }

    const payload: CreateWFHPeriodRequest = {
      start_date: startDate,
      end_date: endDate,
      ...(reason.trim() && { reason: reason.trim() }),
    };

    try {
      setIsSubmitting(true);
      setError("");
      await onSubmit(payload);
      onClose();
    } catch (err: any) {
      const errorMessage =
        err?.error?.message ||
        err?.message ||
        "Failed to create WFH period. Please try again.";
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
            Create WFH Period
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
                <p className="text-sm font-semibold text-red-600">Error</p>
                <p className="text-sm text-red-600 mt-1">{error}</p>
              </div>
            </div>
          )}

          <div className="flex flex-col gap-1.5">
            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
              Start Date
            </label>
            <input
              type="date"
              required
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
              End Date
            </label>
            <input
              type="date"
              required
              value={endDate}
              min={startDate || undefined}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
            />
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
              placeholder="e.g. Office renovation, team retreat…"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all placeholder:text-[var(--color-text-sub)]/50"
            />
          </div>

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
              disabled={!startDate || !endDate || isSubmitting}
              className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
            >
              {isSubmitting ? (
                <>
                  <span className="material-symbols-outlined text-[16px] animate-spin">
                    progress_activity
                  </span>{" "}
                  Creating…
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
