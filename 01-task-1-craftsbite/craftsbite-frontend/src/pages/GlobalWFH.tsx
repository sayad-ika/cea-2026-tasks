import React, { useEffect, useState, useMemo } from "react";
import { Header, Navbar, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as workLocationService from "../services/workLocationService";
import type { GlobalWFHPolicy } from "../services/workLocationService";
import toast from "react-hot-toast";

const getDaysInMonth = (year: number, month: number) => {
    return new Date(year, month + 1, 0).getDate();
};

const getFirstDayOfMonth = (year: number, month: number) => {
    // 0 = Sunday, 1 = Monday, ... 6 = Saturday
    // We want Monday to be 0, so we adjust
    const day = new Date(year, month, 1).getDay();
    return day === 0 ? 6 : day - 1;
};

const isDateInRange = (date: Date, startDate: string, endDate: string) => {
    const d = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    const start = new Date(startDate);
    const end = new Date(endDate);
    // Reset hours to avoid timezone issues affecting comparison
    start.setHours(0, 0, 0, 0);
    end.setHours(0, 0, 0, 0);
    return d >= start && d <= end;
};

export const GlobalWFH: React.FC = () => {
    const { user } = useAuth();
    const [policies, setPolicies] = useState<GlobalWFHPolicy[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Form State
    const [startDate, setStartDate] = useState("");
    const [endDate, setEndDate] = useState("");
    const [reason, setReason] = useState("");

    // Calendar State
    const [currentDate, setCurrentDate] = useState(new Date());

    const fetchPolicies = async () => {
        try {
            setIsLoading(true);
            const year = currentDate.getFullYear();
            const month = currentDate.getMonth();
            const daysInMonth = getDaysInMonth(year, month);

            // Format dates as YYYY-MM-DD
            // Month is 0-indexed, so we add 1 for the API
            const start_date = `${year}-${String(month + 1).padStart(2, '0')}-01`;
            const end_date = `${year}-${String(month + 1).padStart(2, '0')}-${daysInMonth}`;

            const response = await workLocationService.getGlobalWFHPolicies({
                start_date,
                end_date
            });
            if (response.success && response.data) {
                setPolicies(response.data);
            }
        } catch (error) {
            toast.error("Failed to load WFH policies");
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchPolicies();
    }, [currentDate]);

    const handleSubmit = async () => {
        if (!startDate || !endDate) {
            toast.error("Please select both start and end dates");
            return;
        }

        if (new Date(startDate) > new Date(endDate)) {
            toast.error("Start date must be before end date");
            return;
        }

        try {
            setIsSubmitting(true);
            await workLocationService.createGlobalWFHPolicy({
                start_date: startDate,
                end_date: endDate,
                reason: reason || "Admin declared WFH",
            });
            toast.success("Global WFH policy set successfully");
            setStartDate("");
            setEndDate("");
            setReason("");
            fetchPolicies();
        } catch (error) {
            toast.error("Failed to create WFH policy");
        } finally {
            setIsSubmitting(false);
        }
    };

    /* const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this policy?")) return;
        try {
            await workLocationService.deleteGlobalWFHPolicy(id);
            toast.success("Policy deleted successfully");
            await fetchPolicies();
        } catch (error) {
            toast.error("Failed to delete policy");
        }
    }; */

    const calendarDays = useMemo(() => {
        const year = currentDate.getFullYear();
        const month = currentDate.getMonth();
        const daysInMonth = getDaysInMonth(year, month);
        const firstDay = getFirstDayOfMonth(year, month);

        const days = [];

        // Empty cells for days before the first day of the month
        for (let i = 0; i < firstDay; i++) {
            days.push(null);
        }

        // Days of the month
        for (let i = 1; i <= daysInMonth; i++) {
            days.push(new Date(year, month, i));
        }

        return days;
    }, [currentDate]);

    const monthName = currentDate.toLocaleString("default", { month: "long" });
    const year = currentDate.getFullYear();

    const handlePrevMonth = () => {
        setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1));
    };

    const handleNextMonth = () => {
        setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1));
    };

    const isWeekend = (date: Date) => {
        const day = date.getDay();
        return day === 0 || day === 6;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header
                userName={user?.name || "User"}
                userRole={user?.role || "admin"}
            />
            <Navbar />

            <main className="flex-grow container mx-auto px-4 md:px-6 lg:px-12 py-8 flex flex-col gap-8">
                {isLoading ? (
                    <div className="flex-grow flex items-center justify-center">
                        <LoadingSpinner message="Loading policies..." />
                    </div>
                ) : (
                    <>
                        <section className="w-full">
                            <div
                                className="bg-[#FFFDF5] p-6 md:p-8 rounded-3xl border border-white/50 flex flex-col md:flex-row items-center justify-between gap-6"
                                style={{ boxShadow: "var(--shadow-clay-card)" }}
                            >
                                <div className="flex flex-col gap-2">
                                    <h2 className="text-2xl font-bold text-[#23170f]">Policy Configuration</h2>
                                    <p className="text-[var(--color-text-sub)]">Set global work from home days for all employees.</p>
                                </div>
                                <div className="flex flex-col md:flex-row items-end md:items-center gap-4 w-full md:w-auto">
                                    <div className="flex flex-col gap-1 w-full md:w-auto">
                                        <label className="text-xs font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">From</label>
                                        <div className="relative">
                                            <input
                                                type="date"
                                                value={startDate}
                                                onChange={(e) => setStartDate(e.target.value)}
                                                className="pl-10 pr-4 py-3 rounded-2xl bg-[var(--color-background-light)] border-none focus:ring-2 focus:ring-[var(--color-primary)]/50 outline-none w-full md:w-48 text-[#23170f] font-medium"
                                                style={{ boxShadow: "var(--shadow-clay-inset)" }}
                                            />
                                            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-primary)]">calendar_today</span>
                                        </div>
                                    </div>
                                    <div className="flex flex-col gap-1 w-full md:w-auto">
                                        <label className="text-xs font-bold text-[var(--color-text-sub)] uppercase tracking-wider ml-1">To</label>
                                        <div className="relative">
                                            <input
                                                type="date"
                                                value={endDate}
                                                onChange={(e) => setEndDate(e.target.value)}
                                                className="pl-10 pr-4 py-3 rounded-2xl bg-[var(--color-background-light)] border-none focus:ring-2 focus:ring-[var(--color-primary)]/50 outline-none w-full md:w-48 text-[#23170f] font-medium"
                                                style={{ boxShadow: "var(--shadow-clay-inset)" }}
                                            />
                                            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-primary)]">event</span>
                                        </div>
                                    </div>
                                    <button
                                        onClick={handleSubmit}
                                        disabled={isSubmitting}
                                        className="w-full md:w-auto px-8 py-3 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 mt-auto h-[48px] disabled:opacity-50 disabled:cursor-not-allowed"
                                        style={{
                                            boxShadow: "8px 8px 16px #e6dccf, -8px -8px 16px #ffffff"
                                        }}
                                    >
                                        {isSubmitting ? (
                                            <span className="material-symbols-outlined animate-spin">progress_activity</span>
                                        ) : (
                                            <span className="material-symbols-outlined">save</span>
                                        )}
                                        <span>Set Global WFH</span>
                                    </button>
                                </div>
                            </div>
                        </section>

                        <div className="flex flex-col lg:flex-row gap-8">
                            {/* Calendar View */}
                            <div className="flex-grow lg:w-2/3">
                                <div
                                    className="bg-[#FFFDF5] p-6 md:p-8 rounded-3xl border border-white/50 h-full"
                                    style={{ boxShadow: "var(--shadow-clay-card)" }}
                                >
                                    <div className="flex justify-between items-center mb-6">
                                        <h3 className="text-xl font-bold text-[#23170f] flex items-center gap-2">
                                            <span className="material-symbols-outlined text-[var(--color-primary)]">calendar_month</span>
                                            {monthName} {year}
                                        </h3>
                                        <div className="flex gap-2">
                                            <button
                                                onClick={handlePrevMonth}
                                                className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] hover:text-[var(--color-primary)] flex items-center justify-center transition-all"
                                                style={{ boxShadow: "var(--shadow-clay-button)" }}
                                            >
                                                <span className="material-symbols-outlined">chevron_left</span>
                                            </button>
                                            <button
                                                onClick={handleNextMonth}
                                                className="w-10 h-10 rounded-xl bg-[var(--color-background-light)] text-[var(--color-text-sub)] hover:text-[var(--color-primary)] flex items-center justify-center transition-all"
                                                style={{ boxShadow: "var(--shadow-clay-button)" }}
                                            >
                                                <span className="material-symbols-outlined">chevron_right</span>
                                            </button>
                                        </div>
                                    </div>

                                    <div className="flex gap-6 mb-6 text-sm">
                                        <div className="flex items-center gap-2">
                                            <div className="w-3 h-3 rounded-full bg-[var(--color-background-light)] border border-[#e6dccf]"></div>
                                            <span className="text-[var(--color-text-sub)]">Office Day</span>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <div className="w-3 h-3 rounded-full bg-[var(--color-wfh-teal)] border border-[var(--color-wfh-teal-dark)]/30"></div>
                                            <span className="text-[var(--color-text-sub)]">Global WFH</span>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <div className="w-3 h-3 rounded-full bg-gray-100 border border-gray-200"></div>
                                            <span className="text-[var(--color-text-sub)]">Weekend</span>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-7 gap-4 mb-2">
                                        {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((day, i) => (
                                            <div key={day} className={`text-center text-xs font-bold uppercase tracking-wider py-2 ${i >= 5 ? 'text-[var(--color-primary)]' : 'text-[var(--color-text-sub)]'}`}>
                                                {day}
                                            </div>
                                        ))}
                                    </div>

                                    <div className="grid grid-cols-7 gap-3 md:gap-4 auto-rows-fr">
                                        {calendarDays.map((date, index) => {
                                            if (!date) return <div key={`empty-${index}`} className="p-2 md:p-4 min-h-[80px] md:min-h-[100px] rounded-2xl bg-transparent opacity-50"></div>;

                                            const isWeekEndDay = isWeekend(date);
                                            const wfhPolicy = policies.find(p => isDateInRange(date, p.start_date, p.end_date) && p.is_active);
                                            const isWFH = !!wfhPolicy;

                                            return (
                                                <div
                                                    key={date.toISOString()}
                                                    className={`relative p-2 md:p-4 min-h-[80px] md:min-h-[100px] rounded-2xl flex flex-col justify-between items-start transition-colors ${isWFH
                                                        ? "bg-[var(--color-wfh-teal)]/40 shadow-[var(--shadow-clay-inset)] border border-[var(--color-wfh-teal-dark)]/20"
                                                        : isWeekEndDay
                                                            ? "bg-gray-50 border border-white shadow-sm"
                                                            : "bg-[var(--color-background-light)] shadow-[var(--shadow-clay-inset)] group hover:bg-white"
                                                        }`}
                                                >
                                                    <span className={`font-bold ${isWFH ? "text-[var(--color-wfh-teal-dark)]" : isWeekEndDay ? "text-gray-400" : "text-[var(--color-text-sub)] group-hover:text-[var(--color-primary)]"}`}>
                                                        {date.getDate()}
                                                    </span>
                                                    {isWFH && (
                                                        <span className="bg-white/80 backdrop-blur-sm text-[10px] md:text-xs text-[var(--color-wfh-teal-dark)] font-bold px-2 py-1 rounded-lg border border-[var(--color-wfh-teal-dark)]/10 shadow-sm whitespace-nowrap overflow-hidden text-ellipsis max-w-full">
                                                            🏠 Global WFH
                                                        </span>
                                                    )}
                                                </div>
                                            );
                                        })}
                                    </div>
                                </div>
                            </div>

                            {/* Policies List */}
                            {/* <div className="flex-grow lg:w-1/3">
                        <div
                            className="bg-[#FFFDF5] p-6 md:p-8 rounded-3xl border border-white/50 h-full"
                            style={{ boxShadow: "var(--shadow-clay-card)" }}
                        >
                            <div className="flex items-center gap-2 mb-6 border-b border-[#e6dccf] pb-4">
                                <span className="material-symbols-outlined text-[var(--color-primary)]">format_list_bulleted</span>
                                <h3 className="text-xl font-bold text-[#23170f]">Current Policies</h3>
                            </div>

                            <div className="space-y-4">
                                {isLoading ? (
                                    <LoadingSpinner message="Loading policies..." />
                                ) : policies.length === 0 ? (
                                    <p className="text-[var(--color-text-sub)] text-center py-4">No active policies found.</p>
                                ) : (
                                    policies.map(policy => {
                                        const isActive = policy.is_active && new Date(policy.end_date) >= new Date();
                                        return (
                                            <div
                                                key={policy.id}
                                                className={`p-4 rounded-2xl border transition-transform ${isActive ? "bg-white shadow-[var(--shadow-clay-button)] border-white/40 hover:scale-[1.01]" : "bg-[var(--color-background-light)] shadow-[var(--shadow-clay-inset)] border-transparent opacity-80 hover:opacity-100"}`}
                                            >
                                                <div className="flex justify-between items-start mb-2">
                                                    <span className={`text-xs font-bold px-2 py-1 rounded-md ${isActive ? "text-[var(--color-wfh-teal-dark)] bg-[var(--color-wfh-teal)]/50" : "text-[var(--color-text-sub)] bg-gray-200"}`}>
                                                        {isActive ? "Active" : "Expired"}
                                                    </span>
                                                    <div className="flex gap-2">
                                                        <button onClick={() => handleDelete(policy.id)} className="text-gray-400 hover:text-[var(--color-danger)] transition-colors">
                                                        <span className="material-symbols-outlined text-[18px]">delete</span>
                                                    </button>
                                                    </div>
                                                </div>
                                                <h4 className="font-bold text-[#23170f]">{policy.reason}</h4>
                                                <div className="flex items-center gap-2 text-[var(--color-text-sub)] text-sm mt-1">
                                                    <span className="material-symbols-outlined text-[16px]">calendar_today</span>
                                                    <span>{policy.start_date.split('T')[0]} - {policy.end_date.split('T')[0]}</span>
                                                </div>
                                            </div>
                                        )
                                    })
                                )}
                            </div>
                        </div>
                            </div> */
                            }
                        </div>
                    </>
                )}
            </main>
        </div>
    );
};
