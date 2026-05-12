import { createContext, useContext, useState, useEffect, useRef } from 'react';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [user, setUser] = useState(null);
  const [csrfToken, setCsrfToken] = useState(null);
  const [csrfLoading, setCsrfLoading] = useState(false);
  const [loading, setLoading] = useState(true);
  const initialised = useRef(false);

  const fetchCSRFToken = async () => {
    setCsrfLoading(true);
    try {
      const csrfRes = await fetch('/csrf-token', { credentials: 'include' });
      if (!csrfRes.ok) throw new Error('Failed to fetch CSRF token');
      const { csrfToken: token } = await csrfRes.json();
      setCsrfToken(token);
      return token;
    } finally {
      setCsrfLoading(false);
    }
  };

  useEffect(() => {
    if (initialised.current) return;
    initialised.current = true;

    fetch('/api/me', { credentials: 'include' })
      .then((res) => {
        if (!res.ok) throw new Error();
        return res.json();
      })
      .then((data) => {
        setIsLoggedIn(true);
        setUser(data);
        return fetchCSRFToken();
      })
      .catch(() => {
        setIsLoggedIn(false);
        setUser(null);
        setCsrfToken(null);
      })
      .finally(() => setLoading(false));
  }, []);

  const login = async (email, password) => {
    const token = csrfToken || (await fetchCSRFToken());

    const loginRes = await fetch('/login', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': token,
      },
      body: JSON.stringify({ email, password }),
    });

    if (!loginRes.ok) {
      const err = await loginRes.json();
      throw new Error(err.error || 'Login failed');
    }

    const data = await loginRes.json();
    setIsLoggedIn(true);
    setUser({ name: data.name });
    setCsrfToken(data.csrfToken);
    return true;
  };

  const logout = async () => {
    await fetch('/logout', {
      method: 'POST',
      credentials: 'include',
    });
    setIsLoggedIn(false);
    setUser(null);
    setCsrfToken(null);
  };

  const value = {
    isLoggedIn,
    user,
    csrfToken,
    csrfLoading,
    loading,
    prepareLogin: fetchCSRFToken,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
