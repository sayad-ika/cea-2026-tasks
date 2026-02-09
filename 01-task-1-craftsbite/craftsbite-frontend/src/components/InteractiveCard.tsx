import React, { type ReactNode } from 'react';

export interface InteractiveCardProps {
    icon: ReactNode;
    iconColor?: string;
    iconBgColor?: string;
    title: string;
    description: string;
    buttonLabel?: string;
    onButtonClick?: () => void;
    onClick?: () => void;
    className?: string;
}

/**
 * InteractiveCard component with lift animation on hover
 * Used for selection grids or navigation with claymorphism design
 */
export const InteractiveCard: React.FC<InteractiveCardProps> = ({
    icon,
    iconColor = '#fa8c47',
    iconBgColor = '#fff0e6',
    title,
    description,
    buttonLabel = 'View Menu',
    onButtonClick,
    onClick,
    className = '',
}) => {
    const handleCardClick = () => {
        onClick?.();
    };

    const handleButtonClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        onButtonClick?.();
    };

    return (
        <div
            onClick={handleCardClick}
            className={`group cursor-pointer bg-background-base rounded-3xl p-8 shadow-clay-card flex flex-col items-center text-center transition-all duration-300 hover:-translate-y-2 hover:shadow-clay-card-hover ${className}`}
        >
            {/* Icon Container */}
            <div
                className="w-16 h-16 rounded-2xl shadow-clay-button flex items-center justify-center mb-5 group-hover:scale-110 transition-transform duration-300"
                style={{ backgroundColor: iconBgColor, color: iconColor }}
            >
                {icon}
            </div>

            {/* Title */}
            <h4 className="text-xl font-bold text-[#23170f] mb-2">{title}</h4>

            {/* Description */}
            <p className="text-text-sub text-sm mb-6 leading-relaxed">{description}</p>

            {/* Action Button */}
            {buttonLabel && (
                <button
                    onClick={handleButtonClick}
                    className="px-6 py-2 rounded-xl bg-[#fa8c47] text-white font-bold text-sm shadow-clay-button active:shadow-inner transition-all hover:bg-[#e67e3c]"
                >
                    {buttonLabel}
                </button>
            )}
        </div>
    );
};
