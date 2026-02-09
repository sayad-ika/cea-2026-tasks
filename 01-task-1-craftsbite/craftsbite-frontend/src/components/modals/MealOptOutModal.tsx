import React, { useEffect } from 'react';

export interface MealOptOutModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    mealName?: string;
    cutoffTime?: string;
}

export const MealOptOutModal: React.FC<MealOptOutModalProps> = ({
    isOpen,
    onClose,
    onConfirm,
    mealName = 'lunch',
    cutoffTime = '10:30 AM',
}) => {
    // Close modal on Escape key
    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) {
                onClose();
            }
        };

        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }

        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            {/* Backdrop - darker with blur */}
            <div
                className="absolute inset-0 bg-[var(--color-background-dark)]/20 backdrop-blur-sm transition-opacity"
                onClick={onClose}
            />

            {/* Modal */}
            <div
                className="relative w-full max-w-sm bg-[var(--color-background-light)] rounded-3xl p-8 transform transition-all scale-100 border border-white/40"
                style={{ boxShadow: '30px 30px 70px #d4c8b8, -30px -30px 70px #ffffff' }}
            >
                {/* Close Button */}
                <button
                    onClick={onClose}
                    className="absolute top-4 right-4 w-8 h-8 rounded-full flex items-center justify-center text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors"
                    aria-label="Close modal"
                >
                    <span className="material-symbols-outlined text-xl">close</span>
                </button>

                {/* Content */}
                <div className="flex flex-col items-center text-center">
                    {/* Warning Icon */}
                    <div
                        className="w-16 h-16 rounded-2xl bg-[#FFE6E6] flex items-center justify-center text-red-500 mb-6"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <span className="material-symbols-outlined text-[32px] font-bold">warning</span>
                    </div>

                    {/* Title */}
                    <h3 className="text-2xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
                        Cancel {mealName.charAt(0).toUpperCase() + mealName.slice(1)}?
                    </h3>

                    {/* Description */}
                    <p className="text-[var(--color-text-sub)] font-medium mb-8 leading-relaxed">
                        Are you sure you want to opt out of today's {mealName}? This action cannot be undone
                        after {cutoffTime}.
                    </p>

                    {/* Action Buttons */}
                    <div className="flex gap-4 w-full">
                        {/* Keep Button */}
                        <button
                            onClick={onClose}
                            className="flex-1 py-3.5 px-4 bg-[var(--color-background-light)] rounded-xl text-[var(--color-text-main)] font-bold hover:text-[var(--color-primary)] transition-all duration-200 border border-transparent"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        >
                            Keep it
                        </button>

                        {/* Cancel/Confirm Button */}
                        <button
                            onClick={onConfirm}
                            className="flex-1 py-3.5 px-4 bg-red-500 rounded-xl text-white font-bold hover:bg-red-600 transition-all duration-200"
                            style={{
                                boxShadow: '8px 8px 16px #fca5a5, -8px -8px 16px #ffffff',
                            }}
                        >
                            Yes, Cancel
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default MealOptOutModal;
