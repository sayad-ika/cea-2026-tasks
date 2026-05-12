import { Link } from 'react-router-dom';

function ProtectedRoutesExplanation() {
  return (
    <div>
      <h1>🛡️ Protected Routes Explanation</h1>
      <div style={styles.card}>
        <p>
          A <strong>protected route</strong> only allows access when the user is authenticated.
          Otherwise, it redirects to a login page.
        </p>
        <h3>Implementation pattern (React Router v6):</h3>
        <pre style={styles.codeBlock}>
{`// ProtectedRoute component
function ProtectedRoute({ children }) {
  const { isLoggedIn } = useAuth();
  if (!isLoggedIn) return <Navigate to="/login" />;
  return children;
}

// Usage in route config:
<Route path="/admin" element={
  <ProtectedRoute>
    <AdminPage />
  </ProtectedRoute>
} />`}
        </pre>
        <div style={styles.demoBox}>
          <h4>🔐 Live example:</h4>
          <p>
            The <strong>/admin</strong> route in this app is protected.
            <br />
            👉 <Link to="/admin">Try accessing /admin now</Link> (will redirect if not logged in).
          </p>
          <p>
            👉 Go to <Link to="/demo">/demo page</Link> to log in first, then try /admin again.
          </p>
        </div>
      </div>
    </div>
  );
}

const styles = {
  card: { background: '#f8f9fa', padding: '1.5rem', borderRadius: '8px' },
  codeBlock: {
    background: '#2d2d2d',
    color: '#f8f8f2',
    padding: '1rem',
    borderRadius: '6px',
    overflowX: 'auto',
    fontFamily: 'monospace',
  },
  demoBox: { background: '#e9ecef', padding: '1rem', marginTop: '1rem', borderRadius: '4px' },
};

export default ProtectedRoutesExplanation;