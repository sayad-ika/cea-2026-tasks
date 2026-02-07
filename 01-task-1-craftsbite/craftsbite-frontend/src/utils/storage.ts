// Local Storage Utility Functions

import type { User } from '../types';
import { STORAGE_KEYS } from './constants';

/**
 * Get JWT token from localStorage
 */
export function getToken(): string | null {
    return localStorage.getItem(STORAGE_KEYS.TOKEN);
}

/**
 * Set JWT token in localStorage
 */
export function setToken(token: string): void {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);
}

/**
 * Remove JWT token from localStorage
 */
export function removeToken(): void {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
}

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
    removeToken();
    removeUser();
}
