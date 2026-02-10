import React, { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import type { AuthState, User } from '../types';
import * as authService from '../services/authService';
import { getToken, setToken, getUser, setUser as saveUser, clearAuthData } from '../utils/storage';

const AuthContext = createContext<AuthState | undefined>(undefined);

interface AuthProviderProps {
    children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUserState] = useState<User | null>(null);
    const [token, setTokenState] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Initialize auth state from localStorage
    const initialize = () => {
        const storedToken = getToken();
        const storedUser = getUser();

        if (storedToken && storedUser) {
            setTokenState(storedToken);
            setUserState(storedUser);
        }

        setIsLoading(false);
    };

    // Check authentication status
    const checkAuth = async () => {
        try {
            const currentUser = await authService.getCurrentUser();
            setUserState(currentUser);
        } catch (error) {
            // If check fails, clear auth data
            clearAuthData();
            setUserState(null);
            setTokenState(null);
        }
    };

    // Login function
    const login = async (email: string, password: string) => {
        setIsLoading(true);
        try {
            const response = await authService.login(email, password);

            // Get token from the nested data object
            const { token } = response.data;

            // Save token immediately so subsequent requests work
            setToken(token);
            setTokenState(token);

            // Fetch full user details since login response might be partial
            const fullUser = await authService.getCurrentUser();

            // Save full user data
            saveUser(fullUser);
            setUserState(fullUser);
        } catch (error) {
            // Clean up if getting user fails
            clearAuthData();
            setTokenState(null);
            setUserState(null);
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    // Logout function
    const logout = () => {
        clearAuthData();
        setUserState(null);
        setTokenState(null);
    };

    // Set user manually
    const setUser = (newUser: User | null) => {
        setUserState(newUser);
        if (newUser) {
            saveUser(newUser);
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
            const response = await authService.register(name, email, role, password);

            // Get token from the nested data object
            const { token } = response.data;

            // Save token immediately so subsequent requests work
            setToken(token);
            setTokenState(token);

            // Fetch full user details since login response might be partial
            const fullUser = await authService.getCurrentUser();

            // Save full user data
            saveUser(fullUser);
            setUserState(fullUser);
        } catch (error) {
            // Clean up if getting user fails
            clearAuthData();
            setTokenState(null);
            setUserState(null);
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const value: AuthState = {
        user,
        token,
        isAuthenticated: !!token && !!user,
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
