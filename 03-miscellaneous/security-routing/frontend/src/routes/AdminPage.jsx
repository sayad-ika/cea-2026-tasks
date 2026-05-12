import { useAuth } from '../context/AuthContext';

function AdminPage() {
  const { token } = useAuth();
  return (
    <div>
      <h1>👑 Admin Panel (Protected Route)</h1>
      <div style={styles.card}>
        <p>You have accessed the protected admin area because you are authenticated.</p>
        <div style={styles.infoBox}>
          <strong>Current session token:</strong> <code>{token}</code>
        </div>
        <h3>Admin Actions (Demo)</h3>
        <ul>
          <li>📋 Manage users (simulated)</li>
          <li>⚙️ System configuration</li>
          <li>📊 View analytics</li>
        </ul>
        <p style={{ marginTop: '1rem', fontSize: '0.9rem', background: '#fff3cd', padding: '0.5rem' }}>
          🔐 This page is protected by <code>ProtectedRoute</code>. Try logging out from the Demo App and refresh.
        </p>
      </div>
    </div>
  );
}

const styles = {
  card: { background: '#f8f9fa', padding: '1.5rem', borderRadius: '8px' },
  infoBox: { background: '#e2e3e5', padding: '0.75rem', borderRadius: '4px', margin: '1rem 0' },
};

export default AdminPage;