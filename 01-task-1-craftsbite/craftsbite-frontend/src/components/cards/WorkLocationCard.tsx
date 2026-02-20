import React, { useState, useEffect } from "react";
import * as workLocationService from "../../services/workLocationService";
import type { WorkLocationValue } from "../../types/work-location.types";
import toast from "react-hot-toast";

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    weekday: "long",
    month: "short",
    day: "numeric",
  });
}

function getTodayDateStr(): string {
  return new Date().toISOString().split("T")[0];
}

function getTomorrowDateStr(): string {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toISOString().split("T")[0];
}

interface SetWorkLocationModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (date: string, location: WorkLocationValue) => void;
}

const SetWorkLocationModal: React.FC<SetWorkLocationModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [date, setDate] = useState(getTomorrowDateStr());
  const [location, setLocation] = useState<"office" | "wfh">("office");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isOpen) {
      setDate(getTomorrowDateStr());
      setLocation("office");
      setError("");
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!date) return;

    try {
      setIsSubmitting(true);
      setError("");
      await workLocationService.setWorkLocation({ date, location });
      toast.success(
        location === "office"
          ? `Set to Office for ${formatDate(date)}`
          : `Set to WFH for ${formatDate(date)}`
      );
      onSuccess(date, location);
      onClose();
    } catch (err: any) {
      const msg = err?.error?.message || "Failed to set work location.";
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
            Set Work Location
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
          {/* Error */}
          {error && (
            <div className="flex items-center gap-3 p-4 rounded-xl bg-red-50 border border-red-200">
              <div>
                <p className="text-sm font-semibold text-red-600">Error</p>
                <p className="text-sm text-red-600 mt-1">{error}</p>
              </div>
            </div>
          )}

          {/* Date field */}
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

          {/* Location toggle — 2-column grid like Day Status in CreateScheduleModal */}
          <div className="flex flex-col gap-1.5">
            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
              Location
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setLocation("office")}
                className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${
                  location === "office"
                    ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white shadow-[var(--shadow-clay-button)]"
                    : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                }`}
              >
                <span className="material-symbols-outlined text-[16px]">
                  apartment
                </span>
                Office
              </button>
              <button
                type="button"
                onClick={() => setLocation("wfh")}
                className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${
                  location === "wfh"
                    ? "bg-[var(--color-background-dark)] border-[var(--color-background-dark)] text-white shadow-[var(--shadow-clay-button)]"
                    : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)]"
                }`}
              >
                <span className="material-symbols-outlined text-[16px]">
                  home
                </span>
                WFH
              </button>
            </div>
          </div>

          {/* Divider */}
          <div className="border-t border-[#e6dccf]" />

          {/* Action buttons */}
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
                    check
                  </span>{" "}
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

export const WorkLocationCard: React.FC = () => {
  const [currentLocation, setCurrentLocation] = useState<WorkLocationValue>("not_set");
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    const fetch = async () => {
      try {
        const res = await workLocationService.getWorkLocation(
          getTodayDateStr()
        );
        if (res.data) {
          setCurrentLocation(res.data.location);
        }
      } catch (err) {
        console.error("Error fetching work location:", err);
      }
    };
    fetch();
  }, []);

  const handleSuccess = (date: string, location: WorkLocationValue) => {
    if (date === getTodayDateStr()) {
      setCurrentLocation(location);
    }
  };

  // Badge helper
  const locationLabel =
    currentLocation === "office"
      ? "Office"
      : currentLocation === "wfh"
        ? "WFH"
        : "Not Set";

  const locationBadgeClass =
    currentLocation === "not_set"
      ? "bg-gray-100 text-gray-500 border-gray-200"
      : "bg-gradient-to-r from-orange-50 to-orange-100 text-orange-700 border-orange-200";

  return (
    <>
      <div className="mb-6 bg-[var(--color-background-light)] rounded-3xl p-6 md:p-8 flex flex-col md:flex-row items-center justify-between gap-6 border-2 border-white/50 shadow-[var(--shadow-clay)]">

        <div className="w-full md:w-auto">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-full bg-[var(--color-primary)]/10 flex items-center justify-center text-[var(--color-primary)] shadow-[var(--shadow-clay-inset)]">
              <span className="material-symbols-outlined text-2xl">
                location_on
              </span>
            </div>
            <div>
              <h3 className="text-xl font-bold text-left text-[var(--color-background-dark)]">
                Work Location
              </h3>
              <p className="text-sm text-[var(--color-text-sub)]">
                Today's status
              </p>
            </div>
          </div>

          <div
            className={`inline-flex items-center mt-3 ml-1 px-3 py-1 rounded-xl border text-xs font-bold ${locationBadgeClass}`}
          >
            {locationLabel}
          </div>
        </div>

        <button
          type="button"
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-6 py-3 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-sm hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
          style={{ boxShadow: "var(--shadow-clay-button)" }}
        >
          <span className="material-symbols-outlined text-lg">edit_location_alt</span>
          Set Work Location
        </button>
      </div>

      <SetWorkLocationModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={handleSuccess}
      />
    </>
  );
};
