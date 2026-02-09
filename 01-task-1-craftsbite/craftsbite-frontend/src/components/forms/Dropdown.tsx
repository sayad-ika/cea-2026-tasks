import React, { useState, useRef, useEffect } from 'react';

export interface DropdownOption {
    id: string;
    label: string;
    value: string;
}

export interface DropdownProps {
    label?: string;
    options: DropdownOption[];
    value?: string;
    onChange?: (value: string) => void;
    placeholder?: string;
    className?: string;
}

/**
 * Dropdown component with claymorphism styling
 * Features smooth animations and hover effects
 */
export const Dropdown: React.FC<DropdownProps> = ({
    label,
    options,
    value,
    onChange,
    placeholder = 'Select option',
    className = '',
}) => {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    const selectedOption = options.find((opt) => opt.value === value);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleOptionClick = (optionValue: string) => {
        onChange?.(optionValue);
        setIsOpen(false);
    };

    return (
        <div className={`w-full md:w-72 relative z-50 ${className}`} ref={dropdownRef}>
            {label && (
                <label className="block text-sm font-bold text-text-sub mb-2 ml-1">
                    {label}
                </label>
            )}
            <div className="relative group">
                <button
                    type="button"
                    onClick={() => setIsOpen(!isOpen)}
                    className="w-full bg-background-light text-left rounded-2xl px-5 py-3 shadow-clay-button flex items-center justify-between text-[#23170f] font-semibold border-2 border-transparent hover:border-primary/20 transition-all focus:outline-none focus:shadow-clay-button-active focus:text-primary"
                >
                    <span>{selectedOption ? selectedOption.label : placeholder}</span>
                    <span
                        className={`material-symbols-outlined text-text-sub transition-transform duration-300 ${isOpen ? 'rotate-180' : ''
                            }`}
                    >
                        expand_more
                    </span>
                </button>

                {/* Dropdown Menu */}
                <div
                    className={`absolute top-full left-0 w-full mt-3 bg-background-light rounded-2xl shadow-clay-md p-2 overflow-hidden z-50 transition-all duration-200 transform origin-top border border-white/40 ${isOpen
                            ? 'opacity-100 visible translate-y-0'
                            : 'opacity-0 invisible translate-y-2'
                        }`}
                >
                    <ul className="flex flex-col gap-1">
                        {options.map((option) => {
                            const isSelected = option.value === value;

                            return (
                                <li key={option.id}>
                                    <button
                                        type="button"
                                        onClick={() => handleOptionClick(option.value)}
                                        className={`w-full flex items-center justify-between px-4 py-3 rounded-xl transition-all duration-200 ${isSelected
                                                ? 'bg-[#fa8c47] text-white shadow-clay-inset-sm font-semibold'
                                                : 'text-text-main hover:bg-[#fa8c47] hover:text-white hover:shadow-inner font-medium'
                                            }`}
                                    >
                                        <span>{option.label}</span>
                                        {isSelected && (
                                            <span className="material-symbols-outlined text-sm">check</span>
                                        )}
                                    </button>
                                </li>
                            );
                        })}
                    </ul>
                </div>
            </div>
        </div>
    );
};
