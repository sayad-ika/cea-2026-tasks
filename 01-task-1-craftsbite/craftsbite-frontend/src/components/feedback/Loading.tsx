import React from 'react';

export interface LoadingSpinnerProps {
    message?: string;
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
    message = 'Preparing your menu...',
}) => {
    return (
        <div className="fixed inset-0 z-[100] bg-[var(--color-background-light)]/95 flex flex-col items-center justify-center">
            <div className="relative w-24 h-24 mb-6">
                {/* Background circle */}
                <div className="absolute inset-0 border-4 border-[var(--color-clay-shadow)] rounded-full" />
                {/* Spinning arc */}
                <div className="absolute inset-0 border-4 border-t-[var(--color-primary)] border-r-transparent border-b-transparent border-l-transparent rounded-full animate-spin" />
                {/* Center icon */}
                <div className="absolute inset-0 flex items-center justify-center">
                    <span className="material-symbols-outlined text-4xl text-[var(--color-primary)] animate-pulse">
                        restaurant
                    </span>
                </div>
            </div>
            <h2 className="text-xl font-bold text-[var(--color-background-dark)] animate-pulse">
                {message}
            </h2>
        </div>
    );
};

export interface SkeletonProps {
    className?: string;
    variant?: 'text' | 'circle' | 'rounded' | 'button';
}

export const Skeleton: React.FC<SkeletonProps> = ({
    className = '',
    variant = 'text',
}) => {
    const baseClasses = 'bg-gray-200 animate-pulse';

    const variantClasses = {
        text: 'h-4 rounded',
        circle: 'rounded-full',
        rounded: 'rounded-xl',
        button: 'rounded-2xl',
    };

    return (
        <div
            className={`${baseClasses} ${variantClasses[variant]} ${className}`}
            style={{
                background: '#f6f7f8',
                backgroundImage:
                    'linear-gradient(to right, #f6f7f8 0%, #edeef1 20%, #f6f7f8 40%, #f6f7f8 100%)',
                backgroundRepeat: 'no-repeat',
                backgroundSize: '1000px 100%',
                animation: 'shimmer 1.5s linear infinite',
            }}
        />
    );
};

export const MealCardSkeleton: React.FC = () => {
    return (
        <div
            className="group relative bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col items-center text-center"
            style={{ boxShadow: 'var(--shadow-clay)' }}
        >
            {/* Status badge skeleton */}
            <div className="absolute top-4 right-4">
                <Skeleton className="h-6 w-20" variant="rounded" />
            </div>

            {/* Emoji circle skeleton */}
            <Skeleton className="w-20 h-20 mb-6 mt-4" variant="circle" />

            {/* Title skeleton */}
            <Skeleton className="h-8 w-32 mb-2" variant="text" />

            {/* Time skeleton */}
            <Skeleton className="h-5 w-40 mb-6" variant="text" />

            {/* Divider */}
            <div className="w-full h-px bg-gray-100 mb-6" />

            {/* Toggle row skeleton */}
            <div className="flex items-center justify-between w-full px-4">
                <Skeleton className="h-4 w-12" variant="text" />
                <Skeleton className="w-14 h-8" variant="rounded" />
            </div>
        </div>
    );
};

export default LoadingSpinner;
