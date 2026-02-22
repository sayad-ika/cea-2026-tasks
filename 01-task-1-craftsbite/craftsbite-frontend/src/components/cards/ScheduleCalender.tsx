import React from "react";
import {
    format,
    startOfMonth,
    endOfMonth,
    startOfWeek,
    endOfWeek,
    addDays,
    isSameMonth,
    isToday,
    isWeekend,
} from "date-fns";
import { STATUS_COLORS_FOR_SCHEDULE, type ScheduleEntry } from "../../types/schedule.types";


const LEGEND = [
    { color: "bg-emerald-400", label: "Normal" },
    { color: "bg-red-700", label: "Office Closed" },
    { color: "bg-sky-400", label: "Govt Holiday" },
    { color: "bg-violet-400", label: "Celebration" },
    { color: "bg-orange-400", label: "Event Day" },
];

const DAY_HEADERS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

export interface ScheduleCalendarProps {
    currentMonth: Date;
    scheduleMap: Record<string, ScheduleEntry>;
    selectedDate: string;
    onSelectDate: (date: string) => void;
    onPrevMonth: () => void;
    onNextMonth: () => void;
}

export const ScheduleCalendar: React.FC<ScheduleCalendarProps> = ({
    currentMonth,
    scheduleMap,
    selectedDate,
    onSelectDate,
    onPrevMonth,
    onNextMonth,
}) => {
    const monthStart = startOfMonth(currentMonth);
    const monthEnd = endOfMonth(currentMonth);
    const gridStart = startOfWeek(monthStart, { weekStartsOn: 1 });
    const gridEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });

    const days: Date[] = [];
    let cur = gridStart;
    while (cur <= gridEnd) {
        days.push(cur);
        cur = addDays(cur, 1);
    }

    return (
        <div className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-6 shadow-[var(--shadow-clay-card)]">
            {/* Month navigation */}
            <div className="flex items-center justify-between mb-5">
                <button
                    onClick={onPrevMonth}
                    className="w-9 h-9 rounded-xl flex items-center justify-center bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                >
                    <span className="material-symbols-outlined text-[20px]">
                        chevron_left
                    </span>
                </button>

                <h3 className="text-xl font-black text-[var(--color-background-dark)] tracking-tight">
                    {format(currentMonth, "MMMM yyyy")}
                </h3>

                <button
                    onClick={onNextMonth}
                    className="w-9 h-9 rounded-xl flex items-center justify-center bg-[var(--color-background-light)] shadow-[var(--shadow-clay-button)] hover:text-[var(--color-primary)] active:shadow-[var(--shadow-clay-button-active)] transition-all"
                >
                    <span className="material-symbols-outlined text-[20px]">
                        chevron_right
                    </span>
                </button>
            </div>

            {/* Day-of-week headers */}
            <div className="grid grid-cols-7 mb-2">
                {DAY_HEADERS.map((d) => (
                    <div
                        key={d}
                        className="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-sub)] py-1"
                    >
                        {d}
                    </div>
                ))}
            </div>

            {/* Calendar grid */}
            <div className="grid grid-cols-7 gap-1">
                {days.map((day) => {
                    const key = format(day, "yyyy-MM-dd");
                    const entry = scheduleMap[key];
                    const inMonth = isSameMonth(day, currentMonth);
                    const todayDay = isToday(day);
                    const selected = key === selectedDate;
                    const weekend = isWeekend(day);
                    const status = entry?.day_status;

                    return (
                        <button
                            key={key}
                            onClick={() => inMonth && onSelectDate(key)}
                            disabled={!inMonth}
                            className={[
                                "relative flex flex-col items-center justify-start pt-1.5 pb-2 rounded-xl transition-all min-h-[52px]",
                                inMonth
                                    ? "cursor-pointer hover:bg-[#f5ede0]"
                                    : "opacity-0 pointer-events-none",
                                selected
                                    ? "bg-gradient-to-br from-[#fa8c47] to-[#e57a36] shadow-[var(--shadow-clay-button)] text-white"
                                    : weekend && inMonth
                                      ? "bg-[#f5f0e8]"
                                      : "",
                            ]
                                .filter(Boolean)
                                .join(" ")}
                        >
                            <span
                                className={[
                                    "text-sm font-bold leading-none",
                                    !inMonth
                                        ? "text-transparent"
                                        : selected
                                          ? "text-white"
                                          : todayDay
                                            ? "text-[var(--color-primary)]"
                                            : weekend
                                              ? "text-[var(--color-text-sub)]"
                                              : "text-[var(--color-background-dark)]",
                                ].join(" ")}
                            >
                                {format(day, "d")}
                            </span>

                            {/* Today dot */}
                            {todayDay && !selected && (
                                <span className="absolute top-1 right-1 w-1.5 h-1.5 rounded-full bg-[var(--color-primary)]" />
                            )}

                            {/* Schedule dot */}
                            {entry && inMonth && (
                                <span
                                    className={[
                                        "mt-1 w-1.5 h-1.5 rounded-full",
                                        selected
                                            ? "bg-white"
                                            : (STATUS_COLORS_FOR_SCHEDULE[status ?? "normal"] ??
                                              "bg-gray-400"),
                                    ].join(" ")}
                                />
                            )}
                        </button>
                    );
                })}
            </div>

            {/* Legend */}
            <div className="mt-4 pt-4 border-t border-[#e6dccf] flex flex-wrap gap-3">
                {LEGEND.map(({ color, label }) => (
                    <span
                        key={label}
                        className="flex items-center gap-1.5 text-[11px] font-medium text-[var(--color-text-sub)]"
                    >
                        <span className={`w-2 h-2 rounded-full ${color}`} />
                        {label}
                    </span>
                ))}
            </div>
        </div>
    );
};
