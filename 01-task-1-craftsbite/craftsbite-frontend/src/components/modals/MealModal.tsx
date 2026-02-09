import React, { useEffect } from 'react';

export interface MealModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    mealName: string;
    mealEmoji: string;
    mealTime: string;
    mealBackgroundColor: string;
    notifyKitchen?: boolean;
    onNotifyKitchenChange?: (checked: boolean) => void;
}

export const MealModal: React.FC<MealModalProps> = ({
    isOpen,
    onClose,
    onConfirm,
    mealName,
    mealEmoji,
    mealTime,
    mealBackgroundColor,
    notifyKitchen = false,
    onNotifyKitchenChange,
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
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-[var(--color-background-light)]/40 backdrop-blur-md transition-opacity"
                onClick={onClose}
            />

            {/* Modal */}
            <div
                className="relative w-full max-w-md bg-[var(--color-background-light)] rounded-[2.5rem] transform transition-all border border-white/40 p-1"
                style={{ boxShadow: '30px 30px 70px #d4c9bc, -30px -30px 70px #ffffff' }}
            >
                <div className="p-8">
                    {/* Header */}
                    <div className="flex justify-between items-start mb-6">
                        <div>
                            <h3 className="text-2xl font-black text-[var(--color-background-dark)]">
                                Change Preference
                            </h3>
                            <p className="text-sm text-[var(--color-text-sub)] mt-1">
                                Update your dietary settings for today.
                            </p>
                        </div>
                        <button
                            onClick={onClose}
                            className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] flex items-center justify-center text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-all duration-200"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}
                            aria-label="Close modal"
                        >
                            <span className="material-symbols-outlined text-xl">close</span>
                        </button>
                    </div>

                    {/* Content */}
                    <div className="space-y-6">
                        {/* Meal Info Card */}
                        <div className="bg-white/40 rounded-2xl p-4 border border-white/60 shadow-inner">
                            <div className="flex items-center gap-4 mb-2">
                                <div
                                    className="w-12 h-12 rounded-full flex items-center justify-center text-2xl shadow-sm"
                                    style={{ backgroundColor: mealBackgroundColor }}
                                >
                                    {mealEmoji}
                                </div>
                                <div>
                                    <p className="font-bold text-[var(--color-background-dark)]">
                                        {mealName} Menu
                                    </p>
                                    <p className="text-xs text-[var(--color-text-sub)]">
                                        Today • {mealTime}
                                    </p>
                                </div>
                            </div>
                            <p className="text-sm text-[var(--color-text-main)] pl-1">
                                Are you sure you want to opt-out of today's {mealName.toLowerCase()}? This
                                action cannot be undone after 10:30 AM.
                            </p>
                        </div>

                        {/* Checkbox Option */}
                        <div className="flex flex-col gap-3">
                            <label className="flex items-center gap-3 p-3 rounded-xl hover:bg-white/30 transition-colors cursor-pointer group">
                                <input
                                    type="checkbox"
                                    checked={notifyKitchen}
                                    onChange={(e) => onNotifyKitchenChange?.(e.target.checked)}
                                    className="sr-only peer"
                                />
                                <div className="w-5 h-5 rounded-md border-2 border-[var(--color-text-sub)] flex items-center justify-center group-hover:border-[var(--color-primary)] peer-checked:border-[var(--color-primary)] transition-colors">
                                    <div
                                        className={`w-2.5 h-2.5 rounded-sm bg-[var(--color-primary)] transition-opacity ${notifyKitchen ? 'opacity-100' : 'opacity-0'
                                            }`}
                                    />
                                </div>
                                <span className="text-sm font-medium text-[var(--color-text-main)] group-hover:text-[var(--color-background-dark)]">
                                    Notify kitchen about this change
                                </span>
                            </label>
                        </div>
                    </div>

                    {/* Action Buttons */}
                    <div className="mt-8 flex gap-4">
                        <button
                            onClick={onClose}
                            className="flex-1 py-3.5 rounded-2xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] font-bold text-sm tracking-wide transition-all hover:text-[var(--color-text-main)]"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        >
                            Cancel
                        </button>
                        <button
                            onClick={onConfirm}
                            className="flex-1 py-3.5 rounded-2xl bg-[var(--color-primary)] text-white font-bold text-sm tracking-wide transition-all active:translate-y-[1px]"
                            style={{
                                boxShadow: '6px 6px 12px #d1c0b0, -6px -6px 12px #ffffff',
                            }}
                        >
                            Confirm
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default MealModal;
