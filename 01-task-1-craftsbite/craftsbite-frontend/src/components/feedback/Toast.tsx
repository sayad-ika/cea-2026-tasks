import React from 'react';

export type ToastVariant = 'success' | 'error' | 'info';

export interface ToastProps {
    variant: ToastVariant;
    title: string;
    message: string;
    onClose?: () => void;
}

const variantConfig = {
    success: {
        borderColor: 'border-green-500',
        bgGradient: 'from-green-50',
        iconBg: 'bg-green-100',
        iconColor: 'text-green-600',
        icon: 'check_circle',
    },
    error: {
        borderColor: 'border-red-500',
        bgGradient: 'from-red-50',
        iconBg: 'bg-red-100',
        iconColor: 'text-red-600',
        icon: 'error',
    },
    info: {
        borderColor: 'border-blue-500',
        bgGradient: 'from-blue-50',
        iconBg: 'bg-blue-100',
        iconColor: 'text-blue-600',
        icon: 'info',
    },
};

export const Toast: React.FC<ToastProps> = ({ variant, title, message, onClose }) => {
    const config = variantConfig[variant];

    return (
        <div
            className={`pointer-events-auto flex items-start gap-3 bg-white/90 backdrop-blur-sm p-4 rounded-2xl border-l-4 ${config.borderColor} overflow-hidden relative`}
            style={{
                boxShadow: '10px 10px 30px rgba(230, 220, 207, 0.9), -5px -5px 15px rgba(255, 255, 255, 0.9)',
            }}
        >
            {/* Background Gradient */}
            <div
                className={`absolute inset-0 bg-gradient-to-r ${config.bgGradient} to-transparent opacity-50 pointer-events-none`}
            />

            {/* Icon */}
            <div
                className={`w-8 h-8 rounded-full ${config.iconBg} ${config.iconColor} flex items-center justify-center shrink-0 z-10 shadow-sm`}
            >
                <span className="material-symbols-outlined text-lg">{config.icon}</span>
            </div>

            {/* Content */}
            <div className="flex-1 z-10">
                <h4 className="font-bold text-gray-800 text-sm">{title}</h4>
                <p className="text-xs text-gray-500 mt-1 leading-relaxed">{message}</p>
            </div>

            {/* Close Button */}
            {onClose && (
                <button
                    onClick={onClose}
                    className="text-gray-400 hover:text-gray-600 transition-colors z-10"
                    aria-label="Close toast"
                >
                    <span className="material-symbols-outlined text-lg">close</span>
                </button>
            )}
        </div>
    );
};

export interface ToastContainerProps {
    children: React.ReactNode;
}

export const ToastContainer: React.FC<ToastContainerProps> = ({ children }) => {
    return (
        <div className="fixed top-24 right-4 z-50 flex flex-col gap-4 w-full max-w-sm pointer-events-none pr-4 md:pr-0">
            {children}
        </div>
    );
};

export default Toast;
