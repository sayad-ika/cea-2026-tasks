import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';

interface ThemeContextValue {
    isDarkMode: boolean;
    toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue>({
    isDarkMode: false,
    toggleTheme: () => { },
});

const STORAGE_KEY = 'craftsbite_theme';

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [isDarkMode, setIsDarkMode] = useState<boolean>(() => {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored !== null) return stored === 'dark';
        return window.matchMedia('(prefers-color-scheme: dark)').matches;
    });

    // Apply / remove `.dark` on <html>
    useEffect(() => {
        const root = document.documentElement;
        if (isDarkMode) {
            root.classList.add('dark');
        } else {
            root.classList.remove('dark');
        }
        localStorage.setItem(STORAGE_KEY, isDarkMode ? 'dark' : 'light');
    }, [isDarkMode]);

    const toggleTheme = useCallback(() => setIsDarkMode((prev) => !prev), []);

    return (
        <ThemeContext.Provider value={{ isDarkMode, toggleTheme }}>
            {children}
        </ThemeContext.Provider>
    );
};

export const useTheme = () => useContext(ThemeContext);
