import React, { useState, useEffect, useCallback } from "react";
import { format, parseISO } from "date-fns";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as userService from "../services/userService";
import toast from "react-hot-toast";
import type { MealHistoryEntry } from "../types/history.types";
import { MEAL_LABELS, type User } from "../types";
import { getUserMealHistory } from "../services/historyService";


const ACTION_MAP: Record<string, { label: string; icon: string; color: string }> = {
    opted_in:    { label: "Opted In",     icon: "check_circle", color: "text-green-600" },
    opted_out:   { label: "Opted Out",    icon: "cancel",       color: "text-red-500"   },
    override_in: { label: "Override In",  icon: "edit",         color: "text-[var(--color-primary)]" },
    override_out:{ label: "Override Out", icon: "block",        color: "text-[var(--color-primary)]" },
};

function fmtDate(iso: string) {
    try { return format(parseISO(iso), "MMM d, yyyy"); } catch { return iso; }
}
function fmtTime(iso: string) {
    try { return format(parseISO(iso), "hh:mm a · MMM d, yyyy"); } catch { return iso; }
}

function EntryRow({ entry }: { entry: MealHistoryEntry }) {
    const act  = ACTION_MAP[entry.action] ?? { label: entry.action, icon: "info", color: "text-[var(--color-text-sub)]" };
    const meal = MEAL_LABELS[entry.meal_type] ?? entry.meal_type.replace(/_/g, " ");

    return (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 py-4 border-b border-[#e6dccf] last:border-0">
            <div className={`flex items-center gap-1.5 min-w-[130px] ${act.color}`}>
                <span className="material-symbols-outlined text-[18px]">{act.icon}</span>
                <span className="text-sm font-semibold">{act.label}</span>
            </div>

            <span className="text-sm text-[var(--color-text-main)] font-medium min-w-[110px]">
                {fmtDate(entry.date)}
            </span>

            <span className="px-2.5 py-0.5 rounded-lg border border-[#e6dccf] text-xs font-semibold text-[var(--color-text-sub)] bg-[var(--color-background-light)] capitalize min-w-[100px]">
                {meal}
            </span>

            <span className="flex-1 text-xs text-[var(--color-text-sub)] italic truncate">
                {entry.reason ? `"${entry.reason}"` : "—"}
            </span>

            <div className="text-right shrink-0">
                <p className="text-xs font-semibold text-[var(--color-text-main)]">
                    {entry.changed_by ? entry.changed_by.name : "Self"}
                </p>
                <p className="text-[11px] text-[var(--color-text-sub)]">{fmtTime(entry.created_at)}</p>
            </div>
        </div>
    );
}
``
export const MealParticipationHistory: React.FC = () => {
    const { user } = useAuth();

    const [users, setUsers]               = useState<User[]>([]);
    const [usersLoading, setUsersLoading] = useState(true);
    const [search, setSearch]             = useState("");
    const [selectedUserId, setSelectedUserId] = useState("");
    const [startDate, setStartDate]       = useState("");
    const [endDate, setEndDate]           = useState("");
    const [entries, setEntries]           = useState<MealHistoryEntry[]>([]);
    const [isFetching, setIsFetching]     = useState(false);
    const [hasFetched, setHasFetched]     = useState(false);

    useEffect(() => {
        (async () => {
            try {
                const res = await userService.getUsers();
                if (res.success && Array.isArray(res.data)) setUsers(res.data);
            } catch {
                toast.error("Failed to load users.");
            } finally {
                setUsersLoading(false);
            }
        })();
    }, []);

    const filteredUsers = search
        ? users.filter(u =>
              u.name.toLowerCase().includes(search.toLowerCase()) ||
              u.email.toLowerCase().includes(search.toLowerCase()))
        : [];

    const selectedUser = users.find(u => u.id === selectedUserId) ?? null;

    const fetchHistory = useCallback(async () => {
        if (!selectedUserId) { toast.error("Select a user first."); return; }
        setIsFetching(true);
        setHasFetched(false);
        try {
            const res = await getUserMealHistory(selectedUserId, {
                start_date: startDate || undefined,
                end_date:   endDate   || undefined,
            });
            setEntries(res.success && Array.isArray(res.data) ? res.data : []);
        } catch {
            toast.error("Failed to fetch history.");
            setEntries([]);
        } finally {
            setIsFetching(false);
            setHasFetched(true);
        }
    }, [selectedUserId, startDate, endDate]);

    return (
        <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col">
            <Header userName={user?.name} userRole={user?.role} />
            <Navbar />

            <main className="flex-grow container mx-auto px-6 py-8 md:px-12 max-w-5xl flex flex-col gap-6">

                {/* Page header */}
                <div>
                    <h2 className="text-4xl font-black tracking-tight text-[var(--color-background-dark)] mb-1">
                        Meal Participation History
                    </h2>
                    <p className="text-[var(--color-text-sub)] font-medium">
                        View opt-in / opt-out history and admin overrides for any employee.
                    </p>
                </div>

                {/* Filter card */}
                <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 shadow-[var(--shadow-clay-card)] flex flex-col gap-5">

                    {/* Employee search */}
                    <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                            Employee
                        </label>
                        <div className="relative">
                            <span className="absolute left-3 top-1/2 -translate-y-1/2 material-symbols-outlined text-[18px] text-[var(--color-text-sub)]">
                                search
                            </span>
                            <input
                                type="text"
                                placeholder="Search by name or email…"
                                value={search}
                                onChange={e => { setSearch(e.target.value); setSelectedUserId(""); }}
                                className="w-full pl-10 pr-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all placeholder:text-[var(--color-text-sub)]/50"
                            />
                        </div>

                        {/* Dropdown */}
                        {search && !selectedUserId && (
                            <div className="border border-[#e6dccf] rounded-xl bg-[#FFFDF5] shadow-[var(--shadow-clay-card)] max-h-52 overflow-y-auto divide-y divide-[#e6dccf]">
                                {usersLoading ? (
                                    <p className="px-4 py-3 text-sm text-[var(--color-text-sub)]">Loading…</p>
                                ) : filteredUsers.length === 0 ? (
                                    <p className="px-4 py-3 text-sm text-[var(--color-text-sub)]">No users found.</p>
                                ) : filteredUsers.map(u => (
                                    <button
                                        key={u.id}
                                        type="button"
                                        onClick={() => { setSelectedUserId(u.id); setSearch(u.name); }}
                                        className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-[var(--color-background-light)] transition-colors"
                                    >
                                        <div>
                                            <p className="text-sm font-semibold text-[var(--color-text-main)]">{u.name}</p>
                                            <p className="text-xs text-[var(--color-text-sub)]">{u.email} · {u.role}</p>
                                        </div>
                                    </button>
                                ))}
                            </div>
                        )}

                        {/* Selected pill */}
                        {selectedUser && (
                            <div className="flex items-center gap-2 mt-1">
                                <span className="material-symbols-outlined text-[16px] text-[var(--color-primary)]">person</span>
                                <span className="text-sm font-semibold text-[var(--color-text-main)]">{selectedUser.name}</span>
                                <span className="text-xs text-[var(--color-text-sub)]">· {selectedUser.email}</span>
                                <button
                                    type="button"
                                    onClick={() => { setSelectedUserId(""); setSearch(""); }}
                                    className="ml-1 text-[var(--color-text-sub)] hover:text-red-500 transition-colors"
                                >
                                    <span className="material-symbols-outlined text-[16px]">close</span>
                                </button>
                            </div>
                        )}
                    </div>

                    {/* Date range + button */}
                    <div className="flex flex-col sm:flex-row gap-4 items-end">
                        <div className="flex flex-col gap-1.5 flex-1">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                                Start Date <span className="normal-case font-normal opacity-60 tracking-normal">(optional)</span>
                            </label>
                            <input
                                type="date"
                                value={startDate}
                                onChange={e => setStartDate(e.target.value)}
                                className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                            />
                        </div>

                        <div className="flex flex-col gap-1.5 flex-1">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)]">
                                End Date <span className="normal-case font-normal opacity-60 tracking-normal">(optional)</span>
                            </label>
                            <input
                                type="date"
                                value={endDate}
                                onChange={e => setEndDate(e.target.value)}
                                className="w-full px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                            />
                        </div>

                        <button
                            type="button"
                            onClick={fetchHistory}
                            disabled={!selectedUserId || isFetching}
                            className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2 whitespace-nowrap"
                        >
                            {isFetching ? (
                                <>
                                    <span className="material-symbols-outlined text-[16px] animate-spin">progress_activity</span>
                                    Loading…
                                </>
                            ) : (
                                <>
                                    <span className="material-symbols-outlined text-[16px]">history</span>
                                    Fetch History
                                </>
                            )}
                        </button>
                    </div>
                </div>

                {/* Results */}
                {isFetching && <LoadingSpinner message="Fetching history…" />}

                {!isFetching && hasFetched && (
                    <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-8 shadow-[var(--shadow-clay-card)]">
                        {/* Column headers */}
                        {entries.length > 0 && (
                            <div className="flex flex-col sm:flex-row sm:items-center gap-3 pb-2 mb-2 border-b border-[#e6dccf]">
                                <span className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] min-w-[130px]">Action</span>
                                <span className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] min-w-[110px]">Date</span>
                                <span className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] min-w-[100px]">Meal</span>
                                <span className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] flex-1">Reason</span>
                                <span className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] text-right shrink-0">Changed By</span>
                            </div>
                        )}

                        {entries.length > 0 ? (
                            <>
                                <p className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] mb-4">
                                    {entries.length} record{entries.length !== 1 ? "s" : ""} · {selectedUser?.name}
                                </p>
                                {entries.map(entry => <EntryRow key={entry.id} entry={entry} />)}
                            </>
                        ) : (
                            <div className="flex flex-col items-center gap-3 py-12 text-center">
                                <span className="material-symbols-outlined text-4xl text-[var(--color-text-sub)] opacity-30">history</span>
                                <p className="text-[var(--color-text-sub)] font-medium">No history found for the selected filters.</p>
                                <p className="text-sm text-[var(--color-text-sub)] opacity-60">Try a wider date range.</p>
                            </div>
                        )}
                    </div>
                )}

                {/* Initial state */}
                {!isFetching && !hasFetched && (
                    <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] px-8 py-16 flex flex-col items-center gap-3 text-center shadow-[var(--shadow-clay-card)]">
                        <span className="material-symbols-outlined text-4xl text-[var(--color-text-sub)] opacity-30">manage_search</span>
                        <p className="text-[var(--color-text-sub)] font-medium">
                            Select an employee and click <strong>Fetch History</strong>.
                        </p>
                    </div>
                )}

            </main>

            <Footer links={[{ label: "Privacy", href: "#" }, { label: "Terms", href: "#" }, { label: "Support", href: "#" }]} />
        </div>
    );
};
