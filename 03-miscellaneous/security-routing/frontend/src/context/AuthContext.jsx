import { createContext, useContext, useState } from 'react';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  // Mock auth state - stored in React state (in-memory, not persistent)
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState(null);

  // Fake login function (no backend)
  const login = (email, password) => {
    // Simulate API call - accept any non-empty credentials
    if (email && password) {
      setIsLoggedIn(true);
      setToken('fake-jwt-token-' + Date.now());
      return true;
    }
    return false;
  };

  const logout = () => {
    setIsLoggedIn(false);
    setToken(null);
  };

  const value = {
    isLoggedIn,
    token,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Custom hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}