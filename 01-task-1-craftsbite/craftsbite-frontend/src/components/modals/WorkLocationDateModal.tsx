import React, { useEffect, useState } from "react";
import type { WorkLocation } from "../../services/workLocationService";

export interface WorkLocationDateModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (date: string, location: WorkLocation) => Promise<void>;
  currentLocation?: WorkLocation | null;
}

export const WorkLocationDateModal: React.FC<WorkLocationDateModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  currentLocation,
}) => {
  const [selectedDate, setSelectedDate] = useState("");
  const [selectedLocation, setSelectedLocation] = useState<WorkLocation | null>(
    currentLocation || null
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Get today's date in YYYY-MM-DD format for min attribute
  const today = new Date().toISOString().slice(0, 10);

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setSelectedDate("");
      setSelectedLocation(currentLocation || null);
      setIsSubmitting(false);
    }
  }, [isOpen, currentLocation]);

  // Close modal on Escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen && !isSubmitting) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleEscape);
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.body.style.overflow = "unset";
    };
  }, [isOpen, isSubmitting, onClose]);

  const handleSubmit = async () => {
    if (!selectedDate || !selectedLocation || isSubmitting) return;

    setIsSubmitting(true);
    try {
      await onConfirm(selectedDate, selectedLocation);
      onClose();
    } catch (error) {
      // Error handling is done in parent component
      console.error("Error updating work location:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget && !isSubmitting) onClose();
      }}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-[var(--color-background-light)]/40 backdrop-blur-md transition-opacity" />

      {/* Modal */}
      <div
        className="relative w-full max-w-md bg-[var(--color-background-light)] rounded-[2.5rem] transform transition-all border border-white/40 p-1"
        style={{ boxShadow: "var(--shadow-clay-modal)" }}
      >
        <div className="p-8">
          {/* Header */}
          <div className="flex justify-between items-start mb-6">
            <div>
              <h3 className="text-2xl font-black text-[var(--color-background-dark)]">
                Schedule Work Location
              </h3>
              <p className="text-sm text-[var(--color-text-sub)] mt-1">
                Set your work location for a specific date.
              </p>
            </div>
            <button
              onClick={onClose}
              disabled={isSubmitting}
              className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] flex items-center justify-center text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              style={{ boxShadow: "var(--shadow-clay-button)" }}
              aria-label="Close modal"
            >
              <span className="material-symbols-outlined text-xl">close</span>
            </button>
          </div>

          {/* Content */}
          <div className="space-y-6">
            {/* Date Picker */}
            <div>
              <label
                htmlFor="work-location-date"
                className="block text-sm font-bold text-[var(--color-background-dark)] mb-2"
              >
                Select Date
              </label>
              <input
                id="work-location-date"
                type="date"
                min={today}
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                disabled={isSubmitting}
                className="w-full px-4 py-3 rounded-xl bg-white/40 border border-white/60 text-[var(--color-text-main)] font-medium focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ boxShadow: "var(--shadow-clay-inset)" }}
              />
            </div>

            {/* Work Location Selector */}
            <div>
              <label className="block text-sm font-bold text-[var(--color-background-dark)] mb-3">
                Work Location
              </label>
              <div className="flex gap-4">
                <button
                  type="button"
                  onClick={() => setSelectedLocation("office")}
                  disabled={isSubmitting}
                  className={`flex-1 flex items-center justify-center gap-3 px-6 py-4 rounded-2xl font-bold text-sm uppercase tracking-wider relative overflow-hidden transition-all duration-300 group disabled:opacity-50 disabled:cursor-not-allowed ${
                    selectedLocation === "office"
                      ? "bg-gradient-to-r from-[var(--color-primary)] to-[#fb923c] text-white shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                      : "bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                  }`}
                >
                  <span
                    className={`absolute inset-0 bg-white/20 transition-transform duration-300 ${
                      selectedLocation === "office"
                        ? "translate-y-full group-hover:translate-y-0"
                        : "opacity-0"
                    }`}
                  />
                  <span className="relative z-10 text-2xl drop-shadow-sm">
                    🏢
                  </span>
                  <span className="relative z-10">Office</span>
                  {selectedLocation === "office" && (
                    <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-white animate-pulse" />
                  )}
                </button>

                <button
                  type="button"
                  onClick={() => setSelectedLocation("wfh")}
                  disabled={isSubmitting}
                  className={`flex-1 flex items-center justify-center gap-3 px-6 py-4 rounded-2xl font-bold text-sm uppercase tracking-wider relative overflow-hidden transition-all duration-300 group disabled:opacity-50 disabled:cursor-not-allowed ${
                    selectedLocation === "wfh"
                      ? "bg-gradient-to-r from-[var(--color-primary)] to-[#fb923c] text-white shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                      : "bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                  }`}
                >
                  <span
                    className={`absolute inset-0 bg-white/20 transition-transform duration-300 ${
                      selectedLocation === "wfh"
                        ? "translate-y-full group-hover:translate-y-0"
                        : "opacity-0"
                    }`}
                  />
                  <span className="relative z-10 text-2xl drop-shadow-sm">
                    🏡
                  </span>
                  <span className="relative z-10">WFH</span>
                  {selectedLocation === "wfh" && (
                    <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-white animate-pulse" />
                  )}
                </button>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="mt-8 flex gap-4">
            <button
              onClick={onClose}
              disabled={isSubmitting}
              className="flex-1 py-3.5 rounded-2xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-bold text-sm tracking-wide transition-all hover:text-[var(--color-text-main)] disabled:opacity-50 disabled:cursor-not-allowed"
              style={{ boxShadow: "var(--shadow-clay-button)" }}
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={!selectedDate || !selectedLocation || isSubmitting}
              className="flex-1 py-3.5 rounded-2xl bg-gradient-to-r from-[var(--color-primary)] to-[#fb923c] text-white font-bold text-sm tracking-wide transition-all active:translate-y-[1px] disabled:opacity-50 disabled:cursor-not-allowed disabled:active:translate-y-0"
              style={{
                boxShadow: "6px 6px 12px #d1c0b0, -6px -6px 12px #ffffff",
              }}
            >
              {isSubmitting ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="material-symbols-outlined animate-spin text-base">
                    progress_activity
                  </span>
                  Updating...
                </span>
              ) : (
                "Update Location"
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkLocationDateModal;
