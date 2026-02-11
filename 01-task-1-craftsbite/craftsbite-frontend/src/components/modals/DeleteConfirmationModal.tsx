import React from "react";

interface DeleteConfirmationModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    isDeleting: boolean;
    title: string;
    message: string;
}

export const DeleteConfirmationModal: React.FC<DeleteConfirmationModalProps> = ({
    isOpen,
    onClose,
    onConfirm,
    isDeleting,
    title,
    message,
}) => {
    if (!isOpen) return null;

    return (
        <div
            className="fixed inset-0 z-[100] flex items-center justify-center p-4"
            onClick={(e) => {
                if (e.target === e.currentTarget) onClose();
            }}
        >
            {/* Backdrop */}
            <div className="absolute inset-0 bg-[#23170f]/20 backdrop-blur-sm transition-opacity" />

            {/* Modal */}
            <div className="relative w-full max-w-sm bg-[#FFF5E6] rounded-3xl p-8 shadow-[30px_30px_70px_#d4c8b8,-30px_-30px_70px_#ffffff] transform transition-all scale-100 border border-white/40">
                {/* Close button */}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={isDeleting}
                    className="absolute top-4 right-4 w-8 h-8 rounded-full flex items-center justify-center text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors disabled:opacity-50"
                >
                    <span className="material-symbols-outlined text-xl">close</span>
                </button>

                <div className="flex flex-col items-center text-center">
                    {/* Warning icon */}
                    <div className="w-16 h-16 rounded-2xl bg-[#FFE6E6] shadow-[var(--shadow-clay-inset)] flex items-center justify-center text-red-500 mb-6">
                        <span className="material-symbols-outlined text-[32px] font-bold">warning</span>
                    </div>

                    {/* Title */}
                    <h3 className="text-2xl font-black text-[#23170f] mb-2 tracking-tight">{title}</h3>

                    {/* Message */}
                    <p className="text-[var(--color-text-sub)] font-medium mb-8 leading-relaxed">
                        {message}
                    </p>

                    {/* Action Buttons */}
                    <div className="flex gap-4 w-full">
                        <button
                            type="button"
                            onClick={onClose}
                            disabled={isDeleting}
                            className="flex-1 py-3.5 px-4 bg-[var(--color-background-light)] rounded-xl shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)] text-[var(--color-text-main)] font-bold hover:text-[var(--color-primary)] transition-all duration-200 border border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            onClick={onConfirm}
                            disabled={isDeleting}
                            className="flex-1 py-3.5 px-4 bg-red-500 rounded-xl shadow-[8px_8px_16px_#fca5a5,-8px_-8px_16px_#ffffff] active:shadow-[inset_4px_4px_8px_#b91c1c,inset_-4px_-4px_8px_#ef4444] text-white font-bold hover:bg-red-600 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {isDeleting ? "Deleting..." : "Yes, Delete"}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};
