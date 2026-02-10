import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Header, Footer, EmployeeMenuCard, LoadingSpinner, Navbar } from '../components';
import type { MealType as MealTypeEnum } from '../types';
import type { MealType } from '../components/cards/EmployeeMenuCard';
import * as mealService from '../services/mealService';
import { CUTOFF_TIMES } from '../utils/constants';
import toast from 'react-hot-toast';

/**
 * Returns true if the cutoff time for a given meal type has passed today.
 */
function isCutoffPassed(mealType: MealTypeEnum): boolean {
    const cutoff = CUTOFF_TIMES[mealType];
    if (!cutoff) return false;
    const [hours, minutes] = cutoff.split(':').map(Number);
    const now = new Date();
    const cutoffDate = new Date();
    cutoffDate.setHours(hours, minutes, 0, 0);
    return now > cutoffDate;
}

export const Home: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useAuth();
    const [meals, setMeals] = useState<MealType[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState('');
    const [currentDate, setCurrentDate] = useState('');

    // Fetch today's meals
    useEffect(() => {
        const fetchMeals = async () => {
            try {
                setIsLoading(true);
                setError('');
                const response = await mealService.getTodaysMeals();

                if (response.success && response.data) {
                    setCurrentDate(response.data.date);

                    // Show ALL available meal types (Phase 3 — no filtering)
                    const mealsWithStatus: MealType[] = response.data.available_meals.map(
                        (mealType) => {
                            const participation = response.data.participations.find(
                                (p) => p.meal_type === mealType
                            );

                            return {
                                meal_type: mealType,
                                is_participating: participation?.is_participating ?? true,
                                opted_out_at: null,
                            };
                        }
                    );

                    setMeals(mealsWithStatus);
                }
            } catch (err: any) {
                console.error('Error fetching meals:', err);
                const msg = err?.error?.message || 'Failed to load meals. Please try again.';
                setError(msg);
                toast.error(msg);
            } finally {
                setIsLoading(false);
            }
        };

        fetchMeals();
    }, []);

    // Handle meal toggle → POST /meals/participation
    const handleToggle = async (mealType: string) => {
        const mealTypeEnum = mealType as MealTypeEnum;

        // Cutoff guard
        if (isCutoffPassed(mealTypeEnum)) {
            toast.error(`Cutoff time for ${mealType.replace('_', ' ')} has passed.`);
            return;
        }

        const currentMeal = meals.find((m) => m.meal_type === mealType);
        const newStatus = !currentMeal?.is_participating;

        // Optimistic update
        setMeals((prev) =>
            prev.map((meal) =>
                meal.meal_type === mealType
                    ? { ...meal, is_participating: newStatus }
                    : meal
            )
        );

        try {
            await mealService.setMealParticipation({
                date: currentDate,
                meal_type: mealTypeEnum,
                participating: newStatus,
            });
            toast.success(
                newStatus
                    ? `Opted in to ${mealType.replace('_', ' ')} ✓`
                    : `Opted out of ${mealType.replace('_', ' ')}`
            );
        } catch (err: any) {
            console.error('Error toggling meal:', err);

            // Revert on error
            setMeals((prev) =>
                prev.map((meal) =>
                    meal.meal_type === mealType
                        ? { ...meal, is_participating: !newStatus }
                        : meal
                )
            );

            const msg = err?.error?.message || 'Failed to update meal status.';
            toast.error(msg);
        }
    };

    // Format date for display
    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            weekday: 'long',
            month: 'short',
            day: 'numeric',
        });
    };

    if (isLoading) {
        return <LoadingSpinner message="Loading today's meals..." />;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            {/* Header */}
            <Header
                userName={user?.name || 'User'}
                userRole={user?.role || 'employee'}
                onThemeToggle={() => { }}
                isDarkMode={false}
            />
            <Navbar />

            {/* Main Content */}
            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
                {/* Page Title */}
                <div className="mb-10 text-center md:text-left">
                    <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
                        Today's Menu
                    </h2>
                    <p className="text-lg text-[var(--color-text-sub)] font-medium">
                        Manage your daily meals for{' '}
                        <span className="text-[var(--color-primary)] font-bold">
                            {currentDate ? formatDate(currentDate) : 'Today'}
                        </span>
                    </p>
                </div>

                {/* Error Message */}
                {error && (
                    <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0">
                        {error}
                    </div>
                )}

                {/* Meal Cards Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
                    {meals.length > 0 ? (
                        meals.map((meal) => {
                            const locked = isCutoffPassed(meal.meal_type as MealTypeEnum);
                            return (
                                <div key={meal.meal_type} className="relative">
                                    <EmployeeMenuCard
                                        meal={meal}
                                        onToggle={locked ? () => { } : handleToggle}
                                    />
                                    {/* Cutoff badge overlay */}
                                    {locked && (
                                        <div className="absolute top-4 right-4 flex items-center gap-1 bg-red-50 text-red-600 text-xs font-bold px-3 py-1.5 rounded-xl border border-red-200">
                                            <span className="material-symbols-outlined text-sm">lock_clock</span>
                                            Cutoff passed
                                        </div>
                                    )}
                                </div>
                            );
                        })
                    ) : (
                        <div className="col-span-full text-center py-12">
                            <span className="material-symbols-outlined text-6xl text-[var(--color-text-sub)] mb-4 block">
                                no_meals
                            </span>
                            <p className="text-[var(--color-text-sub)] text-lg">
                                No meals available for today.
                            </p>
                        </div>
                    )}
                </div>

                {/* Action Buttons */}
                <div className="flex flex-col md:flex-row justify-center items-center gap-6 pb-8">
                    <button
                        className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        onClick={() => navigate('/calendar')}
                    >
                        <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
                            calendar_month
                        </span>
                        <span className="font-bold text-sm uppercase tracking-wider">View Calendar</span>
                    </button>
                    <button
                        className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        onClick={() => navigate('/history')}
                    >
                        <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
                            history
                        </span>
                        <span className="font-bold text-sm uppercase tracking-wider">History</span>
                    </button>
                    <button
                        className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        onClick={() => navigate('/preferences')}
                    >
                        <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
                            settings
                        </span>
                        <span className="font-bold text-sm uppercase tracking-wider">Preferences</span>
                    </button>
                </div>
            </main>

            {/* Footer */}
            <Footer
                links={[
                    { label: 'Privacy', href: '#' },
                    { label: 'Terms', href: '#' },
                    { label: 'Support', href: '#' },
                ]}
            />
        </div>
    );
};
