// Date Utility Functions

import { format, isToday as isTodayFns, parseISO, startOfDay, endOfDay, eachDayOfInterval, isBefore } from 'date-fns';
import { DATE_FORMATS } from './constants';

/**
 * Format a date string or Date object to display format
 */
export function formatDate(date: string | Date, formatStr: string = DATE_FORMATS.DISPLAY): string {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    return format(dateObj, formatStr);
}

/**
 * Format a time string to display format
 */
export function formatTime(time: string): string {
    const [hours, minutes] = time.split(':');
    const date = new Date();
    date.setHours(parseInt(hours, 10), parseInt(minutes, 10));
    return format(date, DATE_FORMATS.TIME);
}

/**
 * Check if a date is today
 */
export function isToday(date: string | Date): boolean {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    return isTodayFns(dateObj);
}

/**
 * Check if current time is before cutoff time
 */
export function isBeforeCutoff(cutoffTime: string): boolean {
    const now = new Date();
    const [hours, minutes] = cutoffTime.split(':');
    const cutoff = new Date();
    cutoff.setHours(parseInt(hours, 10), parseInt(minutes, 10), 0, 0);
    return isBefore(now, cutoff);
}

/**
 * Get an array of dates between start and end dates (inclusive)
 */
export function getDateRange(startDate: string | Date, endDate: string | Date): Date[] {
    const start = typeof startDate === 'string' ? parseISO(startDate) : startDate;
    const end = typeof endDate === 'string' ? parseISO(endDate) : endDate;

    return eachDayOfInterval({
        start: startOfDay(start),
        end: endOfDay(end),
    });
}

/**
 * Format date to API format (YYYY-MM-DD)
 */
export function toApiDate(date: Date | string): string {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    return format(dateObj, DATE_FORMATS.API);
}

/**
 * Get today's date in API format
 */
export function getTodayApiDate(): string {
    return format(new Date(), DATE_FORMATS.API);
}

/**
 * Parse API date string to Date object
 */
export function fromApiDate(dateString: string): Date {
    return parseISO(dateString);
}
