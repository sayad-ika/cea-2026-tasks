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
    Header,
    Navbar,
    BottomActionButtons,
    Footer,
    Dropdown,
    InteractiveCard,
    StandardCard,
    AccentBorderCard,
    type MealType,
} from '../components';


export const ComponentShowcase: React.FC = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isOptOutModalOpen, setIsOptOutModalOpen] = useState(false);
    const [showLoading, setShowLoading] = useState(false);
    const [notifyKitchen, setNotifyKitchen] = useState(false);
    const [activeNavItem, setActiveNavItem] = useState('dashboard');
    const [isDarkMode, setIsDarkMode] = useState(false);
    const [selectedDepartment, setSelectedDepartment] = useState('engineering');


    // Sample meal data
    const meals: MealType[] = [
        {
            meal_type: 'lunch',
            is_participating: true,
            opted_out_at: null,
        },
        {
            meal_type: 'snacks',
            is_participating: true,
            opted_out_at: null,
        },
        {
            meal_type: 'optional_dinner',
            is_participating: false,
            opted_out_at: new Date().toISOString(),
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
                            <EmployeeMenuCard meal={meal} onToggle={handleMealToggle} />
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

                {/* Header Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Header
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Application header with logo, user info, theme toggle, and profile avatar
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <Header
                            userName="Sarah Johnson"
                            userRole="Software Engineer"
                            isDarkMode={isDarkMode}
                            onThemeToggle={() => setIsDarkMode(!isDarkMode)}
                        />
                    </div>
                </section>

                {/* Navbar Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Navigation Bar
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Tab-based navigation with smooth animations and hover effects
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <Navbar
                            activeItemId={activeNavItem}
                            onNavItemClick={(itemId: React.SetStateAction<string>) => {
                                setActiveNavItem(itemId);
                                console.log('Nav item clicked:', itemId);
                            }}
                        />
                        <div className="mt-4 text-sm text-[var(--color-text-sub)] text-center">
                            Current Active Tab: <span className="font-bold text-[var(--color-primary)]">{activeNavItem}</span>
                        </div>
                    </div>
                </section>

                {/* Bottom Action Buttons Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Bottom Action Buttons
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Claymorphism-styled action buttons with icons and hover animations
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <BottomActionButtons
                            buttons={[
                                {
                                    id: 'calendar',
                                    label: 'View Calendar',
                                    icon: 'calendar_month',
                                    onClick: () => console.log('Calendar clicked'),
                                },
                                {
                                    id: 'history',
                                    label: 'History',
                                    icon: 'history',
                                    onClick: () => console.log('History clicked'),
                                },
                                {
                                    id: 'preferences',
                                    label: 'Preferences',
                                    icon: 'settings',
                                    onClick: () => console.log('Preferences clicked'),
                                },
                            ]}
                        />
                    </div>
                </section>

                {/* Dropdown Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Dropdown
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Claymorphism-styled dropdown with smooth animations and hover effects
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <Dropdown
                            label="Department"
                            value={selectedDepartment}
                            onChange={setSelectedDepartment}
                            options={[
                                { id: '1', label: 'Design', value: 'design' },
                                { id: '2', label: 'Engineering', value: 'engineering' },
                                { id: '3', label: 'Marketing', value: 'marketing' },
                                { id: '4', label: 'Human Resources', value: 'hr' },
                            ]}
                        />
                        <div className="mt-4 text-sm text-[var(--color-text-sub)] text-center">
                            Selected: <span className="font-bold text-[var(--color-primary)]">{selectedDepartment}</span>
                        </div>
                    </div>
                </section>

                {/* Interactive Card Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Interactive Card
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Clickable card with lift animation on hover for selection grids
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <InteractiveCard
                                icon={<span className="material-symbols-outlined text-3xl">local_cafe</span>}
                                title="Snack Bar"
                                description="Grab a quick bite or coffee from the pantry."
                                buttonLabel="View Menu"
                                onButtonClick={() => console.log('Snack Bar menu clicked')}
                                onClick={() => console.log('Snack Bar card clicked')}
                            />
                            <InteractiveCard
                                icon={<span className="material-symbols-outlined text-3xl">restaurant</span>}
                                iconColor="#22c55e"
                                iconBgColor="#dcfce7"
                                title="Main Dining"
                                description="Full course meals prepared by our chefs."
                                buttonLabel="Reserve"
                                onButtonClick={() => console.log('Main Dining reserve clicked')}
                            />
                            <InteractiveCard
                                icon={<span className="material-symbols-outlined text-3xl">fastfood</span>}
                                iconColor="#8b5cf6"
                                iconBgColor="#ede9fe"
                                title="Quick Bites"
                                description="Fast and delicious options on the go."
                                buttonLabel="Order Now"
                                onButtonClick={() => console.log('Quick Bites order clicked')}
                            />
                        </div>
                    </div>
                </section>

                {/* Standard Card Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Standard Card
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Base card style for static content groups like summaries or informational panels
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <StandardCard
                            title="Nutritional Info"
                            subtitle="Daily intake summary"
                            icon={<span className="material-symbols-outlined">pie_chart</span>}
                            iconBgColor="#f0f9ff"
                            iconColor="#3b82f6"
                        >
                            <div className="space-y-4">
                                <div className="flex justify-between items-center">
                                    <span className="text-sm font-medium text-text-sub">Calories</span>
                                    <span className="font-bold text-[#23170f]">2,100 kcal</span>
                                </div>
                                <div className="h-2 w-full bg-[#e6dccf] rounded-full overflow-hidden shadow-inner">
                                    <div className="h-full bg-blue-400 w-3/4 rounded-full"></div>
                                </div>
                                <div className="flex justify-between items-center mt-2">
                                    <span className="text-sm font-medium text-text-sub">Protein</span>
                                    <span className="font-bold text-[#23170f]">120g</span>
                                </div>
                                <div className="h-2 w-full bg-[#e6dccf] rounded-full overflow-hidden shadow-inner">
                                    <div className="h-full bg-green-400 w-1/2 rounded-full"></div>
                                </div>
                            </div>
                        </StandardCard>
                    </div>
                </section>

                {/* Accent Border Card Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Accent Border Card
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Highlighted state for active items or featured content with pulse indicator
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl p-8 border border-white/60"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <AccentBorderCard
                            title="Chef's Special"
                            badge="TODAY"
                            avatarUrl="https://lh3.googleusercontent.com/aida-public/AB6AXuDttB-Zd88STJy72M0f6YLArBV0Tl6JcqsbvalmTAy_8AywOe4phVL02tIDhmBzF_0en7PxQymXKbuWs2ZmYTeA5h17TdH_T0CECPuSQMs_7Cuslryyjv-n7bG8lVMl9tZ9EePyF-WJQamvjji2HePEm22UcTO3MIOqv41gtGMGEXGZtE6lDv86hD3Y70w-RcQCy8mfewQdK4dEh4csDJ4TTzjE1Y3UdYxTEheOhJC3ApTbqT-BpQDOsIb0KgRR7XtrQHCNGpthJvHv"
                            avatarAlt="Chef Avatar"
                            showPulse={true}
                            footer={
                                <div className="flex justify-between items-center">
                                    <div className="flex -space-x-2">
                                        <div className="w-8 h-8 rounded-full bg-gray-200 border-2 border-white"></div>
                                        <div className="w-8 h-8 rounded-full bg-gray-300 border-2 border-white"></div>
                                        <div className="w-8 h-8 rounded-full bg-gray-400 border-2 border-white flex items-center justify-center text-[10px] text-white font-bold">
                                            +12
                                        </div>
                                    </div>
                                    <button className="text-[#fa8c47] font-bold text-sm hover:underline">
                                        Details
                                    </button>
                                </div>
                            }
                        >
                            <div className="bg-[#FFF5E6] rounded-xl p-4 shadow-clay-inset mb-4">
                                <p className="text-[#23170f] font-medium text-sm italic">
                                    "Grilled Salmon with Asparagus"
                                </p>
                            </div>
                        </AccentBorderCard>
                    </div>
                </section>

                {/* Footer Component Section */}
                <section>
                    <div className="mb-6">
                        <h2 className="text-2xl font-bold text-[var(--color-background-dark)] mb-2">
                            Footer
                        </h2>
                        <p className="text-[var(--color-text-sub)]">
                            Application footer with brand, copyright, and navigation links
                        </p>
                    </div>
                    <div
                        className="bg-white/40 rounded-3xl border border-white/60 overflow-hidden"
                        style={{ boxShadow: 'var(--shadow-clay-inset)' }}
                    >
                        <Footer />
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

