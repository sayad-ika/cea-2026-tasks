import { Link, useLocation, useNavigate } from 'react-router-dom';

function RoutingBasics() {
  const location = useLocation();
  const navigate = useNavigate();

  const goToInvoice = () => {
    // Simulates a support or notifications flow deep-linking to a specific entity.
    navigate('/params/8472?tab=billing');
  };

  const goToAdminGuardDemo = () => {
    // Demonstrates protected page access when typed directly in the URL bar.
    navigate('/admin');
  };

  return (
    <div>
      <h1>🧭 Routing Basics</h1>
      <div style={styles.card}>
        <p style={styles.helperText}>
          This page is about production-style routing decisions, not only "how
          to click links". Think of routes as product contracts: stable URLs,
          shareable state, and predictable access control.
        </p>

        <div style={styles.urlBox}>
          <p>
            <strong>Current URL:</strong> <code>{location.pathname}</code>
          </p>
          <p>
            <strong>Full location:</strong>{' '}
            <code>
              {location.pathname}
              {location.search}
              {location.hash}
            </code>
          </p>
        </div>

        <h3 style={styles.sectionTitle}>Real-world routing examples</h3>
        <div style={styles.grid}>
          <div style={styles.exampleCard}>
            <h4>SaaS dashboard routes</h4>
            <p>
              Keep URL hierarchy aligned to product areas so onboarding is
              easier.
            </p>
            <pre style={styles.pre}>{`/dashboard
/projects
/projects/:projectId
/projects/:projectId/settings`}</pre>
          </div>

          <div style={styles.exampleCard}>
            <h4>Deep-link support workflow</h4>
            <p>
              Support can paste a URL to open the exact tab a user is seeing.
            </p>
            <pre style={styles.pre}>{`/params/8472?tab=billing
/params/8472?tab=activity`}</pre>
          </div>

          <div style={styles.exampleCard}>
            <h4>Access control by route</h4>
            <p>
              Public URLs stay open. Sensitive routes must pass through route
              guards.
            </p>
            <pre style={styles.pre}>{`/demo        -> public login page
/admin       -> protected route
/storage     -> concept page (public)`}</pre>
          </div>

          <div style={styles.exampleCard}>
            <h4>Predictable fallback behavior</h4>
            <p>Unknown paths should render a useful 404, not a blank screen.</p>
            <pre
              style={styles.pre}
            >{`<Route path="*" element={<NotFound />} />`}</pre>
          </div>
        </div>

        <h3 style={styles.sectionTitle}>
          How navigation works in React Router v6
        </h3>
        <p>Build intuition for when to use each API:</p>
        <ul>
          <li>
            <code>&lt;Link to="/path" /&gt;</code>: user-driven navigation
            (menus, cards, breadcrumbs).
          </li>
          <li>
            <code>useNavigate()</code>: event-driven navigation (login success,
            wizard next step, retry actions).
          </li>
          <li>
            <code>useLocation()</code>: read current route for analytics, active
            states, and debug tools.
          </li>
          <li>
            <code>Navigate</code> component: guard/redirect logic in render tree
            (auth and role checks).
          </li>
        </ul>

        <h3 style={styles.sectionTitle}>Interactive scenarios</h3>
        <div style={styles.demoBox}>
          <h4 style={styles.demoHeading}>Scenario 1: Notification deep-link</h4>
          <p>
            Simulate clicking "Invoice overdue" notification and opening a
            specific entity + tab.
          </p>
          <button style={styles.button} onClick={goToInvoice}>
            Open invoice route (/params/8472?tab=billing)
          </button>
        </div>

        <div style={styles.demoBox}>
          <h4 style={styles.demoHeading}>
            Scenario 2: Manual URL access to protected page
          </h4>
          <p>
            Simulate a user typing <code>/admin</code> directly. Route guard
            decides access.
          </p>
          <button style={styles.button} onClick={goToAdminGuardDemo}>
            Try /admin now
          </button>
          <p style={styles.smallText}>
            If logged out, you should be redirected to <code>/demo</code>.
          </p>
        </div>

        <div style={styles.demoBox}>
          <h4 style={styles.demoHeading}>
            Scenario 3: Regular link navigation
          </h4>
          <p>All links below use client-side routing (no full page refresh):</p>
          <div style={styles.linksRow}>
            <Link to="/routing">Stay on routing</Link>
            <Link to="/params/99?tab=profile">Go to params demo</Link>
            <Link to="/demo">Go to demo app</Link>
            <Link to="/xss">Go to security demo</Link>
          </div>
        </div>

        <div style={styles.tipBox}>
          <strong>Team guideline:</strong> Put shareable UI state in the URL
          (path/query), and keep ephemeral state in memory. This improves
          supportability, QA reproducibility, and analytics quality.
        </div>
      </div>
    </div>
  );
}

const styles = {
  card: {
    background: '#f8f9fa',
    padding: '1.5rem',
    borderRadius: '8px',
    border: '1px solid #dfe3e8',
  },
  helperText: {
    marginBottom: '1rem',
    color: '#334155',
    lineHeight: '1.5',
  },
  sectionTitle: {
    marginTop: '1.1rem',
    marginBottom: '0.6rem',
  },
  urlBox: {
    background: '#eef2ff',
    border: '1px solid #c7d2fe',
    borderRadius: '6px',
    padding: '0.8rem',
    marginBottom: '1rem',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(230px, 1fr))',
    gap: '0.8rem',
    marginBottom: '1rem',
  },
  exampleCard: {
    background: '#ffffff',
    border: '1px solid #e5e7eb',
    borderRadius: '6px',
    padding: '0.85rem',
    lineHeight: '1.45',
  },
  pre: {
    marginTop: '0.55rem',
    background: '#111827',
    color: '#f3f4f6',
    borderRadius: '6px',
    padding: '0.6rem',
    fontSize: '0.8rem',
    overflowX: 'auto',
  },
  demoBox: {
    background: '#eef2f7',
    border: '1px solid #d5dee8',
    padding: '1rem',
    marginTop: '1rem',
    borderRadius: '6px',
  },
  demoHeading: {
    marginBottom: '0.45rem',
  },
  linksRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '0.9rem',
    marginTop: '0.35rem',
  },
  button: {
    marginTop: '0.35rem',
    background: '#0f766e',
    color: '#ffffff',
    border: 'none',
    borderRadius: '6px',
    padding: '0.5rem 0.75rem',
    cursor: 'pointer',
  },
  smallText: {
    marginTop: '0.45rem',
    color: '#475569',
    fontSize: '0.9rem',
  },
  tipBox: {
    marginTop: '1rem',
    background: '#fff7ed',
    border: '1px solid #fed7aa',
    borderRadius: '6px',
    padding: '0.85rem',
    lineHeight: '1.5',
  },
};

export default RoutingBasics;
