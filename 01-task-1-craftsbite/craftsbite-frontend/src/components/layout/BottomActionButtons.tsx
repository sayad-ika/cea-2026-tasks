import React from 'react';

export interface ActionButton {
    id: string;
    label: string;
    icon: string;
    onClick?: () => void;
}

export interface BottomActionButtonsProps {
    buttons?: ActionButton[];
    className?: string;
}

const defaultButtons: ActionButton[] = [
    {
        id: 'calendar',
        label: 'View Calendar',
        icon: 'calendar_month',
    },
    {
        id: 'history',
        label: 'History',
        icon: 'history',
    },
    {
        id: 'preferences',
        label: 'Preferences',
        icon: 'settings',
    },
];

/**
 * BottomActionButtons component
 * Displays a row of clay-styled action buttons with icons
 * Features hover animations and claymorphism design
 */
export const BottomActionButtons: React.FC<BottomActionButtonsProps> = ({
    buttons = defaultButtons,
    className = '',
}) => {
    return (
        <div
            className={`flex flex-col md:flex-row justify-center items-center gap-6 pb-8 ${className}`}
        >
            {buttons.map((button) => (
                <button
                    key={button.id}
                    onClick={button.onClick}
                    className="flex items-center gap-3 px-8 py-4 bg-background-light rounded-2xl shadow-clay-button active:shadow-clay-button-active text-text-main hover:text-primary transition-all duration-200 min-w-[200px] justify-center group hover:scale-105"
                    aria-label={button.label}
                >
                    <span className="material-symbols-outlined group-hover:scale-110 transition-transform duration-200">
                        {button.icon}
                    </span>
                    <span className="font-bold text-sm uppercase tracking-wider">
                        {button.label}
                    </span>
                </button>
            ))}
        </div>
    );
};
