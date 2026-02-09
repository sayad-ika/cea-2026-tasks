import React from 'react';

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
    onThemeToggle,
    isDarkMode = false,
}) => {
    return (
        <header className="w-full px-6 py-4 md:px-12 flex justify-between items-center z-10">
            {/* Logo Section */}
            <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-[#fa8c47] text-white flex items-center justify-center shadow-clay-button">
                    <span className="material-symbols-outlined">restaurant</span>
                </div>
                <h1 className="text-2xl font-black tracking-tight text-[#23170f]">
                    CraftsBite
                </h1>
            </div>

            {/* User Info & Actions Section */}
            <div className="flex items-center gap-4">
                {/* User Info - Hidden on mobile */}
                <div className="hidden md:flex flex-col items-end mr-4">
                    <span className="text-sm font-bold text-[#23170f]">{userName}</span>
                    <span className="text-xs text-text-sub">{userRole}</span>
                </div>

                {/* Theme Toggle Button */}
                <button
                    onClick={onThemeToggle}
                    className="w-12 h-12 rounded-xl bg-background-light shadow-clay-button flex items-center justify-center text-primary active:shadow-clay-button-active transition-all duration-200 hover:scale-105"
                    aria-label="Toggle theme"
                >
                    <span className="material-symbols-outlined">
                        {isDarkMode ? 'dark_mode' : 'light_mode'}
                    </span>
                </button>

                {/* User Avatar */}
                <div className="w-12 h-12 rounded-full overflow-hidden shadow-clay-button border-2 border-white cursor-pointer hover:scale-105 transition-transform duration-200">
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
