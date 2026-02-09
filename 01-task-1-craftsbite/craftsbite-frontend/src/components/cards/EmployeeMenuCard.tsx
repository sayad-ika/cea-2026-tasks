import React from 'react';

export interface MealType {
    meal_type: 'lunch' | 'snacks' | 'iftar' | 'event_dinner' | 'optional_dinner';
    is_participating: boolean;
    opted_out_at: string | null;
}

// Updated Component
export interface EmployeeMenuCardProps {
    meal: MealType;
    onToggle: (mealType: string) => void;
}

const mealConfig = {
    lunch: {
        emoji: '🍱',
        name: 'Lunch',
        timeRange: '12:00 PM - 02:00 PM',
        backgroundColor: '#FFF0E6',
        toggleColor: 'var(--color-primary)',
    },
    snacks: {
        emoji: '🍩',
        name: 'Snacks',
        timeRange: '04:00 PM - 05:00 PM',
        backgroundColor: '#FFE6E6',
        toggleColor: '#f87171',
    },
    iftar: {
        emoji: '🌙',
        name: 'Iftar',
        timeRange: '06:00 PM - 07:00 PM',
        backgroundColor: '#E6F0FF',
        toggleColor: '#60a5fa',
    },
    event_dinner: {
        emoji: '🍽️',
        name: 'Event Dinner',
        timeRange: '07:00 PM - 09:00 PM',
        backgroundColor: '#FFE6F0',
        toggleColor: '#f472b6',
    },
    optional_dinner: {
        emoji: '🍝',
        name: 'Optional Dinner',
        timeRange: '07:00 PM - 09:00 PM',
        backgroundColor: '#FFE6E6',
        toggleColor: '#f87171',
    },
};

export const EmployeeMenuCard: React.FC<EmployeeMenuCardProps> = ({ meal, onToggle }) => {
    const config = mealConfig[meal.meal_type];

    return (
        <div
            className={`group relative bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col items-center text-center transition-transform hover:-translate-y-1 duration-300 ${meal.meal_type === 'lunch' ? 'border-2 border-[var(--color-primary)]/10' : ''
                }`}
            style={{ boxShadow: 'var(--shadow-clay)' }}
        >
            {/* Status Badge */}
            <div className="absolute top-4 right-4">
                <span
                    className={`px-3 py-1 rounded-full text-xs font-bold shadow-sm flex items-center gap-1 ${meal.is_participating
                        ? 'bg-green-100 text-green-700'
                        : 'bg-red-100 text-red-700'
                        }`}
                >
                    <span
                        className={`w-2 h-2 rounded-full ${meal.is_participating ? 'bg-green-500' : 'bg-red-500'
                            }`}
                    />
                    {meal.is_participating ? 'Opted In' : 'Opted Out'}
                </span>
            </div>

            {/* Emoji Icon */}
            <div
                className="w-20 h-20 rounded-full flex items-center justify-center text-5xl mb-6 mt-4"
                style={{
                    backgroundColor: config.backgroundColor,
                    boxShadow: 'var(--shadow-clay-inset)',
                }}
            >
                {config.emoji}
            </div>

            {/* Meal Name */}
            <h3 className="text-2xl font-bold text-[var(--color-background-dark)] mb-1">
                {config.name}
            </h3>

            {/* Time Range */}
            <p className="text-[var(--color-text-sub)] font-medium mb-6">
                {config.timeRange}
            </p>

            {/* Divider */}
            <div
                className={`w-full h-px mb-6 ${meal.meal_type === 'lunch'
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
                        checked={meal.is_participating}
                        onChange={() => onToggle(meal.meal_type)}
                    />
                    <div
                        className="w-14 h-8 rounded-full peer peer-focus:outline-none after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-6 after:w-6 after:transition-all peer-checked:after:translate-x-full peer-checked:after:border-white shadow-inner"
                        style={{
                            backgroundColor: meal.is_participating
                                ? config.toggleColor
                                : 'var(--color-clay-shadow)',
                        }}
                    />
                </label>
            </div>
        </div>
    );
};

export default EmployeeMenuCard;