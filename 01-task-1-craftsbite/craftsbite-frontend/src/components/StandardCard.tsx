import React, { type ReactNode } from 'react';

export interface StandardCardProps {
    title: string;
    subtitle?: string;
    icon?: ReactNode;
    iconBgColor?: string;
    iconColor?: string;
    children: ReactNode;
    className?: string;
}

/**
 * StandardCard component
 * Base card style for static content groups like summaries or informational panels
 * Uses claymorphism design with shadow-clay-card
 */
export const StandardCard: React.FC<StandardCardProps> = ({
    title,
    subtitle,
    icon,
    iconBgColor = '#f0f9ff',
    iconColor = '#3b82f6',
    children,
    className = '',
}) => {
    return (
        <div className={`bg-background-base rounded-3xl p-8 shadow-clay-card flex flex-col ${className}`}>
            {/* Header Section */}
            <div className="flex justify-between items-start mb-6">
                <div>
                    <h4 className="text-xl font-bold text-[#23170f]">{title}</h4>
                    {subtitle && <p className="text-text-sub text-sm">{subtitle}</p>}
                </div>
                {icon && (
                    <div
                        className="w-10 h-10 rounded-full shadow-clay-inset flex items-center justify-center"
                        style={{ backgroundColor: iconBgColor, color: iconColor }}
                    >
                        {icon}
                    </div>
                )}
            </div>

            {/* Content Section */}
            <div>{children}</div>
        </div>
    );
};
