import React, { useState } from "react";
import {
    fetchWFHReport,
    type TeamMonthlyReport,
} from "../../services/workLocationService";

interface WFHMonthlyReportProps {
    memberNames?: Record<string, string>;
}

function formatMonthLabel(yearMonth: string): string {
    const [year, month] = yearMonth.split("-");
    const date = new Date(Number(year), Number(month) - 1, 1);
    return date.toLocaleString("default", { month: "long", year: "numeric" });
}

function buildCopyText(
    report: TeamMonthlyReport,
    memberNames: Record<string, string>,
): string {
    const title = `WFH Monthly Report — ${formatMonthLabel(report.year_month)}`;
    const divider = "─".repeat(56);
    const summary = [
        `Allowance : ${report.allowance} day${report.allowance !== 1 ? "s" : ""}`,
        `Employees : ${report.total_employees}`,
        `Over Limit: ${report.over_limit_count}`,
        `Extra Days: ${report.total_extra_days}`,
    ].join("   |   ");

    const colName = "Name / ID";
    const colUsed = "WFH Used";
    const colLimit = "Over Limit";
    const colExtra = "Extra Days";
    const pad = (s: string, n: number) => s.padEnd(n);

    const header = `${pad(colName, 28)}${pad(colUsed, 12)}${pad(colLimit, 12)}${colExtra}`;
    const rows = [...report.members]
        .sort((a, b) => Number(b.is_over_limit) - Number(a.is_over_limit))
        .map((m) => {
            const name = memberNames[m.user_id] ?? m.user_id;
            return `${pad(name, 28)}${pad(String(m.wfh_days), 12)}${pad(m.is_over_limit ? "⚠ Yes" : "No", 12)}${m.extra_days > 0 ? `+${m.extra_days}` : m.extra_days}`;
        });

    return [
        title,
        divider,
        summary,
        "",
        divider,
        header,
        divider,
        ...rows,
        divider,
    ].join("\n");
}

