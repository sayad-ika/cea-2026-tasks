import React, { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import type { AuthState, User } from '../types';
import * as authService from '../services/authService';
import { getUser, setUser as saveUser, clearAuthData } from '../utils/storage';

const AuthContext = createContext<AuthState | undefined>(undefined);

interface AuthProviderProps {
    children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUserState] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    const initialize = async () => {
        const storedUser = getUser();

        if (storedUser) {
            setUserState(storedUser);
        }

        try {
            const currentUser = await authService.getCurrentUser();
            saveUser(currentUser);
            setUserState(currentUser);
        } catch {
            clearAuthData();
            setUserState(null);
        } finally {
            setIsLoading(false);
        }
    };

    // Check authentication status
    const checkAuth = async () => {
        try {
            const currentUser = await authService.getCurrentUser();
            saveUser(currentUser);
            setUserState(currentUser);
        } catch (error) {
            clearAuthData();
            setUserState(null);
        }
    };

    // Login function
    const login = async (email: string, password: string) => {
        setIsLoading(true);
        try {
            await authService.login(email, password);

            const fullUser = await authService.getCurrentUser();

            saveUser(fullUser);
            setUserState(fullUser);
        } catch (error) {
            clearAuthData();
            setUserState(null);
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    // Logout function
    const logout = async () => {
        setIsLoading(true);
        try {
            await authService.logout();
        } catch (error) {
            throw error;
        } finally {
            clearAuthData();
            setUserState(null);
            setIsLoading(false);
        }
    };

    // Set user manually
    const setUser = (newUser: User | null) => {
        setUserState(newUser);
        if (newUser) {
            saveUser(newUser);
        }   else    {
            clearAuthData();
        }
    };

    // Initialize on mount
    useEffect(() => {
        initialize();
    }, []);

    // Register function
    const register = async (name: string, email: string, role: any, password: string) => {
        setIsLoading(true);
        try {
            await authService.register(name, email, role, password);

            // Fetch full user details since login response might be partial
            const fullUser = await authService.getCurrentUser();

            // Save full user data
            saveUser(fullUser);
            setUserState(fullUser);
        } catch (error) {
            // Clean up if getting user fails
            clearAuthData();
            setUserState(null);
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const value: AuthState = {
        user,
        token: null,             
        isAuthenticated: !!user, // ← was: !!token && !!user — cookie isn't visible so just check user
        isLoading,
        login,
        register,
        logout,
        setUser,
        checkAuth,
        initialize,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};


// Custom hook to use auth context
export const useAuth = (): AuthState => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
