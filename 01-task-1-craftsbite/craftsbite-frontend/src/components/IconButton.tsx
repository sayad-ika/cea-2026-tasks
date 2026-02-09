import React from 'react';

export type IconButtonVariant = 'primary' | 'secondary' | 'disabled';
export type IconButtonShape = 'rounded' | 'circle';

export interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: IconButtonVariant;
    shape?: IconButtonShape;
    icon: React.ReactNode;
    ariaLabel: string;
}

export const IconButton: React.FC<IconButtonProps> = ({
    variant = 'secondary',
    shape = 'rounded',
    icon,
    ariaLabel,
    className = '',
    disabled,
    ...rest
}) => {
    const baseClasses =
        'w-12 h-12 flex items-center justify-center transition-all duration-200';

    const shapeClasses = {
        rounded: 'rounded-xl',
        circle: 'rounded-full',
    };

    const variantClasses = {
        primary:
            'bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white hover:scale-110 active:scale-95',
        secondary:
            'bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] hover:scale-110 active:scale-95',
        disabled:
            'bg-[#f0eadd] text-gray-400 border border-gray-200 cursor-not-allowed opacity-60',
    };

    const shadowStyles: Record<IconButtonVariant, React.CSSProperties> = {
        primary: {
            boxShadow: 'var(--shadow-icon-button)',
        },
        secondary: {
            boxShadow: 'var(--shadow-clay-button)',
        },
        disabled: {
            boxShadow: 'none',
        },
    };

    const activeShadowClass = {
        primary: 'active:shadow-[var(--shadow-icon-button-active)]',
        secondary: 'active:shadow-[var(--shadow-clay-button-active)]',
        disabled: '',
    };

    return (
        <button
            className={`${baseClasses} ${shapeClasses[shape]} ${variantClasses[disabled ? 'disabled' : variant]} ${!disabled ? activeShadowClass[variant] : ''} ${className}`}
            style={shadowStyles[disabled ? 'disabled' : variant]}
            disabled={disabled}
            aria-label={ariaLabel}
            {...rest}
        >
            {icon}
        </button>
    );
};

export default IconButton;
