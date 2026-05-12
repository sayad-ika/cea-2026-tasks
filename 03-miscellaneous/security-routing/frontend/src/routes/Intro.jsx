import { Link } from 'react-router-dom';

const routes = [
  {
    path: '/',
    label: 'Intro',
    badge: 'Concept',
    time: '0–5 min',
    desc: 'Session overview and app structure.',
  },
  {
    path: '/routing',
    label: 'Routing Basics',
    badge: 'Concept',
    time: '5–15 min',
    desc: 'BrowserRouter, Link, useLocation, client-side navigation.',
  },
  {
    path: '/params/42',
    label: 'Params & Query',
    badge: 'Demo',
    time: '15–25 min',
    desc: 'useParams (:userId) and useSearchParams (?tab=).',
  },
  {
    path: '/protected',
    label: 'Protected Routes',
    badge: 'Concept',
    time: '25–35 min',
    desc: 'ProtectedRoute component, Navigate redirect pattern.',
  },
  {
    path: '/demo',
    label: 'Demo App',
    badge: 'Demo',
    time: '35–45 min',
    desc: 'Mock login, useNavigate after login, /admin redirect.',
  },
  {
    path: '/xss',
    label: 'XSS + CSRF',
    badge: 'Security',
    time: '45–55 min',
    desc: 'dangerouslySetInnerHTML vs safe rendering, CSRF token sim.',
  },
  {
    path: '/storage',
    label: 'Token Storage',
    badge: 'Security',
    time: '55–60 min',
    desc: 'localStorage vs cookie vs in-memory: trade-offs.',
  },
  {
    path: '/admin',
    label: 'Admin (protected)',
    badge: 'Auth',
    time: '—',
    desc: 'Only reachable when logged in. Try it without logging in!',
  },
];

const badgeColors = {
  Concept: { background: '#dbeafe', color: '#1e40af' },
  Demo: { background: '#dcfce7', color: '#166534' },
  Security: { background: '#fee2e2', color: '#991b1b' },
  Auth: { background: '#fef3c7', color: '#92400e' },
};

function Intro() {
  return (
    <div>
      <h1>📖 Frontend Training: Routing & Security</h1>
      <div style={styles.card}>
        <p style={{ marginBottom: '0.75rem' }}>
          This interactive app teaches React Router v6 and frontend security
          through live demos. Use the <strong>navbar above</strong> to navigate
          — each page is one "slide".
        </p>
        <p>
          🔐 The <strong>Admin</strong> page is protected — try visiting it
          without logging in!
        </p>
      </div>

      <h2 style={{ marginTop: '1.5rem', marginBottom: '0.75rem' }}>
        📋 Route Map
      </h2>
      <table style={styles.table}>
        <thead>
          <tr style={styles.thead}>
            <th style={styles.th}>Path</th>
            <th style={styles.th}>Page</th>
            <th style={styles.th}>Type</th>
            <th style={styles.th}>Time</th>
            <th style={styles.th}>What you'll learn</th>
          </tr>
        </thead>
        <tbody>
          {routes.map((r) => (
            <tr key={r.path} style={styles.tr}>
              <td style={styles.tdMono}>
                <Link to={r.path} style={styles.pathLink}>
                  {r.path}
                </Link>
              </td>
              <td style={styles.td}>{r.label}</td>
              <td style={styles.td}>
                <span style={{ ...styles.badge, ...badgeColors[r.badge] }}>
                  {r.badge}
                </span>
              </td>
              <td
                style={{ ...styles.td, whiteSpace: 'nowrap', color: '#6b7280' }}
              >
                {r.time}
              </td>
              <td style={styles.td}>{r.desc}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const styles = {
  card: {
    background: '#f8f9fa',
    padding: '1.25rem',
    borderRadius: '8px',
    border: '1px solid #dee2e6',
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse',
    background: '#fff',
    border: '1px solid #dee2e6',
    borderRadius: '8px',
    overflow: 'hidden',
    fontSize: '0.9rem',
  },
  thead: { background: '#2c3e50' },
  th: {
    color: 'white',
    padding: '0.65rem 0.9rem',
    textAlign: 'left',
    fontWeight: 600,
  },
  tr: { borderBottom: '1px solid #dee2e6' },
  td: { padding: '0.6rem 0.9rem', verticalAlign: 'top' },
  tdMono: {
    padding: '0.6rem 0.9rem',
    fontFamily: 'monospace',
    fontSize: '0.85rem',
  },
  pathLink: { color: '#1d4ed8', textDecoration: 'none', fontWeight: 500 },
  badge: {
    display: 'inline-block',
    padding: '2px 8px',
    borderRadius: '12px',
    fontSize: '0.75rem',
    fontWeight: 600,
    whiteSpace: 'nowrap',
  },
};

export default Intro;
