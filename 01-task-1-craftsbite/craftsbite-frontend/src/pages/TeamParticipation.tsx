import React, { useState, useEffect } from "react";
import { format, addDays, subDays } from "date-fns";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as mealService from "../services/mealService";
import type { TeamData } from "../types";
import toast from "react-hot-toast";

export const TeamParticipation: React.FC = () => {
    const { user } = useAuth();
    const [selectedDate, setSelectedDate] = useState(new Date());
    const [teamData, setTeamData] = useState<TeamData[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState("");

    const fetchTeamParticipation = async (date: Date) => {
        try {
            setIsLoading(true);
            setError("");
            const dateStr = format(date, "yyyy-MM-dd");
            const response = await mealService.getTeamParticipation(dateStr);

            if (response.success && response.data) {
                setTeamData(response.data.teams);
            }
        } catch (err: any) {
            console.error("Error fetching team participation:", err);
            const msg = err?.response?.data?.error?.message || "Failed to load team participation.";
            setError(msg);
            toast.error(msg);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchTeamParticipation(selectedDate);
    }, [selectedDate]);

    const handlePrevDay = () => setSelectedDate((prev) => subDays(prev, 1));
    const handleNextDay = () => setSelectedDate((prev) => addDays(prev, 1));

    if (isLoading && teamData.length === 0) {
        return <LoadingSpinner message="Loading team participation..." />;
    }

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
            <Header userName={user?.name} userRole={user?.role} />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
                {/* Page Title & Controls */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
                    <div>
                        <h2 className="text-4xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
                            Team Participation
                        </h2>
                        <p className="text-lg text-[var(--color-text-sub)] font-medium">
                            View daily meal status for your team members
                        </p>
                    </div>

                    {/* Date Navigation */}
                    <div className="flex items-center gap-4 bg-[var(--color-background-light)] p-2 rounded-2xl border border-[#e6dccf] shadow-sm">
                        <button
                            onClick={handlePrevDay}
                            className="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-orange-50 text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors"
                        >
                            <span className="material-symbols-outlined">chevron_left</span>
                        </button>

                        <div className="flex flex-col items-center min-w-[140px]">
                            <span className="text-sm font-bold text-[var(--color-text-sub)] uppercase tracking-wider">
                                {format(selectedDate, "EEEE")}
                            </span>
                            <span className="text-lg font-black text-[var(--color-background-dark)]">
                                {format(selectedDate, "MMM d, yyyy")}
                            </span>
                        </div>

                        <button
                            onClick={handleNextDay}
                            className="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-orange-50 text-[var(--color-text-sub)] hover:text-[var(--color-primary)] transition-colors"
                        >
                            <span className="material-symbols-outlined">chevron_right</span>
                        </button>
                    </div>
                </div>

                {error && (
                    <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm">
                        {error}
                    </div>
                )}

                {/* Team Data */}
                <div className="space-y-8">
                    {teamData.length > 0 ? (
                        teamData.map((team) => (
                            <div key={team.team_id} className="bg-[#e8d5c4] rounded-3xl p-6 md:p-8 shadow-sm border border-[#e6dccf]">
                                <div className="flex items-center gap-3 mb-6 pb-4 border-b border-gray-100">
                                    <div className="w-10 h-10 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center">
                                        <span className="material-symbols-outlined">groups</span>
                                    </div>
                                    <h3 className="text-2xl font-bold text-[var(--color-background-dark)]">
                                        {team.team_name}
                                    </h3>
                                    <span className="ml-auto bg-gray-100 text-gray-600 px-3 py-1 rounded-full text-xs font-bold">
                                        {team.members.length} Members
                                    </span>
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                    {team.members.map((member) => (
                                        <div
                                            key={member.user_id}
                                            className="bg-[#ffeddc] rounded-2xl p-5 border border-[#f0eadd] hover:border-orange-200 transition-colors"
                                        >
                                            <div className="flex items-center gap-3 mb-4">
                                                <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-xs font-bold text-gray-600">
                                                    {member.name.charAt(0)}
                                                </div>
                                                <span className="font-bold text-[var(--color-background-dark)]">
                                                    {member.name}
                                                </span>
                                            </div>

                                            <div className="space-y-3">
                                                {member.participations.map((p) => (
                                                    <div key={p.meal_type} className="flex items-center justify-between text-sm">
                                                        <span className="capitalize text-[var(--color-text-sub)] font-medium">
                                                            {p.meal_type.replace("_", " ")}
                                                        </span>
                                                        <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg border ${p.is_participating
                                                            ? "bg-green-50 text-green-700 border-green-200"
                                                            : "bg-red-50 text-red-700 border-red-200"
                                                            }`}>
                                                            <span className="material-symbols-outlined text-[16px]">
                                                                {p.is_participating ? "check_circle" : "cancel"}
                                                            </span>
                                                            <span className="font-bold text-xs">
                                                                {p.is_participating ? "Yes" : "No"}
                                                            </span>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ))
                    ) : (
                        !isLoading && (
                            <div className="text-center py-20 bg-white rounded-3xl border border-dashed border-gray-300">
                                <span className="material-symbols-outlined text-6xl text-gray-300 mb-4 block">
                                    groups_off
                                </span>
                                <p className="text-gray-500 text-lg font-medium">
                                    No team participation data found for this date.
                                </p>
                            </div>
                        )
                    )}
                </div>
            </main>

            <Footer links={[{ label: "Privacy", href: "#" }, { label: "Terms", href: "#" }, { label: "Support", href: "#" }]} />
        </div>
    );
};
