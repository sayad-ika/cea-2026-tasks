import React from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'danger';
export type ButtonSize = 'md' | 'lg';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
    size?: ButtonSize;
    icon?: React.ReactNode;
    iconPosition?: 'left' | 'right';
    children: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
    variant = 'primary',
    size = 'md',
    icon,
    iconPosition = 'left',
    children,
    className = '',
    disabled,
    ...rest
}) => {
    const baseClasses =
        'rounded-2xl font-bold transition-all duration-200 flex items-center gap-2 justify-center';

    const sizeClasses = {
        md: 'px-6 py-3',
        lg: 'px-8 py-3',
    };

    const variantClasses = {
        primary:
            'bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white hover:scale-[1.02] active:scale-[0.98]',
        secondary:
            'bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] active:scale-[0.98]',
        danger:
            'bg-gradient-to-br from-[var(--color-danger-gradient-start)] to-[var(--color-danger-gradient-end)] text-white hover:scale-[1.02] active:scale-[0.98]',
    };

    const shadowStyles: Record<ButtonVariant, React.CSSProperties> = {
        primary: {
            boxShadow: 'var(--shadow-clay-button)',
        },
        secondary: {
            boxShadow: 'var(--shadow-clay-button)',
        },
        danger: {
            boxShadow: 'var(--shadow-clay-danger)',
        },
    };

    const hoverShadowClass = {
        primary: 'hover:shadow-[var(--shadow-clay-button-hover)]',
        secondary: 'hover:shadow-[var(--shadow-clay-button-hover)]',
        danger: 'hover:shadow-[var(--shadow-clay-danger-hover)]',
    };

    const activeShadowClass = {
        primary: 'active:shadow-[var(--shadow-primary-active)]',
        secondary: 'active:shadow-[var(--shadow-clay-button-active)]',
        danger: 'active:shadow-[var(--shadow-clay-danger-active)]',
    };

    const disabledClasses = disabled
        ? 'opacity-60 cursor-not-allowed hover:scale-100 active:scale-100'
        : '';

    return (
        <button
            className={`${baseClasses} ${sizeClasses[size]} ${variantClasses[variant]} ${hoverShadowClass[variant]} ${activeShadowClass[variant]} ${disabledClasses} ${className}`}
            style={shadowStyles[variant]}
            disabled={disabled}
            {...rest}
        >
            {icon && iconPosition === 'left' && (
                <span className="flex items-center">{icon}</span>
            )}
            <span>{children}</span>
            {icon && iconPosition === 'right' && (
                <span className="flex items-center">{icon}</span>
            )}
        </button>
    );
};

export default Button;
