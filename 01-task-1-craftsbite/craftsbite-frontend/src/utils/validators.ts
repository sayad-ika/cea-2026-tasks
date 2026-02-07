// Validation Utility Functions

import { parseISO, isBefore, isAfter } from 'date-fns';

/**
 * Validate email format
 */
export function validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

/**
 * Validate password strength
 * Requirements: At least 8 characters
 */
export function validatePassword(password: string): {
    isValid: boolean;
    errors: string[];
} {
    const errors: string[] = [];

    if (password.length < 8) {
        errors.push('Password must be at least 8 characters long');
    }

    if (!/[A-Z]/.test(password)) {
        errors.push('Password must contain at least one uppercase letter');
    }

    if (!/[a-z]/.test(password)) {
        errors.push('Password must contain at least one lowercase letter');
    }

    if (!/[0-9]/.test(password)) {
        errors.push('Password must contain at least one number');
    }

    return {
        isValid: errors.length === 0,
        errors,
    };
}

/**
 * Validate date range
 */
export function validateDateRange(startDate: string, endDate: string): {
    isValid: boolean;
    error?: string;
} {
    try {
        const start = parseISO(startDate);
        const end = parseISO(endDate);

        if (isAfter(start, end)) {
            return {
                isValid: false,
                error: 'Start date must be before or equal to end date',
            };
        }

        return { isValid: true };
    } catch (error) {
        return {
            isValid: false,
            error: 'Invalid date format',
        };
    }
}

/**
 * Validate required field
 */
export function validateRequired(value: string | null | undefined): boolean {
    return value !== null && value !== undefined && value.trim().length > 0;
}

/**
 * Validate date is not in the past
 */
export function validateFutureDate(date: string): {
    isValid: boolean;
    error?: string;
} {
    try {
        const dateObj = parseISO(date);
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        if (isBefore(dateObj, today)) {
            return {
                isValid: false,
                error: 'Date cannot be in the past',
            };
        }

        return { isValid: true };
    } catch (error) {
        return {
            isValid: false,
            error: 'Invalid date format',
        };
    }
}
