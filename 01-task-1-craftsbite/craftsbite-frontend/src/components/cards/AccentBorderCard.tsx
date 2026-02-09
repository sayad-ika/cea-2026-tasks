import React, { type ReactNode } from 'react';

export interface AccentBorderCardProps {
    title: string;
    badge?: string;
    badgeColor?: string;
    badgeBgColor?: string;
    avatarUrl?: string;
    avatarAlt?: string;
    showPulse?: boolean;
    pulseColor?: string;
    children: ReactNode;
    footer?: ReactNode;
    className?: string;
}

/**
 * AccentBorderCard component
 * Highlighted state for active items or featured content
 * Uses subtle border tints with optional pulse indicator
 */
export const AccentBorderCard: React.FC<AccentBorderCardProps> = ({
    title,
    badge,
    badgeColor = '#fa8c47',
    badgeBgColor = 'rgba(250, 140, 71, 0.1)',
    avatarUrl,
    avatarAlt = 'Avatar',
    showPulse = true,
    pulseColor = '#fa8c47',
    children,
    footer,
    className = '',
}) => {
    return (
        <div className={`relative bg-background-base rounded-3xl p-8 shadow-clay-card flex flex-col border-2 border-[#fa8c47]/20 ${className}`}>
            {/* Pulse Indicator */}
            {showPulse && (
                <div className="absolute top-4 right-4">
                    <span
                        className="w-3 h-3 rounded-full block animate-pulse"
                        style={{ backgroundColor: pulseColor }}
                    />
                </div>
            )}

            {/* Header with Avatar and Badge */}
            <div className="flex items-center gap-4 mb-6">
                {avatarUrl && (
                    <div className="w-14 h-14 rounded-full overflow-hidden shadow-clay-inset border-2 border-white">
                        <img alt={avatarAlt} className="w-full h-full object-cover" src={avatarUrl} />
                    </div>
                )}
                <div>
                    <h4 className="text-lg font-bold text-[#23170f]">{title}</h4>
                    {badge && (
                        <span
                            className="text-xs font-bold px-2 py-1 rounded-md inline-block mt-1"
                            style={{ color: badgeColor, backgroundColor: badgeBgColor }}
                        >
                            {badge}
                        </span>
                    )}
                </div>
            </div>

            {/* Content Section */}
            <div className="flex-1">{children}</div>

            {/* Footer Section */}
            {footer && <div className="mt-auto pt-2">{footer}</div>}
        </div>
    );
};
