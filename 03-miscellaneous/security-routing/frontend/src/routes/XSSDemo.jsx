import { useState } from 'react';

function XSSDemo() {
  const [vulnerableInput, setVulnerableInput] = useState(
    '<b>Welcome back, Alex</b> <img src=x onerror="alert(\'XSS fired\')" />'
  );
  const [safeInput, setSafeInput] = useState(
    '<b>Welcome back, Alex</b> <img src=x onerror="alert(\'XSS fired\')" />'
  );
  const [securityLog, setSecurityLog] = useState([]);

  const [cookieSessionEnabled, setCookieSessionEnabled] = useState(true);
  const [sameSiteMode, setSameSiteMode] = useState('none');
  const [csrfTokenOnForm, setCsrfTokenOnForm] = useState(false);
  const [csrfTokenOnServer, setCsrfTokenOnServer] = useState('secure-123');
  const [requestResult, setRequestResult] = useState('No request yet');

  const appendLog = (line) => {
    setSecurityLog((prev) => [line, ...prev].slice(0, 7));
  };

  const usePayload = (payload) => {
    setVulnerableInput(payload);
    setSafeInput(payload);
  };

  const csrfProtectedTransfer = ({ crossSite, providedToken }) => {
    const browserSendsCookie = cookieSessionEnabled && (sameSiteMode === 'none' || !crossSite);
    const csrfTokenValid = providedToken && providedToken === csrfTokenOnServer;

    if (!browserSendsCookie) {
      return {
        ok: false,
        reason: 'Blocked: session cookie not sent (SameSite + cross-site request).',
      };
    }

    if (!csrfTokenValid) {
      return {
        ok: false,
        reason: 'Blocked: missing or invalid CSRF token.',
      };
    }

    return {
      ok: true,
      reason: 'Success: transfer accepted (cookie + valid CSRF token).',
    };
  };

  const runAttackSimulation = () => {
    const result = csrfProtectedTransfer({ crossSite: true, providedToken: null });
    setRequestResult(result.reason);
    appendLog(`[attack] POST /transfer cross-site -> ${result.reason}`);
  };

  const runLegitTransfer = () => {
    const providedToken = csrfTokenOnForm ? csrfTokenOnServer : null;
    const result = csrfProtectedTransfer({ crossSite: false, providedToken });
    setRequestResult(result.reason);
    appendLog(`[user] POST /transfer same-site -> ${result.reason}`);
  };

  return (
    <div>
      <h1>Security Demos: XSS + CSRF (Real-world style)</h1>
      <p style={styles.helperText}>
        These examples mirror what happens in dashboards, comments, support notes, and payment actions.
        The goal is to build practical engineering intuition, not only definitions.
      </p>

      <div style={styles.card}>
        <h2 style={{ color: '#9f1239' }}>1) XSS: vulnerable rendering vs safe rendering</h2>
        <p>
          Scenario: product allows rich text in "internal notes" and directly injects user input into HTML.
          One malicious note can execute script in every viewer&apos;s browser session.
        </p>

        <div style={styles.payloads}>
          <button onClick={() => usePayload('<img src=x onerror="alert(\'XSS fired\')" />')} style={styles.payloadBtn}>
            Use onerror payload
          </button>
          <button onClick={() => usePayload('<a href="javascript:alert(\'XSS fired\')">Click me</a>')} style={styles.payloadBtn}>
            Use javascript: URL payload
          </button>
          <button onClick={() => usePayload('<b>Normal formatting only</b>')} style={styles.payloadBtn}>
            Use safe-looking markup
          </button>
        </div>

        <div style={styles.twoCol}>
          <div style={styles.col}>
            <h3 style={styles.subHeading}>Vulnerable version</h3>
            <p style={styles.small}>Uses <code>dangerouslySetInnerHTML</code> with unsanitized user input.</p>
            <input
              type="text"
              placeholder="Enter untrusted HTML"
              value={vulnerableInput}
              onChange={(e) => setVulnerableInput(e.target.value)}
              style={styles.input}
            />
            <div style={styles.vulnerableBox}>
              <strong>Rendered output:</strong>
              <div dangerouslySetInnerHTML={{ __html: vulnerableInput }} />
            </div>
          </div>

          <div style={styles.col}>
            <h3 style={styles.subHeading}>Safe version</h3>
            <p style={styles.small}>Renders as text; React escapes HTML by default.</p>
            <input
              type="text"
              placeholder="Enter same input"
              value={safeInput}
              onChange={(e) => setSafeInput(e.target.value)}
              style={styles.input}
            />
            <div style={styles.safeBox}>
              <strong>Rendered output:</strong>
              <div>{safeInput}</div>
            </div>
          </div>
        </div>

        <div style={styles.noteBox}>
          <strong>Engineering takeaway:</strong> if rich HTML is required, sanitize on server + sanitize again
          before rendering + apply CSP. If rich HTML is not required, never use <code>dangerouslySetInnerHTML</code>.
        </div>
      </div>

      <div style={styles.card}>
        <h2 style={{ color: '#1d4ed8' }}>2) CSRF: why cookie auth is vulnerable without token checks</h2>
        <p>
          CSRF works when the browser automatically includes cookies on requests that a victim did not intend.
          Defenses: CSRF token validation and SameSite cookies.
        </p>

        <div style={styles.twoCol}>
          <div style={styles.csrfVuln}>
            <h3 style={styles.subHeading}>Environment toggles</h3>
            <label style={styles.checkboxLine}>
              <input
                type="checkbox"
                checked={cookieSessionEnabled}
                onChange={(e) => setCookieSessionEnabled(e.target.checked)}
              />
              Browser has session cookie
            </label>
            <label style={styles.label}>SameSite mode</label>
            <select
              value={sameSiteMode}
              onChange={(e) => setSameSiteMode(e.target.value)}
              style={styles.select}
            >
              <option value="none">None (least safe)</option>
              <option value="lax">Lax</option>
              <option value="strict">Strict</option>
            </select>

            <label style={styles.checkboxLine}>
              <input
                type="checkbox"
                checked={csrfTokenOnForm}
                onChange={(e) => setCsrfTokenOnForm(e.target.checked)}
              />
              Legit form includes CSRF token
            </label>

            <label style={styles.label}>Server expected CSRF token</label>
            <input
              value={csrfTokenOnServer}
              onChange={(e) => setCsrfTokenOnServer(e.target.value)}
              style={styles.input}
            />
          </div>

          <div style={styles.csrfFixed}>
            <h3 style={styles.subHeading}>Request simulations</h3>
            <button onClick={runAttackSimulation} style={styles.dangerButton}>
              Simulate malicious cross-site request
            </button>
            <button onClick={runLegitTransfer} style={styles.safeButton}>
              Simulate legitimate user transfer
            </button>

            <div style={styles.resultBox}>
              <strong>Latest result:</strong> {requestResult}
            </div>

            <p style={styles.small}>
              Cross-site attack sends no trusted CSRF token. If server requires token, request is blocked.
            </p>
          </div>
        </div>

        <div style={styles.logBox}>
          <strong>Security event log:</strong>
          {securityLog.length === 0 ? (
            <p style={styles.small}>No events yet. Run either simulation.</p>
          ) : (
            <ul style={styles.logList}>
              {securityLog.map((line) => (
                <li key={line}>{line}</li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <div style={styles.card}>
        <h2>3) Practical checklist for frontend engineers</h2>
        <ul style={styles.checkList}>
          <li>Never render untrusted input via <code>dangerouslySetInnerHTML</code> without sanitization.</li>
          <li>Prefer framework-escaped rendering whenever possible.</li>
          <li>Assume cookies are auto-sent; CSRF tokens must be verified server-side.</li>
          <li>Use SameSite cookies as an extra layer, not the only protection.</li>
          <li>Keep security controls testable with deterministic demo URLs and states.</li>
        </ul>
      </div>
    </div>
  );
}

const styles = {
  card: {
    background: '#f8f9fa',
    padding: '1.2rem',
    borderRadius: '8px',
    marginBottom: '1.5rem',
    border: '1px solid #dee2e6',
  },
  helperText: {
    color: '#334155',
    marginBottom: '1rem',
    lineHeight: '1.5',
  },
  twoCol: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
    gap: '1rem',
    marginTop: '0.6rem',
  },
  col: {
    background: '#f8fafc',
    border: '1px solid #e2e8f0',
    borderRadius: '6px',
    padding: '0.8rem',
  },
  subHeading: {
    marginBottom: '0.35rem',
  },
  small: {
    color: '#475569',
    fontSize: '0.9rem',
    marginBottom: '0.5rem',
  },
  payloads: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '0.5rem',
    marginTop: '0.6rem',
    marginBottom: '0.6rem',
  },
  payloadBtn: {
    border: '1px solid #94a3b8',
    borderRadius: '6px',
    padding: '0.35rem 0.7rem',
    cursor: 'pointer',
    background: '#ffffff',
  },
  input: {
    width: '100%',
    padding: '0.5rem',
    margin: '0.5rem 0',
    fontFamily: 'monospace',
  },
  vulnerableBox: {
    background: '#f8d7da',
    padding: '0.8rem',
    border: '1px solid #f5c6cb',
    borderRadius: '4px',
    marginTop: '0.5rem',
  },
  safeBox: {
    background: '#d4edda',
    padding: '0.8rem',
    border: '1px solid #c3e6cb',
    borderRadius: '4px',
    marginTop: '0.5rem',
  },
  noteBox: {
    marginTop: '0.8rem',
    background: '#fff7ed',
    border: '1px solid #fed7aa',
    borderRadius: '6px',
    padding: '0.8rem',
    lineHeight: '1.5',
  },
  csrfVuln: {
    background: '#fce7f3',
    border: '1px solid #fbcfe8',
    padding: '1rem',
    borderRadius: '6px',
  },
  csrfFixed: {
    background: '#e0f2fe',
    border: '1px solid #bae6fd',
    padding: '1rem',
    borderRadius: '6px',
  },
  checkboxLine: {
    display: 'block',
    marginBottom: '0.6rem',
  },
  label: {
    display: 'block',
    marginBottom: '0.25rem',
    fontWeight: 600,
    fontSize: '0.9rem',
  },
  select: {
    width: '100%',
    padding: '0.45rem',
    marginBottom: '0.6rem',
  },
  dangerButton: {
    background: '#e74c3c',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1rem',
    cursor: 'pointer',
    borderRadius: '4px',
  },
  safeButton: {
    background: '#2ecc71',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1rem',
    cursor: 'pointer',
    borderRadius: '4px',
    marginLeft: '0.5rem',
  },
  resultBox: {
    marginTop: '1rem',
    padding: '0.75rem',
    background: '#fff3cd',
    border: '1px solid #ffeeba',
    borderRadius: '4px',
  },
  logBox: {
    marginTop: '0.9rem',
    background: '#f8fafc',
    border: '1px solid #e2e8f0',
    borderRadius: '6px',
    padding: '0.8rem',
  },
  logList: {
    marginTop: '0.45rem',
    paddingLeft: '1rem',
    fontFamily: 'monospace',
    fontSize: '0.86rem',
    lineHeight: '1.45',
  },
  checkList: {
    paddingLeft: '1.1rem',
    lineHeight: '1.6',
  },
};

export default XSSDemo;
