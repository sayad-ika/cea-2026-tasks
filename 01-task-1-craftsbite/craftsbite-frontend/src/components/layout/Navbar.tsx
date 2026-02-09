import React from 'react';

export interface NavItem {
    id: string;
    label: string;
    icon: string;
    href: string;
}

export interface NavbarProps {
    items?: NavItem[];
    activeItemId?: string;
    onNavItemClick?: (itemId: string) => void;
}

const defaultNavItems: NavItem[] = [
    {
        id: 'dashboard',
        label: 'Dashboard',
        icon: 'dashboard',
        href: '#',
    },
    {
        id: 'my-meals',
        label: 'My Meals',
        icon: 'restaurant_menu',
        href: '#',
    },
    {
        id: 'admin-panel',
        label: 'Admin Panel',
        icon: 'admin_panel_settings',
        href: '#',
    },
];

/**
 * Navbar component with tabs
 * Features smooth animations, hover effects, and active state indication
 */
export const Navbar: React.FC<NavbarProps> = ({
    items = defaultNavItems,
    activeItemId = 'dashboard',
    onNavItemClick,
}) => {
    const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, itemId: string) => {
        e.preventDefault();
        onNavItemClick?.(itemId);
    };

    return (
        <div className="w-full mb-10 border-b border-[#e6dccf] relative">
            <nav aria-label="Tabs" className="flex gap-8 overflow-x-auto no-scrollbar">
                {items.map((item) => {
                    const isActive = item.id === activeItemId;

                    return (
                        <a
                            key={item.id}
                            href={item.href}
                            onClick={(e) => handleClick(e, item.id)}
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