export const WFHMonthlyReport: React.FC<WFHMonthlyReportProps> = ({
    memberNames = {},
}) => {
    const currentMonth = new Date().toISOString().slice(0, 7);
    const [selectedMonth, setSelectedMonth] = useState(currentMonth);
    const [report, setReport] = useState<TeamMonthlyReport | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string>("");
    const [copied, setCopied] = useState(false);

    const handleFetch = async () => {
        setError("");
        setReport(null);
        setIsLoading(true);
        try {
            const res = await fetchWFHReport(selectedMonth);
            if (res.success && res.data) {
                setReport(res.data);
            } else {
                setError("No report data returned.");
            }
        } catch (err: any) {
            setError(
                err?.response?.data?.error?.message ??
                    err?.message ??
                    "Failed to fetch WFH report.",
            );
        } finally {
            setIsLoading(false);
        }
    };

    const handleCopy = async () => {
        if (!report) return;
        const text = buildCopyText(report, memberNames);
        try {
            await navigator.clipboard.writeText(text);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            const el = document.createElement("textarea");
            el.value = text;
            document.body.appendChild(el);
            el.select();
            document.execCommand("copy");
            document.body.removeChild(el);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    const sortedMembers = report
        ? [...report.members].sort(
              (a, b) => Number(b.is_over_limit) - Number(a.is_over_limit),
          )
        : [];

    return (
        <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] shadow-[var(--shadow-clay-card)] p-6 md:p-8 flex flex-col gap-6">
            <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
                <div>
                    <h3 className="text-xl font-black tracking-tight text-[var(--color-background-dark)]">
                        WFH Monthly Report
                    </h3>
                    <p className="text-sm text-[var(--color-text-sub)] font-medium mt-0.5">
                        Team-wide work-from-home usage and allowance rollup.
                    </p>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                    <input
                        type="month"
                        value={selectedMonth}
                        onChange={(e) => setSelectedMonth(e.target.value)}
                        className="px-4 py-2.5 rounded-xl border border-[#e6dccf] bg-[var(--color-background-light)] text-[var(--color-text-main)] text-sm font-medium shadow-[var(--shadow-clay-button)] outline-none focus:border-[var(--color-primary)] transition-all"
                    />
                    <button
                        type="button"
                        onClick={handleFetch}
                        disabled={isLoading}
                        className="px-5 py-2.5 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white text-sm font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all flex items-center gap-2"
                    >
                        {isLoading ? (
                            <>
                                <span className="material-symbols-outlined text-[16px] animate-spin">
                                    progress_activity
                                </span>
                                Loading…
                            </>
                        ) : (
                            <>
                                <span className="material-symbols-outlined text-[16px]">
                                    bar_chart
                                </span>
                                Fetch Report
                            </>
                        )}
                    </button>
                </div>
            </div>

            {error && (
                <div className="flex items-start gap-3 p-4 rounded-xl bg-red-50 border border-red-200">
                    <span className="material-symbols-outlined text-red-500 text-[20px] shrink-0 mt-0.5">
                        error
                    </span>
                    <div>
                        <p className="text-sm font-semibold text-red-600">
                            Failed to load report
                        </p>
                        <p className="text-sm text-red-500 mt-0.5">{error}</p>
                    </div>
                </div>
            )}

            {!report && !error && !isLoading && (
                <div className="flex flex-col items-center justify-center py-10 gap-3 text-center">
                    <span className="material-symbols-outlined text-5xl text-[var(--color-text-sub)] opacity-30">
                        home_work
                    </span>
                    <p className="text-sm font-semibold text-[var(--color-text-sub)] opacity-60">
                        Select a month and click{" "}
                        <span className="font-black">Fetch Report</span> to see
                        the WFH rollup.
                    </p>
                </div>
            )}

            {report && (
                <>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                        <StatPill
                            icon="calendar_month"
                            label="Month"
                            value={formatMonthLabel(report.year_month)}
                            accent={false}
                        />
                        <StatPill
                            icon="beach_access"
                            label="WFH Allowance"
                            value={`${report.allowance} day${report.allowance !== 1 ? "s" : ""}`}
                            accent={false}
                        />
                        <StatPill
                            icon="group"
                            label="Employees"
                            value={String(report.total_employees)}
                            accent={false}
                        />
                        <StatPill
                            icon="warning"
                            label="Over Limit"
                            value={`${report.over_limit_count} member${report.over_limit_count !== 1 ? "s" : ""}`}
                            accent={report.over_limit_count > 0}
                        />
                    </div>

                    {/* Extra-days callout */}
                    {report.total_extra_days > 0 && (
                        <div className="flex items-center gap-3 p-3.5 rounded-xl bg-amber-50 border border-amber-200">
                            <span className="material-symbols-outlined text-amber-500 text-[20px] shrink-0">
                                info
                            </span>
                            <p className="text-sm text-amber-700 font-semibold">
                                {report.total_extra_days} extra WFH day
                                {report.total_extra_days !== 1 ? "s" : ""}{" "}
                                recorded above the allowance this month.
                            </p>
                        </div>
                    )}

                    <div className="overflow-x-auto -mx-1">
                        <table className="w-full text-left border-separate border-spacing-y-2 min-w-[480px]">
                            <thead>
                                <tr>
                                    <th className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-2 px-4">
                                        Member
                                    </th>
                                    <th className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-2 px-4 text-center">
                                        WFH Used
                                    </th>
                                    <th className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-2 px-4 text-center">
                                        Allowance
                                    </th>
                                    <th className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-2 px-4 text-center">
                                        Extra Days
                                    </th>
                                    <th className="text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-2 px-4 text-center">
                                        Status
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {sortedMembers.map((member) => {
                                    const displayName =
                                        memberNames[member.user_id] ??
                                        member.user_id;
                                    const initials = displayName
                                        .split(/[\s_]/)
                                        .map((p) => p[0]?.toUpperCase() ?? "")
                                        .slice(0, 2)
                                        .join("");

                                    return (
                                        <tr
                                            key={member.user_id}
                                            className={`rounded-xl ${
                                                member.is_over_limit
                                                    ? "bg-red-50/60"
                                                    : "bg-white/40"
                                            }`}
                                        >
                                            <td className="py-3 px-4 rounded-l-xl">
                                                <div className="flex items-center gap-3">
                                                    <div
                                                        className={`w-9 h-9 rounded-full flex items-center justify-center text-xs font-bold shrink-0 ${
                                                            member.is_over_limit
                                                                ? "bg-red-100 text-red-700"
                                                                : "bg-orange-100 text-orange-700"
                                                        }`}
                                                    >
                                                        {initials}
                                                    </div>
                                                    <span className="font-bold text-sm text-[var(--color-background-dark)] break-all">
                                                        {displayName}
                                                    </span>
                                                </div>
                                            </td>

                                            <td className="py-3 px-4 text-center">
                                                <span className="font-bold text-sm text-[var(--color-background-dark)]">
                                                    {member.wfh_days}
                                                </span>
                                            </td>

                                            <td className="py-3 px-4 text-center">
                                                <span className="text-sm font-medium text-[var(--color-text-sub)]">
                                                    {report.allowance}
                                                </span>
                                            </td>

                                            <td className="py-3 px-4 text-center">
                                                {member.extra_days > 0 ? (
                                                    <span className="inline-flex items-center gap-1 text-xs font-bold text-red-600 bg-red-50 border border-red-200 rounded-lg px-2 py-0.5">
                                                        +{member.extra_days}
                                                    </span>
                                                ) : (
                                                    <span className="text-sm text-[var(--color-text-sub)] opacity-50">
                                                        —
                                                    </span>
                                                )}
                                            </td>

                                            <td className="py-3 px-4 rounded-r-xl text-center">
                                                {member.is_over_limit ? (
                                                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold bg-red-50 text-red-700 border border-red-200">
                                                        <span className="material-symbols-outlined text-[14px]">
                                                            warning
                                                        </span>
                                                        Over
                                                    </span>
                                                ) : (
                                                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold bg-green-50 text-green-700 border border-green-200">
                                                        <span className="material-symbols-outlined text-[14px]">
                                                            check_circle
                                                        </span>
                                                        OK
                                                    </span>
                                                )}
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>

                    <div className="border-t border-[#e6dccf] pt-4 flex justify-end">
                        <button
                            type="button"
                            onClick={handleCopy}
                            className={`px-4 py-2.5 rounded-xl border text-sm font-semibold transition-all flex items-center gap-2 ${
                                copied
                                    ? "bg-green-50 border-green-200 text-green-700"
                                    : "bg-[var(--color-background-light)] border-[#e6dccf] text-[var(--color-text-sub)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-text-main)] active:shadow-[var(--shadow-clay-button-active)]"
                            }`}
                        >
                            <span className="material-symbols-outlined text-[16px]">
                                {copied ? "check" : "content_copy"}
                            </span>
                            {copied ? "Copied!" : "Copy Report"}
                        </button>
                    </div>
                </>
            )}
        </div>
    );
};

/* ─── Small helper component ─── */
interface StatPillProps {
    icon: string;
    label: string;
    value: string;
    accent: boolean;
}

const StatPill: React.FC<StatPillProps> = ({ icon, label, value, accent }) => (
    <div
        className={`flex flex-col gap-1.5 p-4 rounded-2xl border shadow-[var(--shadow-clay-button)] ${
            accent
                ? "bg-red-50 border-red-200"
                : "bg-[var(--color-background-light)] border-[#e6dccf]"
        }`}
    >
        <div className="flex items-center gap-1.5">
            <span
                className={`material-symbols-outlined text-[16px] ${
                    accent ? "text-red-500" : "text-[var(--color-text-sub)]"
                }`}
            >
                {icon}
            </span>
            <span
                className={`text-[10px] font-bold uppercase tracking-widest ${
                    accent ? "text-red-500" : "text-[var(--color-text-sub)]"
                }`}
            >
                {label}
            </span>
        </div>
        <p
            className={`text-base font-black tracking-tight ${
                accent ? "text-red-700" : "text-[var(--color-background-dark)]"
            }`}
        >
            {value}
        </p>
    </div>
);
