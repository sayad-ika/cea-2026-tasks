import React from 'react';
import { useTheme } from '../../contexts/ThemeContext';

export interface HeaderProps {
    userName?: string;
    userRole?: string;
    userAvatarUrl?: string;
    onThemeToggle?: () => void;
    isDarkMode?: boolean;
}

/**
 * Header component for CraftsBite application
 * Displays logo, user information, theme toggle, and profile avatar
 */
export const Header: React.FC<HeaderProps> = ({
    userName = 'John Doe',
    userRole = 'UX Designer',
    userAvatarUrl = 'https://lh3.googleusercontent.com/aida-public/AB6AXuDttB-Zd88STJy72M0f6YLArBV0Tl6JcqsbvalmTAy_8AywOe4phVL02tIDhmBzF_0en7PxQymXKbuWs2ZmYTeA5h17TdH_T0CECPuSQMs_7Cuslryyjv-n7bG8lVMl9tZ9EePyF-WJQamvjji2HePEm22UcTO3MIOqv41gtGMGEXGZtE6lDv86hD3Y70w-RcQCy8mfewQdK4dEh4csDJ4TTzjE1Y3UdYxTEheOhJC3ApTbqT-BpQDOsIb0KgRR7XtrQHCNGpthJvHv',
    onThemeToggle: externalToggle,
    isDarkMode: externalDark,
}) => {
    const theme = useTheme();
    const isDark = externalDark ?? theme.isDarkMode;
    const handleToggle = externalToggle ?? theme.toggleTheme;

    return (
        <header className="w-full px-6 py-4 md:px-12 flex justify-between items-center z-10">
            {/* Logo Section */}
            <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-[var(--color-primary)] text-white flex items-center justify-center"
                    style={{ boxShadow: 'var(--shadow-clay-button)' }}>
                    <span className="material-symbols-outlined">restaurant</span>
                </div>
                <h1 className="text-2xl font-black tracking-tight text-[var(--color-background-dark)]">
                    CraftsBite
                </h1>
            </div>

            {/* User Info & Actions Section */}
            <div className="flex items-center gap-4">
                {/* User Info - Hidden on mobile */}
                <div className="hidden md:flex flex-col items-end mr-4">
                    <span className="text-sm font-bold text-[var(--color-background-dark)]">{userName}</span>
                    <span className="text-xs text-[var(--color-text-sub)]">{userRole}</span>
                </div>

                {/* Theme Toggle Button */}
                <button
                    onClick={handleToggle}
                    className="w-12 h-12 rounded-xl bg-[var(--color-background-light)] flex items-center justify-center text-[var(--color-primary)] transition-all duration-200 hover:scale-105"
                    style={{ boxShadow: 'var(--shadow-clay-button)' }}
                    aria-label="Toggle theme"
                >
                    <span className="material-symbols-outlined">
                        {isDark ? 'dark_mode' : 'light_mode'}
                    </span>
                </button>

                {/* User Avatar */}
                <div className="w-12 h-12 rounded-full overflow-hidden border-2 border-[var(--color-clay-light)] cursor-pointer hover:scale-105 transition-transform duration-200"
                    style={{ boxShadow: 'var(--shadow-clay-button)' }}>
                    <img
                        alt={`${userName} profile avatar`}
                        className="w-full h-full object-cover"
                        src={userAvatarUrl}
                    />
                </div>
            </div>
        </header>
    );
};
