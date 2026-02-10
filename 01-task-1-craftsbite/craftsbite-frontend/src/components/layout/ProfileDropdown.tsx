import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import toast from 'react-hot-toast';

export interface ProfileDropdownProps {
    userName?: string;
    userRole?: string;
    userAvatarUrl?: string;
}

/**
 * ProfileDropdown component for user profile menu
 * Displays a dropdown menu with logout functionality
 * Supports both hover and click interactions
 */
export const ProfileDropdown: React.FC<ProfileDropdownProps> = ({
    userName = 'John Doe',
    userRole = 'UX Designer',
    userAvatarUrl = 'https://lh3.googleusercontent.com/aida-public/AB6AXuDttB-Zd88STJy72M0f6YLArBV0Tl6JcqsbvalmTAy_8AywOe4phVL02tIDhmBzF_0en7PxQymXKbuWs2ZmYTeA5h17TdH_T0CECPuSQMs_7Cuslryyjv-n7bG8lVMl9tZ9EePyF-WJQamvjji2HePEm22UcTO3MIOqv41gtGMGEXGZtE6lDv86hD3Y70w-RcQCy8mfewQdK4dEh4csDJ4TTzjE1Y3UdYxTEheOhJC3ApTbqT-BpQDOsIb0KgRR7XtrQHCNGpthJvHv',
}) => {
    const [isOpen, setIsOpen] = useState(false);
    const [isHovering, setIsHovering] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);
    const { logout } = useAuth();
    const navigate = useNavigate();

    // Handle click outside to close dropdown
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Handle logout
    const handleLogout = () => {
        logout();
        toast.success('Logged out successfully');
        navigate('/login');
        setIsOpen(false);
    };

    // Toggle dropdown on click
    const toggleDropdown = () => {
        setIsOpen(!isOpen);
    };

    // Handle hover states
    const handleMouseEnter = () => {
        setIsHovering(true);
        setIsOpen(true);
    };

    const handleMouseLeave = () => {
        setIsHovering(false);
        // Small delay to allow clicking on dropdown items
        setTimeout(() => {
            if (!isHovering) {
                setIsOpen(false);
            }
        }, 200);
    };

    return (
        <div 
            className="relative" 
            ref={dropdownRef}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            {/* User Avatar - Clickable and Hoverable */}
            <div 
                className="w-12 h-12 rounded-full overflow-hidden border-2 border-[var(--color-clay-light)] cursor-pointer hover:scale-105 transition-transform duration-200"
                style={{ boxShadow: 'var(--shadow-clay-button)' }}
                onClick={toggleDropdown}
            >
                <img
                    alt={`${userName} profile avatar`}
                    className="w-full h-full object-cover"
                    src={userAvatarUrl}
                />
            </div>

            {/* Dropdown Menu */}
            <div
                className={`absolute right-0 top-full mt-3 w-64 bg-[var(--color-background-light)] rounded-2xl p-4 overflow-hidden z-50 transition-all duration-200 transform origin-top border border-white/40 ${
                    isOpen
                        ? 'opacity-100 visible translate-y-0'
                        : 'opacity-0 invisible translate-y-2'
                }`}
                style={{ boxShadow: 'var(--shadow-clay-md)' }}
            >
                {/* User Info Section */}
                <div className="flex flex-col items-center pb-4 border-b border-[var(--color-clay-shadow)] mb-3">
                    <div className="w-16 h-16 rounded-full overflow-hidden border-2 border-[var(--color-primary)] mb-3"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}>
                        <img
                            alt={`${userName} profile avatar`}
                            className="w-full h-full object-cover"
                            src={userAvatarUrl}
                        />
                    </div>
                    <h3 className="text-base font-bold text-[var(--color-background-dark)]">{userName}</h3>
                    <p className="text-xs text-[var(--color-text-sub)] capitalize">{userRole.replace('_', ' ')}</p>
                </div>

                {/* Menu Items */}
                <ul className="flex flex-col gap-1">
                    <li>
                        <button
                            type="button"
                            onClick={handleLogout}
                            className="w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 text-[var(--color-danger)] hover:bg-[var(--color-danger)] hover:text-white cursor-pointer font-medium"
                            style={{ 
                                boxShadow: 'var(--shadow-clay-button)',
                            }}
                        >
                            <span className="material-symbols-outlined text-xl">logout</span>
                            <span>Logout</span>
                        </button>
                    </li>
                </ul>
            </div>
        </div>
    );
};
