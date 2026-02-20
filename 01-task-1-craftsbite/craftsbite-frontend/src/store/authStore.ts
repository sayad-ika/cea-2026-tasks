// Authentication Store using Zustand

import { create } from 'zustand';
import type { AuthState, User } from '../types';
import { getUser, setUser as setUserStorage, removeUser } from '../utils/storage';

export const useAuthStore = create<AuthState>((set, get) => ({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: true,

    // Initialize auth state from localStorage
    initialize: () => {
        const user = getUser();

        if (user) {
            set({
                user,
                isAuthenticated: true,
                isLoading: false,
            });
        } else {
            set({
                user: null,
                isAuthenticated: false,
                isLoading: false,
            });
        }
    },

    // Login action
    login: async (_email: string, _password: string) => {
        set({ isLoading: true });

        try {
            // This will be implemented when we create authService
            // For now, this is a placeholder
            throw new Error('Login service not yet implemented');
        } catch (error) {
            set({ isLoading: false });
            throw error;
        }
    },

    // Register action
    register: async (_name: string, _email: string, _role: any, _password: string) => {
        set({ isLoading: true });
        try {
            throw new Error('Register service not yet implemented');
        } catch (error) {
            set({ isLoading: false });
            throw error;
        }
    },

    // Logout action
    logout: () => {
        removeUser();
        set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
        });
    },

    // Set user and token (called after successful login)
    setUser: (user: User | null) => {
        if (user) {
            setUserStorage(user);
            set({
                user,
                isAuthenticated: true,
            });
        } else {
            removeUser();
            set({
                user: null,
                isAuthenticated: false,
            });
        }
    },

    // Check authentication status
    checkAuth: async () => {
        const user = getUser();

        if (!user) {
            get().logout();
            return;
        }

        set({
            user,
            isAuthenticated: true,
            isLoading: false,
        });
    },
}));
