// Local Storage Utility Functions

import type { User } from '../types';
import { STORAGE_KEYS } from './constants';

/**
 * Get user data from localStorage
 */
export function getUser(): User | null {
    const userJson = localStorage.getItem(STORAGE_KEYS.USER);
    if (!userJson) return null;

    try {
        return JSON.parse(userJson) as User;
    } catch (error) {
        console.error('Failed to parse user data from localStorage:', error);
        return null;
    }
}

/**
 * Set user data in localStorage
 */
export function setUser(user: User): void {
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
}

/**
 * Remove user data from localStorage
 */
export function removeUser(): void {
    localStorage.removeItem(STORAGE_KEYS.USER);
}

/**
 * Clear all auth data from localStorage
 */
export function clearAuthData(): void {
    removeUser();
}
