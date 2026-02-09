import React, { useState } from 'react';
import {
    Button,
    IconButton,
    EmployeeMenuCard,
    MealModal,
    Toast,
    ToastContainer,
    LoadingSpinner,
    MealCardSkeleton,
    MealOptOutModal,
    type MealType,
} from '../components';

export const ComponentShowcase: React.FC = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isOptOutModalOpen, setIsOptOutModalOpen] = useState(false);
    const [showLoading, setShowLoading] = useState(false);
    const [notifyKitchen, setNotifyKitchen] = useState(false);

    // Sample meal data
    const meals: MealType[] = [
        {
            id: 'breakfast',
            name: 'Breakfast',
            emoji: '🥞',
            timeRange: '07:00 AM - 10:00 AM',
            isOptedIn: true,
            backgroundColor: '#fff9e6',
        },
        {
            id: 'lunch',
            name: 'Lunch',
            emoji: '🍱',
            timeRange: '12:00 PM - 02:00 PM',
            isOptedIn: true,
            backgroundColor: '#fff0e6',
        },
        {
            id: 'dinner',
            name: 'Dinner',
            emoji: '🍝',
            timeRange: '07:00 PM - 09:00 PM',
            isOptedIn: false,
            backgroundColor: '#ffe6e6',
        },
    ];

    const handleMealToggle = (mealId: string) => {
        console.log('Toggled meal:', mealId);
    };

    const handleModalConfirm = () => {
        console.log('Confirmed with notify kitchen:', notifyKitchen);
        setIsModalOpen(false);
    };

    const handleShowLoading = () => {
        setShowLoading(true);
        setTimeout(() => setShowLoading(false), 2000);
    };

    return (
        <div className="min-h-screen bg-[var(--color-background-light)] font-[var(--font-family-display)] p-8">
            {/* Loading Spinner (when active) */}
            {showLoading && <LoadingSpinner message="Loading demo..." />}

            {/* Toast Container with demo toasts */}
            <ToastContainer>
                <Toast
                    variant="success"
                    title="Update Successful"
                    message="Your lunch preference has been successfully updated to 'Opted In'."
                />
                <Toast
                    variant="error"
                    title="Connection Error"
                    message="Failed to save dinner changes. Please check your network."
                />
                <Toast
                    variant="info"
                    title="Menu Update"
                    message="Next week's menu is now available for viewing."
                />
            </ToastContainer>

            {/* Header */}
            <header className="mb-12 text-center">
                <div className="flex items-center justify-center gap-3 mb-4">
                    <div
                        className="w-12 h-12 rounded-xl bg-[var(--color-primary)] text-white flex items-center justify-center"
                        style={{ boxShadow: 'var(--shadow-clay-button)' }}
                    >
                        <span className="material-symbols-outlined">design_services</span>
                    </div>
                    <h1 className="text-4xl font-black text-[var(--color-background-dark)]">
                        CraftsBite Design System
                    </h1>
                </div>
                <p className="text-[var(--color-text-sub)] text-lg font-medium">
                    Component Showcase - Claymorphism Theme
                </p>
            </header>

            <div className="max-w-7xl mx-auto space-y-16">
                {/* Buttons Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Buttons
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Primary, Secondary, and Danger variants with icon support
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <div className="space-y-6">
                            <div>
                                <h3 className="text-sm font-bold text-[var(--color-text-sub)] uppercase mb-3">
                                    Primary
                                </h3>
                                <div className="flex flex-wrap gap-4">
                                    <Button variant="primary">Get Started</Button>
                                    <Button variant="primary" size="lg">Large Button</Button>
                                    <Button
                                        variant="primary"
                                        icon={<span className="material-symbols-outlined text-[20px]">add_circle</span>}
                                    >
                                        Add Meal
                                    </Button>
                                </div>
                            </div>
                            <div>
                                <h3 className="text-sm font-bold text-[var(--color-text-sub)] uppercase mb-3">
                                    Secondary
                                </h3>
                                <div className="flex flex-wrap gap-4">
                                    <Button variant="secondary">Cancel</Button>
                                    <Button
                                        variant="secondary"
                                        icon={<span className="material-symbols-outlined text-[20px]">filter_list</span>}
                                    >
                                        Filters
                                    </Button>
                                </div>
                            </div>
                            <div>
                                <h3 className="text-sm font-bold text-[var(--color-text-sub)] uppercase mb-3">
                                    Danger
                                </h3>
                                <div className="flex flex-wrap gap-4">
                                    <Button variant="danger">Delete Account</Button>
                                    <Button
                                        variant="danger"
                                        icon={<span className="material-symbols-outlined text-[20px]">block</span>}
                                    >
                                        Opt Out
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Icon Buttons Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Icon Buttons
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Compact icon-only buttons with rounded and circular shapes
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <div className="flex flex-wrap gap-6 items-center">
                            <div className="flex flex-col items-center gap-2">
                                <IconButton
                                    variant="primary"
                                    icon={<span className="material-symbols-outlined">edit</span>}
                                    ariaLabel="Edit"
                                />
                                <span className="text-xs text-[var(--color-text-sub)]">Primary</span>
                            </div>
                            <div className="flex flex-col items-center gap-2">
                                <IconButton
                                    variant="secondary"
                                    shape="rounded"
                                    icon={<span className="material-symbols-outlined">notifications</span>}
                                    ariaLabel="Notifications"
                                />
                                <span className="text-xs text-[var(--color-text-sub)]">Secondary</span>
                            </div>
                            <div className="flex flex-col items-center gap-2">
                                <IconButton
                                    variant="secondary"
                                    shape="circle"
                                    icon={<span className="material-symbols-outlined">search</span>}
                                    ariaLabel="Search"
                                />
                                <span className="text-xs text-[var(--color-text-sub)]">Circle</span>
                            </div>
                            <div className="flex flex-col items-center gap-2">
                                <IconButton
                                    disabled
                                    icon={<span className="material-symbols-outlined">lock</span>}
                                    ariaLabel="Locked"
                                />
                                <span className="text-xs text-[var(--color-text-sub)]">Disabled</span>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Meal Cards Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Employee Menu Cards
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Interactive meal selection cards with toggle switches
                        </p>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        {meals.map((meal) => (
                            <EmployeeMenuCard key={meal.id} meal={meal} onToggle={handleMealToggle} />
                        ))}
                    </div>
                </section>

                {/* Skeleton Loading Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Skeleton Loading
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Shimmer effect skeletons for loading states
                        </p>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        <MealCardSkeleton />
                        <MealCardSkeleton />
                        <MealCardSkeleton />
                    </div>
                </section>

                {/* Modals & Loading Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Modals & Loading
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Interactive modal dialogs and loading states
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <div className="flex flex-wrap gap-4 justify-center">
                            <Button variant="primary" onClick={() => setIsModalOpen(true)}>
                                Open Meal Preference Modal
                            </Button>
                            <Button variant="danger" onClick={() => setIsOptOutModalOpen(true)}>
                                Open Opt-Out Modal
                            </Button>
                            <Button variant="secondary" onClick={handleShowLoading}>
                                Show Loading Spinner (2s)
                            </Button>
                        </div>
                    </div>
                </section>
            </div>

            {/* Meal Modal */}
            <MealModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onConfirm={handleModalConfirm}
                mealName="Lunch"
                mealEmoji="🍱"
                mealTime="12:00 PM - 02:00 PM"
                mealBackgroundColor="#fff0e6"
                notifyKitchen={notifyKitchen}
                onNotifyKitchenChange={setNotifyKitchen}
            />

            {/* Opt-Out Modal */}
            <MealOptOutModal
                isOpen={isOptOutModalOpen}
                onClose={() => setIsOptOutModalOpen(false)}
                onConfirm={() => {
                    console.log('Opt out confirmed');
                    setIsOptOutModalOpen(false);
                }}
                mealName="lunch"
                cutoffTime="10:30 AM"
            />
        </div>
    );
};

export default ComponentShowcase;

