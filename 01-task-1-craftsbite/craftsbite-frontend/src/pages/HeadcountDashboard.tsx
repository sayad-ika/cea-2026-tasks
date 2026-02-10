import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Header, Footer, Navbar, LoadingSpinner } from '../components';
import type { HeadcountData, MealType as MealTypeEnum } from '../types';
import { MEAL_TYPES } from '../utils/constants';
import * as headcountService from '../services/headcountService';

export const HeadcountDashboard: React.FC = () => {
    const { user } = useAuth();
    const [headcount, setHeadcount] = useState<HeadcountData | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchHeadcount = async () => {
            try {
                setIsLoading(true);
                const res = await headcountService.getTodayHeadcount();
                if (res.success && res.data) {
                    setHeadcount(res.data);
                }
            } catch (err: any) {
                setError(err?.error?.message || 'Failed to load headcount data.');
            } finally {
                setIsLoading(false);
            }
        };
        fetchHeadcount();
    }, []);

    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            weekday: 'long',
            month: 'short',
            day: 'numeric',
        });
    };

    if (isLoading) {
        return <LoadingSpinner message="Loading headcount data..." />;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header
                userName={user?.name || 'User'}
                userRole={user?.role || 'admin'}
            />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
                {/* Page Title */}
                <div className="mb-10 text-center md:text-left">
                    <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
                        Today's Headcount
                    </h2>
                    <p className="text-lg text-[var(--color-text-sub)] font-medium">
                        Meal participation summary for{' '}
                        <span className="text-[var(--color-primary)] font-bold">
                            {headcount ? formatDate(headcount.date) : 'Today'}
                        </span>
                    </p>
                </div>

                {/* Error Message */}
                {error && (
                    <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm max-w-2xl mx-auto md:mx-0">
                        {error}
                    </div>
                )}

                {/* Total Active Users Badge */}
                {headcount && (
                    <div
                        className="mb-8 inline-flex items-center gap-3 bg-[var(--color-background-light)] rounded-2xl px-6 py-4 self-start"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}
                    >
                        <span className="material-symbols-outlined text-[var(--color-primary)] text-2xl">
                            groups
                        </span>
                        <div>
                            <p className="text-2xl font-black text-[var(--color-background-dark)]">
                                {headcount.total_active_users}
                            </p>
                            <p className="text-xs text-[var(--color-text-sub)] font-medium">Total Active Users</p>
                        </div>
                    </div>
                )}

                {/* Headcount Cards Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
                    {headcount?.meals && Object.keys(headcount.meals).length > 0 ? (
                        Object.entries(headcount.meals).map(([mealType, counts]) => {
                            const total = counts.participating + counts.opted_out;
                            const participationRate = total > 0
                                ? Math.round((counts.participating / total) * 100)
                                : 0;

                            return (
                                <div
                                    key={mealType}
                                    className="bg-[var(--color-background-light)] rounded-3xl p-8 flex flex-col transition-transform hover:-translate-y-1 duration-300"
                                    style={{ boxShadow: 'var(--shadow-clay)' }}
                                >
                                    {/* Meal Name */}
                                    <h3 className="text-xl font-bold text-[var(--color-background-dark)] mb-6 capitalize">
                                        {MEAL_TYPES[mealType as MealTypeEnum] || mealType}
                                    </h3>

                                    {/* Counts */}
                                    <div className="flex justify-between items-end mb-6">
                                        <div className="text-center">
                                            <p className="text-4xl font-black text-green-600">{counts.participating}</p>
                                            <p className="text-sm text-[var(--color-text-sub)] font-medium mt-1">Eating</p>
                                        </div>
                                        <div className="text-center">
                                            <p className="text-4xl font-black text-red-500">{counts.opted_out}</p>
                                            <p className="text-sm text-[var(--color-text-sub)] font-medium mt-1">Opted Out</p>
                                        </div>
                                    </div>

                                    {/* Divider */}
                                    <div className="w-full h-px bg-gradient-to-r from-transparent via-[var(--color-clay-shadow)] to-transparent mb-4" />

                                    {/* Participation Rate */}
                                    <div className="flex items-center justify-between">
                                        <span className="text-sm font-bold text-[var(--color-text-sub)]">
                                            Participation
                                        </span>
                                        <span className="text-sm font-black text-[var(--color-primary)]">
                                            {participationRate}%
                                        </span>
                                    </div>

                                    {/* Progress Bar */}
                                    <div
                                        className="w-full h-2 rounded-full mt-2"
                                        style={{ backgroundColor: 'var(--color-clay-shadow)' }}
                                    >
                                        <div
                                            className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-dark)] transition-all duration-500"
                                            style={{ width: `${participationRate}%` }}
                                        />
                                    </div>
                                </div>
                            );
                        })
                    ) : (
                        <div className="col-span-full text-center py-12">
                            <span className="material-symbols-outlined text-6xl text-[var(--color-text-sub)] mb-4 block">
                                no_meals
                            </span>
                            <p className="text-[var(--color-text-sub)] text-lg">
                                No headcount data available for today.
                            </p>
                        </div>
                    )}
                </div>
            </main>

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
