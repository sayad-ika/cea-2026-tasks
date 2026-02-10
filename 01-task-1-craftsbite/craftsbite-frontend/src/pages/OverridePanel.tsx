import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Header, Footer, Navbar, LoadingSpinner } from '../components';
import type { MealType as MealTypeEnum } from '../types';
import { MEAL_TYPES } from '../utils/constants';
import * as userService from '../services/userService';
import * as mealService from '../services/mealService';
import toast from 'react-hot-toast';

// Normalized user entry for the dropdown (works for both admin and team_lead)
interface SelectableUser {
    id: string;
    name: string;
    email: string;
}

export const OverridePanel: React.FC = () => {
    const { user } = useAuth();
    const [selectableUsers, setSelectableUsers] = useState<SelectableUser[]>([]);
    const [teamName, setTeamName] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Form state
    const [selectedUserId, setSelectedUserId] = useState('');
    const [selectedDate, setSelectedDate] = useState(
        new Date().toISOString().split('T')[0] // YYYY-MM-DD
    );
    const [selectedMealType, setSelectedMealType] = useState<MealTypeEnum>('lunch');
    const [participating, setParticipating] = useState(false);
    const [reason, setReason] = useState('');
    const [error, setError] = useState('');

    const isTeamLead = user?.role === 'team_lead';

    // Fetch users — admin gets all, team_lead gets only team members
    useEffect(() => {
        const fetchUsers = async () => {
            try {
                if (isTeamLead) {
                    const res = await userService.getTeamMembers();
                    if (res.success && res.data) {
                        setSelectableUsers(
                            res.data.members.map((m) => ({
                                id: m.id,
                                name: m.name,
                                email: m.email,
                            }))
                        );
                        setTeamName(res.data.members[0]?.team_name || null);
                    }
                } else {
                    // Admin flow
                    const res = await userService.getUsers();
                    if (res.success && res.data) {
                        const list = Array.isArray(res.data) ? res.data : [];
                        setSelectableUsers(
                            list.map((u) => ({ id: u.id, name: u.name, email: u.email }))
                        );
                    }
                }
            } catch (err: any) {
                setError(err?.error?.message || 'Failed to load users.');
            } finally {
                setIsLoading(false);
            }
        };
        fetchUsers();
    }, [isTeamLead]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (!selectedUserId) {
            setError('Please select a user.');
            return;
        }
        if (!reason.trim()) {
            setError('Please provide a reason for the override.');
            return;
        }

        try {
            setIsSubmitting(true);
            await mealService.overrideParticipation({
                user_id: selectedUserId,
                date: selectedDate,
                meal_type: selectedMealType,
                participating,
                reason: reason.trim(),
            });
            toast.success('Participation overridden successfully');
            setReason('');
        } catch (err: any) {
            const msg = err?.error?.message || 'Failed to override participation.';
            setError(msg);
            toast.error(msg);
        } finally {
            setIsSubmitting(false);
        }
    };

    const mealTypeOptions = Object.entries(MEAL_TYPES).map(([value, label]) => ({
        value,
        label,
    }));

    if (isLoading) {
        return <LoadingSpinner message="Loading users..." />;
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
                        Override Participation
                    </h2>
                    <p className="text-lg text-[var(--color-text-sub)] font-medium">
                        {isTeamLead
                            ? 'Override meal participation for your team members.'
                            : 'Mark meal participation on behalf of an employee.'}
                    </p>
                    {isTeamLead && teamName && (
                        <div className="mt-3 inline-flex items-center gap-2 bg-[var(--color-background-light)] rounded-2xl px-4 py-2 text-sm font-bold text-[var(--color-primary)]"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}>
                            <span className="material-symbols-outlined text-base">groups</span>
                            Team: {teamName}
                        </div>
                    )}
                </div>

                {/* Override Form Card */}
                <div
                    className="bg-[var(--color-background-light)] rounded-3xl p-8 md:p-10 max-w-2xl w-full mx-auto md:mx-0"
                    style={{ boxShadow: 'var(--shadow-clay)' }}
                >
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {/* Error Message */}
                        {error && (
                            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                                {error}
                            </div>
                        )}

                        {/* User Selector */}
                        <div className="space-y-2">
                            <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                                {isTeamLead ? 'Team Member' : 'Employee'}
                            </label>
                            <select
                                value={selectedUserId}
                                onChange={(e) => setSelectedUserId(e.target.value)}
                                className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                                style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                            >
                                <option value="">{isTeamLead ? 'Select a team member...' : 'Select an employee...'}</option>
                                {selectableUsers.map((u) => (
                                    <option key={u.id} value={u.id}>
                                        {u.name} ({u.email})
                                    </option>
                                ))}
                            </select>
                        </div>

                        {/* Date Selector */}
                        <div className="space-y-2">
                            <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                                Date
                            </label>
                            <input
                                type="date"
                                value={selectedDate}
                                onChange={(e) => setSelectedDate(e.target.value)}
                                className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50"
                                style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                            />
                        </div>

                        {/* Meal Type Selector */}
                        <div className="space-y-2">
                            <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                                Meal Type
                            </label>
                            <select
                                value={selectedMealType}
                                onChange={(e) => setSelectedMealType(e.target.value as MealTypeEnum)}
                                className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 cursor-pointer"
                                style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                            >
                                {mealTypeOptions.map((opt) => (
                                    <option key={opt.value} value={opt.value}>
                                        {opt.label}
                                    </option>
                                ))}
                            </select>
                        </div>

                        {/* Participation Toggle */}
                        <div className="space-y-2">
                            <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                                Participation Status
                            </label>
                            <div className="flex gap-4">
                                <button
                                    type="button"
                                    onClick={() => setParticipating(true)}
                                    className={`flex-1 py-3.5 rounded-2xl font-bold text-sm transition-all duration-200 ${participating
                                        ? 'bg-green-500 text-white'
                                        : 'bg-[var(--color-background-light)] text-[var(--color-text-sub)]'
                                        }`}
                                    style={{
                                        boxShadow: participating
                                            ? '6px 6px 12px #a3d9a5, -6px -6px 12px #ffffff'
                                            : 'var(--shadow-clay-button)',
                                    }}
                                >
                                    ✓ Opt In
                                </button>
                                <button
                                    type="button"
                                    onClick={() => setParticipating(false)}
                                    className={`flex-1 py-3.5 rounded-2xl font-bold text-sm transition-all duration-200 ${!participating
                                        ? 'bg-red-500 text-white'
                                        : 'bg-[var(--color-background-light)] text-[var(--color-text-sub)]'
                                        }`}
                                    style={{
                                        boxShadow: !participating
                                            ? '6px 6px 12px #fca5a5, -6px -6px 12px #ffffff'
                                            : 'var(--shadow-clay-button)',
                                    }}
                                >
                                    ✗ Opt Out
                                </button>
                            </div>
                        </div>

                        {/* Reason */}
                        <div className="space-y-2">
                            <label className="block text-sm font-bold text-[var(--color-text-main)] ml-1">
                                Reason <span className="text-red-400">*</span>
                            </label>
                            <textarea
                                value={reason}
                                onChange={(e) => setReason(e.target.value)}
                                placeholder="e.g., User on leave, requested via email..."
                                rows={3}
                                className="w-full px-4 py-4 rounded-2xl bg-[var(--color-background-light)] border-none text-[var(--color-text-main)] placeholder-[var(--color-text-sub)]/50 outline-none transition-all duration-200 focus:ring-2 focus:ring-[var(--color-primary)]/50 resize-none"
                                style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                            />
                        </div>

                        {/* Submit Button */}
                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full py-4 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold text-lg hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        >
                            {isSubmitting ? (
                                <>
                                    <span className="animate-spin material-symbols-outlined">progress_activity</span>
                                    <span>Submitting...</span>
                                </>
                            ) : (
                                <>
                                    <span className="material-symbols-outlined">swap_horiz</span>
                                    <span>Override Participation</span>
                                </>
                            )}
                        </button>
                    </form>
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
