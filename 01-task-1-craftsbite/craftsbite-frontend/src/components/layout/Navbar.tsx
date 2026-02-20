import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

export interface NavItem {
    id: string;
    label: string;
    icon: string;
    href: string;
}

/**
 * Navbar component with role-based tabs
 * Features smooth animations, hover effects, and active state indication
 */
export const Navbar: React.FC = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const { user } = useAuth();

    // Build nav items based on user role
    const navItems: NavItem[] = [
        { id: 'dashboard', label: 'Dashboard', icon: 'dashboard', href: '/home' },
        ...(user?.role === 'admin' || user?.role === 'logistics'
            ? [
                { id: 'headcount', label: 'Headcount', icon: 'groups', href: '/headcount' },
                { id: 'schedule', label: 'Schedule', icon: 'calendar_today', href: '/schedule' },
                { id: 'wfh-periods', label: 'WFH Periods', icon: 'date_range', href: '/wfh-periods' }
            ]
            : []),
        ...(user?.role === 'admin' || user?.role === 'team_lead'
            ? [
                { id: 'override', label: 'Override', icon: 'swap_horiz', href: '/override' }
            ]
            : []),
        ...(user?.role === 'admin' || user?.role === 'team_lead' || user?.role === 'logistics'
            ? [
                { id: 'team', label: 'Team', icon: 'people', href: '/team' },
            ]
            : []),
    ];

    // Determine active item from current path
    const activeItemId = navItems.find((i) => i.href === location.pathname)?.id || 'dashboard';

    const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, item: NavItem) => {
        e.preventDefault();
        navigate(item.href);
    };

    return (
        <div className="w-full mb-10 border-b border-[#e6dccf] relative px-12">
            <nav aria-label="Tabs" className="flex gap-8 overflow-x-auto no-scrollbar">
                {navItems.map((item) => {
                    const isActive = item.id === activeItemId;

                    return (
                        <a
                            key={item.id}
                            href={item.href}
                            onClick={(e) => handleClick(e, item)}
                            className={`group relative py-4 px-1 text-sm font-${isActive ? 'bold' : 'medium'
                                } ${isActive
                                    ? 'text-[#fa8c47]'
                                    : 'text-text-sub hover:text-[#23170f]'
                                } focus:outline-none min-w-fit transition-colors duration-200`}
                        >
                            {/* Hover Background */}
                            <span className="absolute inset-0 rounded-t-lg bg-[#fa8c47]/5 opacity-0 group-hover:opacity-100 transition-opacity duration-200" />

                            {/* Nav Item Content */}
                            <div className="flex items-center gap-2 relative z-10 px-3">
                                <span className="material-symbols-outlined text-[20px]">
                                    {item.icon}
                                </span>
                                {item.label}
                            </div>

                            {/* Active/Hover Underline */}
                            <span
                                className={`absolute bottom-0 left-0 w-full h-0.5 bg-[#fa8c47] rounded-t-full transition-transform duration-200 origin-left ${isActive
                                    ? ''
                                    : 'scale-x-0 group-hover:scale-x-100'
                                    }`}
                            />
                        </a>
                    );
                })}
            </nav>
        </div>
    );
};
