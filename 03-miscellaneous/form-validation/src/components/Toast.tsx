import { useEffect } from "react";
import type { ToastType } from "../type/toast";

interface ToastProps {
    message?: string;
    type?: ToastType;
    onClose: () => void;
}

const Toast = ({ message, type, onClose }: ToastProps) => {
    useEffect(() => {
        if (!type) return;

        const id = setTimeout(onClose, 3000);
        return () => clearTimeout(id);
    }, [type, onClose]);

    if (!type) return null;

    const colors =
        type === "success"
            ? "bg-green-600 text-white"
            : "bg-red-600 text-white";

    return (
        <div
            role="alert"
            aria-live="assertive"
            className={`fixed bottom-6 right-6 px-5 py-3 rounded-lg shadow-lg
                  flex items-center gap-3 ${colors} z-50`}
        >
            <span>{message}</span>
            <button
                onClick={onClose}
                aria-label="Dismiss notification"
                className="ml-auto font-bold text-lg leading-none"
            >
                ×
            </button>
        </div>
    );
};

export default Toast;
