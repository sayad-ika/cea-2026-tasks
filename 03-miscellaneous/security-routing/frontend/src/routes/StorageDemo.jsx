import { useState, useEffect } from 'react';

function StorageDemo() {
  // LocalStorage demo
  const [localToken, setLocalToken] = useState('');
  const [cookieToken, setCookieToken] = useState('');
  const [memoryToken, setMemoryToken] = useState('');

  // Load initial values
  useEffect(() => {
    const saved = localStorage.getItem('demo_token');
    if (saved) setLocalToken(saved);
    // Read mock cookie (document.cookie)
    const match = document.cookie.match(/demo_cookie_token=([^;]+)/);
    if (match) setCookieToken(match[1]);
  }, []);

  const saveToLocalStorage = () => {
    const token = 'ls_token_' + Date.now();
    localStorage.setItem('demo_token', token);
    setLocalToken(token);
  };

  const saveToCookie = () => {
    const token = 'cookie_token_' + Date.now();
    document.cookie = `demo_cookie_token=${token}; path=/; max-age=3600`;
    setCookieToken(token);
  };

  const saveToMemory = () => {
    setMemoryToken('memory_token_' + Date.now());
  };

  const clearAll = () => {
    localStorage.removeItem('demo_token');
    document.cookie = 'demo_cookie_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    setLocalToken('');
    setCookieToken('');
    setMemoryToken('');
  };

  return (
    <div>
      <h1>💾 Token Storage: localStorage vs Cookie vs Memory</h1>
      <div style={styles.card}>
        <p>
          Different storage mechanisms for tokens (JWT, session IDs) have security and persistence trade-offs.
        </p>

        <div style={styles.grid}>
          <div style={styles.box}>
            <h3>📦 localStorage</h3>
            <p>Persists across browser restarts. Accessible via any JS (XSS risk).</p>
            <button onClick={saveToLocalStorage} style={styles.btn}>Save Demo Token</button>
            <div><strong>Stored token:</strong> {localToken || '(empty)'}</div>
          </div>

          <div style={styles.box}>
            <h3>🍪 Cookie (HttpOnly recommended but demo)</h3>
            <p>Sent automatically with requests. Can be flagged HttpOnly (prevents JS access) but vulnerable to CSRF.</p>
            <button onClick={saveToCookie} style={styles.btn}>Save Demo Cookie</button>
            <div><strong>Cookie token:</strong> {cookieToken || '(empty)'}</div>
          </div>

          <div style={styles.box}>
            <h3>🧠 In-Memory (React state)</h3>
            <p>Not persistent - lost on page reload. Most secure against XSS but user must re-authenticate.</p>
            <button onClick={saveToMemory} style={styles.btn}>Save Memory Token</button>
            <div><strong>Memory token:</strong> {memoryToken || '(empty)'}</div>
          </div>
        </div>

        <button onClick={clearAll} style={styles.clearBtn}>Clear All Storage</button>

        <div style={styles.refreshNote}>
          <p>🔄 <strong>Try refreshing the page:</strong></p>
          <ul>
            <li>localStorage token persists ✅</li>
            <li>Cookie token persists ✅ (if not expired)</li>
            <li>Memory token resets to empty ❌</li>
          </ul>
          <p>🔐 Security tip: For SPAs, many apps store tokens in memory + refresh token in httpOnly cookie.</p>
        </div>
      </div>
    </div>
  );
}

const styles = {
  card: { background: '#f8f9fa', padding: '1.5rem', borderRadius: '8px' },
  grid: { display: 'flex', gap: '1rem', flexWrap: 'wrap', marginBottom: '1rem' },
  box: { flex: '1', background: '#fff', padding: '1rem', borderRadius: '6px', border: '1px solid #ced4da' },
  btn: { margin: '0.5rem 0', padding: '0.3rem 0.8rem', cursor: 'pointer', background: '#007bff', color: 'white', border: 'none', borderRadius: '4px' },
  clearBtn: { marginTop: '1rem', padding: '0.5rem 1rem', background: '#6c757d', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' },
  refreshNote: { background: '#e9ecef', padding: '1rem', marginTop: '1rem', borderRadius: '6px' },
};

export default StorageDemo;