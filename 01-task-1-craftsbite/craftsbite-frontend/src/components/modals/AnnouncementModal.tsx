import { useState, useCallback, useEffect } from "react";
import { getTodayString, formatDate } from "../../utils";

export interface AnnouncementModalProps {
  isOpen: boolean;
  date: string;
  message: string;
  onClose: () => void;
}

export const AnnouncementModal: React.FC<AnnouncementModalProps> = ({
  isOpen,
  date,
  message,
  onClose,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(message).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }, [message]);

  useEffect(() => {
    if (!isOpen) setCopied(false);
  }, [isOpen]);

  if (!isOpen) return null;

  const displayDate =
    date === getTodayString() ? "Today" : formatDate(date);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/30 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-full max-w-lg bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 flex flex-col gap-6"
        style={{ boxShadow: "var(--shadow-clay-card)" }}
      >
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-black tracking-tight text-[var(--color-background-dark)]">
              Daily Announcement
            </h2>
            <p className="text-xs text-[var(--color-text-sub)] font-medium mt-0.5">
              {displayDate}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="w-9 h-9 flex items-center justify-center rounded-xl text-[var(--color-text-sub)] hover:text-[var(--color-text-main)] hover:bg-[var(--color-background-light)] transition-all"
          >
            <span className="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>

        {/* Message box */}
        <div className="relative">
          <textarea
            readOnly
            value={message}
            rows={10}
            className="w-full px-4 py-3 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-mono leading-relaxed shadow-[var(--shadow-clay-button)] outline-none resize-none"
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
            Close
          </button>
          <button
            type="button"
            onClick={handleCopy}
            className={`px-5 py-2.5 rounded-xl text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 ${
              copied
                ? "bg-green-500 text-white"
                : "bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white"
            }`}
          >
            <span className="material-symbols-outlined text-[16px]">
              {copied ? "check" : "content_copy"}
            </span>
            {copied ? "Copied!" : "Copy"}
          </button>
        </div>
      </div>
    </div>
  );
};
