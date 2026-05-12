import { useAuth } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';
import { useEffect, useRef, useState } from 'react';

function DemoApp() {
  const { isLoggedIn, user, csrfToken, csrfLoading, prepareLogin, login, logout, loading } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [loggingIn, setLoggingIn] = useState(false);
  const [csrfResult, setCsrfResult] = useState(null);
  const preparedLogin = useRef(false);

  useEffect(() => {
    if (loading || isLoggedIn || csrfToken || preparedLogin.current) return;
    preparedLogin.current = true;

    prepareLogin().catch((err) => {
      setLoginError(err.message || 'Failed to prepare secure login');
      preparedLogin.current = false;
    });
  }, [csrfToken, isLoggedIn, loading, prepareLogin]);

  const handleLogin = async () => {
    setLoginError('');
    setLoggingIn(true);
    try {
      await login(email, password);
      navigate('/demo');
    } catch (err) {
      setLoginError(err.message || 'Login failed');
    } finally {
      setLoggingIn(false);
    }
  };

  if (loading) {
    return (
      <div>
        <h1>🎮 Demo App: Authentication + Protected Area</h1>
        <div style={styles.card}>
          <p>Checking session...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <h1>🎮 Demo App: Authentication + Protected Area</h1>
      <div style={styles.card}>
        {!isLoggedIn ? (
          <div style={styles.loginBox}>
            <h3>Login</h3>
            <p>Use one of the test accounts:</p>
            <ul style={{ fontSize: '0.85rem', color: '#475569' }}>
              <li><code>doha@craftsmensoftware.com</code> / <code>password123</code></li>
              <li><code>sayad.ibn@craftsmensoftware.com</code> / <code>letmein</code></li>
            </ul>
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={styles.input}
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={styles.input}
              onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
            />
            <button onClick={handleLogin} disabled={loggingIn} style={styles.button}>
              {loggingIn ? 'Logging in...' : 'Login'}
            </button>
            {csrfLoading && <p style={{ color: '#475569', marginTop: '0.5rem' }}>Preparing secure login...</p>}
            {loginError && <p style={{ color: 'red', marginTop: '0.5rem' }}>{loginError}</p>}
          </div>
        ) : (
          <div>
            <div style={styles.loggedBox}>
              <h3>✅ You are logged in!</h3>
              <p><strong>Name:</strong> {user?.name}</p>
              <button onClick={logout} style={styles.button}>Logout</button>
            </div>

            <hr />
            <h3>Dashboard (accessible to any logged-in user)</h3>
            <div style={styles.dashboard}>
              <p>📊 Welcome to your dashboard, {user?.name}.</p>
            </div>

            <hr />
            <h3>CSRF Protection Demo</h3>
            <div style={styles.csrfBox}>
              <p>Test that the backend enforces CSRF tokens on state-changing requests:</p>
              <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                <button
                  onClick={async () => {
                    try {
                      console.log('[CSRF PASS token]', csrfToken);
                      const res = await fetch('/api/csrf-test', {
                        method: 'POST',
                        credentials: 'include',
                        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                      });
                      const data = await res.json();
                      console.log('[CSRF PASS]', res.status, data);
                      setCsrfResult({ pass: true, status: res.status, data });
                    } catch (err) {
                      console.error('[CSRF PASS ERROR]', err);
                      setCsrfResult({ pass: true, status: 0, data: { error: err.message } });
                    }
                  }}
                  disabled={!csrfToken}
                  style={styles.csrfPassBtn}
                >
                  ✅ Test CSRF Pass (real token)
                </button>
                <button
                  onClick={async () => {
                    try {
                      const res = await fetch('/api/csrf-test', {
                        method: 'POST',
                        credentials: 'include',
                        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': 'bad-token' },
                      });
                      const data = await res.json();
                      console.log('[CSRF FAIL]', res.status, data);
                      setCsrfResult({ pass: false, status: res.status, data });
                    } catch (err) {
                      console.error('[CSRF FAIL ERROR]', err);
                      setCsrfResult({ pass: false, status: 0, data: { error: err.message } });
                    }
                  }}
                  style={styles.csrfFailBtn}
                >
                  ❌ Test CSRF Fail (bad token)
                </button>
              </div>
              {!csrfToken && (
                <p style={{ color: '#92400e', marginTop: '0.5rem' }}>
                  No CSRF token loaded yet. Try logging out and logging in again, or refresh after the backend is running.
                </p>
              )}
              {csrfResult && (
                <div style={{
                  ...styles.csrfResultBox,
                  background: csrfResult.pass && csrfResult.status === 200 ? '#d4edda' : '#f8d7da',
                  borderColor: csrfResult.pass && csrfResult.status === 200 ? '#c3e6cb' : '#f5c6cb',
                }}>
                  <strong>{csrfResult.pass && csrfResult.status === 200 ? '✅ PASS' : '❌ BLOCKED'}</strong>
                  <br />
                  Status: {csrfResult.status}
                  <br />
                  Response: <code>{JSON.stringify(csrfResult.data)}</code>
                </div>
              )}
            </div>

            <hr />
            <h3>Admin Section (Protected)</h3>
            <div style={styles.adminLinkBox}>
              <p>Admin page requires authentication (already satisfied).</p>
              <Link to="/admin" style={styles.linkButton}>Go to Admin Panel →</Link>
              <p style={{ fontSize: '0.9rem', marginTop: '0.5rem' }}>
                💡 Try manually typing <code>/admin</code> in URL while logged out → redirects here.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

const styles = {
  card: { background: '#f8f9fa', padding: '1.5rem', borderRadius: '8px' },
  loginBox: { padding: '1rem', background: '#e9ecef', borderRadius: '6px' },
  loggedBox: { background: '#d4edda', padding: '1rem', borderRadius: '6px' },
  dashboard: { background: '#fff', padding: '1rem', border: '1px solid #dee2e6', borderRadius: '6px' },
  adminLinkBox: { background: '#cfe2ff', padding: '1rem', borderRadius: '6px' },
  csrfBox: { background: '#fff', padding: '1rem', border: '1px solid #dee2e6', borderRadius: '6px' },
  csrfPassBtn: { padding: '0.5rem 1rem', cursor: 'pointer', background: '#28a745', color: '#fff', border: 'none', borderRadius: '4px' },
  csrfFailBtn: { padding: '0.5rem 1rem', cursor: 'pointer', background: '#dc3545', color: '#fff', border: 'none', borderRadius: '4px' },
  csrfResultBox: { marginTop: '0.75rem', padding: '0.75rem', borderRadius: '4px', border: '1px solid' },
  input: { display: 'block', margin: '0.5rem 0', padding: '0.5rem', width: '250px' },
  button: { padding: '0.5rem 1rem', cursor: 'pointer', background: '#007bff', color: '#fff', border: 'none', borderRadius: '4px' },
  linkButton: { display: 'inline-block', marginTop: '0.5rem', padding: '0.4rem 0.8rem', background: '#28a745', color: 'white', textDecoration: 'none', borderRadius: '4px' },
};

export default DemoApp;
