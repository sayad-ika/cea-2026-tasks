import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Header, Footer, EmployeeMenuCard, LoadingSpinner } from '../components';
import type { MealType as MealTypeEnum } from '../types';
import type { MealType } from '../components/cards/EmployeeMenuCard';
import * as mealService from '../services/mealService';

export const Home: React.FC = () => {
    const navigate = useNavigate();
    const { user, isAuthenticated } = useAuth();
    const [meals, setMeals] = useState<MealType[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState('');
    const [currentDate, setCurrentDate] = useState('');

    // Redirect to login if not authenticated
    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/login');
        }
    }, [isAuthenticated, navigate]);

    // Fetch today's meals
    useEffect(() => {
        const fetchMeals = async () => {
            try {
                setIsLoading(true);
                setError('');
                const response = await mealService.getTodaysMeals();

                if (response.success && response.data) {
                    setCurrentDate(response.data.date);

                    // Filter available meals for lunch and snacks only
                    const availableMealTypes = response.data.available_meals.filter(
                        (mealType) => mealType === 'lunch' || mealType === 'snacks'
                    );

                    // Map available meals with their participation status
                    const mealsWithStatus: MealType[] = availableMealTypes.map((mealType) => {
                        const participation = response.data.participations.find(
                            (p) => p.meal_type === mealType
                        );

                        return {
                            meal_type: mealType,
                            is_participating: participation?.is_participating ?? true,
                            opted_out_at: null, // This info isn't in the current API response
                        };
                    });

                    setMeals(mealsWithStatus);
                }
            } catch (err: any) {
                console.error('Error fetching meals:', err);
                setError(err?.error?.message || 'Failed to load meals. Please try again.');
            } finally {
                setIsLoading(false);
            }
        };

        if (isAuthenticated) {
            fetchMeals();
        }
    }, [isAuthenticated]);

    // Handle meal toggle
    const handleToggle = async (mealType: string) => {
        try {
            const mealTypeEnum = mealType as MealTypeEnum;

            // Optimistic update
            setMeals((prevMeals) =>
                prevMeals.map((meal) =>
                    meal.meal_type === mealType
                        ? { ...meal, is_participating: !meal.is_participating }
                        : meal
                )
            );

            // Call API
            await mealService.toggleMealParticipation(mealTypeEnum, currentDate);
        } catch (err: any) {
            console.error('Error toggling meal:', err);

            // Revert optimistic update on error
            setMeals((prevMeals) =>
                prevMeals.map((meal) =>
                    meal.meal_type === mealType
                        ? { ...meal, is_participating: !meal.is_participating }
                        : meal
                )
            );

            setError(err?.error?.message || 'Failed to update meal status. Please try again.');
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

            {/* Main Content */}
            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col justify-center">
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
                        meals.map((meal) => (
                            <EmployeeMenuCard
                                key={meal.meal_type}
                                meal={meal}
                                onToggle={handleToggle}
                            />
                        ))
                    ) : (
                        <div className="col-span-full text-center py-12">
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