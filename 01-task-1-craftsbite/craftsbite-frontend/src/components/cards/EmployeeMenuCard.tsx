import React from 'react';

export interface MealType {
    id: string;
    name: string;
    emoji: string;
    timeRange: string;
    isOptedIn: boolean;
    backgroundColor: string;
}

export interface EmployeeMenuCardProps {
    meal: MealType;
    onToggle: (mealId: string) => void;
}

export const EmployeeMenuCard: React.FC<EmployeeMenuCardProps> = ({ meal, onToggle }) => {
    return (
        <div
            className={`group relative bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col items-center text-center transition-transform hover:-translate-y-1 duration-300 ${meal.name === 'Lunch' ? 'border-2 border-[var(--color-primary)]/10' : ''
                }`}
            style={{ boxShadow: 'var(--shadow-clay)' }}
        >
            {/* Status Badge */}
            <div className="absolute top-4 right-4">
                <span
                    className={`px-3 py-1 rounded-full text-xs font-bold shadow-sm flex items-center gap-1 ${meal.isOptedIn
                            ? 'bg-green-100 text-green-700'
                            : 'bg-red-100 text-red-700'
                        }`}
                >
                    <span
                        className={`w-2 h-2 rounded-full ${meal.isOptedIn ? 'bg-green-500' : 'bg-red-500'
                            }`}
                    />
                    {meal.isOptedIn ? 'Opted In' : 'Opted Out'}
                </span>
            </div>

            {/* Emoji Icon */}
            <div
                className="w-20 h-20 rounded-full flex items-center justify-center text-5xl mb-6 mt-4"
                style={{
                    backgroundColor: meal.backgroundColor,
                    boxShadow: 'var(--shadow-clay-inset)',
                }}
            >
                {meal.emoji}
            </div>

            {/* Meal Name */}
            <h3 className="text-2xl font-bold text-[var(--color-background-dark)] mb-1">
                {meal.name}
            </h3>

            {/* Time Range */}
            <p className="text-[var(--color-text-sub)] font-medium mb-6">
                {meal.timeRange}
            </p>

            {/* Divider */}
            <div
                className={`w-full h-px mb-6 ${meal.name === 'Lunch'
                        ? 'bg-gradient-to-r from-transparent via-[var(--color-primary)]/30 to-transparent'
                        : 'bg-gradient-to-r from-transparent via-[var(--color-clay-shadow)] to-transparent'
                    }`}
            />

            {/* Toggle Switch */}
            <div className="flex items-center justify-between w-full px-4">
                <span className="text-sm font-bold text-[var(--color-text-sub)]">
                    Status
                </span>
                <label className="relative inline-flex items-center cursor-pointer">
                    <input
                        type="checkbox"
                        className="sr-only peer"
                        checked={meal.isOptedIn}
                        onChange={() => onToggle(meal.id)}
                    />
                    <div
                        className="w-14 h-8 rounded-full peer peer-focus:outline-none after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-6 after:w-6 after:transition-all peer-checked:after:translate-x-full peer-checked:after:border-white shadow-inner"
                        style={{
                            backgroundColor: meal.isOptedIn
                                ? meal.name === 'Breakfast'
                                    ? '#fbbf24'
                                    : meal.name === 'Lunch'
                                        ? 'var(--color-primary)'
                                        : '#f87171'
                                : 'var(--color-clay-shadow)',
                        }}
                    />
                </label>
            </div>
        </div>
    );
};

export default EmployeeMenuCard;
