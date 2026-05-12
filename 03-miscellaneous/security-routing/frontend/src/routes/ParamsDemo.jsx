import { useMemo } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';

function ParamsDemo() {
  const { userId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = searchParams.get('tab') || 'profile';
  const sort = searchParams.get('sort') || 'recent';
  const page = Number(searchParams.get('page') || 1);
  const view = searchParams.get('view') || 'table';
  const q = searchParams.get('q') || '';

  const fullUrl = useMemo(() => {
    const query = searchParams.toString();
    return `/params/${userId}${query ? `?${query}` : ''}`;
  }, [searchParams, userId]);

  const updateQuery = (updates) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value === '' || value === null || value === undefined) {
        next.delete(key);
      } else {
        next.set(key, String(value));
      }
    });
    setSearchParams(next);
  };

  const changeTab = (newTab) => {
    // Reset page when moving to a new tab so pagination does not look broken.
    updateQuery({ tab: newTab, page: 1 });
  };

  const nextPage = () => {
    updateQuery({ page: page + 1 });
  };

  const prevPage = () => {
    if (page > 1) {
      updateQuery({ page: page - 1 });
    }
  };

  const clearOptionalFilters = () => {
    updateQuery({ q: '', sort: null });
  };

  return (
    <div>
      <h1>🔢 Route Params & Query Strings</h1>
      <div style={styles.card}>
        <p style={styles.helperText}>
          In production apps, route params identify the resource (which user/project),
          and query params capture UI state (tab, sort, pagination, search). This makes
          the view shareable and reproducible.
        </p>

        <div style={styles.paramBox}>
          <strong>Route param (:userId):</strong> <code>{userId}</code>
          <br />
          <small>Extracted using <code>useParams()</code> and usually maps to API/resource identity.</small>
        </div>

        <div style={styles.queryBox}>
          <strong>Current URL:</strong> <code>{fullUrl}</code>
          <br />
          <strong>Resolved state:</strong>
          <ul style={styles.stateList}>
            <li><code>tab</code>: {tab}</li>
            <li><code>sort</code>: {sort}</li>
            <li><code>page</code>: {page}</li>
            <li><code>view</code>: {view}</li>
            <li><code>q</code>: {q || '(empty)'}</li>
          </ul>
        </div>

        <div style={styles.controlsBox}>
          <h4>Tab (query param)</h4>
          <div style={styles.btnRow}>
            <button onClick={() => changeTab('profile')} style={styles.btn}>
              tab=profile
            </button>
            <button onClick={() => changeTab('settings')} style={styles.btn}>
              tab=settings
            </button>
            <button onClick={() => changeTab('billing')} style={styles.btn}>
              tab=billing
            </button>
          </div>

          <h4 style={styles.subTitle}>Sort + View</h4>
          <div style={styles.btnRow}>
            <button onClick={() => updateQuery({ sort: 'recent' })} style={styles.btn}>sort=recent</button>
            <button onClick={() => updateQuery({ sort: 'oldest' })} style={styles.btn}>sort=oldest</button>
            <button onClick={() => updateQuery({ view: 'table' })} style={styles.btn}>view=table</button>
            <button onClick={() => updateQuery({ view: 'card' })} style={styles.btn}>view=card</button>
          </div>

          <h4 style={styles.subTitle}>Pagination</h4>
          <div style={styles.btnRow}>
            <button onClick={prevPage} style={styles.btn} disabled={page <= 1}>Previous page</button>
            <button onClick={nextPage} style={styles.btn}>Next page</button>
          </div>

          <h4 style={styles.subTitle}>Search Filter</h4>
          <div style={styles.btnRow}>
            <button onClick={() => updateQuery({ q: 'invoice' })} style={styles.btn}>q=invoice</button>
            <button onClick={() => updateQuery({ q: 'failed-payment' })} style={styles.btn}>q=failed-payment</button>
            <button onClick={() => updateQuery({ q: '' })} style={styles.btn}>clear q</button>
          </div>

          <button onClick={clearOptionalFilters} style={styles.secondaryBtn}>
            Clear optional filters (q, sort)
          </button>
        </div>

        <div style={styles.exampleBox}>
          <h4>Real-world URL examples</h4>
          <ul>
            <li><code>/params/123?tab=profile&sort=recent&page=1</code></li>
            <li><code>/params/acme-finance?tab=billing&view=card&q=invoice</code></li>
            <li><code>/params/8472?tab=settings&page=3&sort=oldest</code></li>
          </ul>
          <p style={styles.noteText}>
            Tip: route param usually maps to resource identity. Query params should store UI controls that
            users may share with teammates.
          </p>
          <p style={styles.quickLinks}>
            Quick links:{' '}
            <Link to="/params/123?tab=profile&sort=recent&page=1">User 123</Link>
            {' | '}
            <Link to="/params/8472?tab=billing&view=card&q=invoice">Billing case</Link>
          </p>
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
    border: '1px solid #dee2e6',
  },
  helperText: {
    color: '#334155',
    lineHeight: '1.5',
    marginBottom: '1rem',
  },
  paramBox: {
    background: '#e2e3e5',
    padding: '1rem',
    borderRadius: '4px',
    marginBottom: '1rem',
  },
  queryBox: {
    background: '#d1ecf1',
    padding: '1rem',
    borderRadius: '4px',
    marginBottom: '1rem',
  },
  controlsBox: {
    background: '#eef2ff',
    padding: '1rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    border: '1px solid #c7d2fe',
  },
  stateList: {
    marginTop: '0.5rem',
    paddingLeft: '1.1rem',
    lineHeight: '1.5',
  },
  btnRow: {
    display: 'flex',
    gap: '0.5rem',
    flexWrap: 'wrap',
    marginBottom: '0.75rem',
  },
  subTitle: {
    marginTop: '0.35rem',
    marginBottom: '0.35rem',
  },
  btn: {
    padding: '0.35rem 0.8rem',
    cursor: 'pointer',
    border: '1px solid #94a3b8',
    borderRadius: '5px',
    background: '#ffffff',
  },
  secondaryBtn: {
    marginTop: '0.25rem',
    padding: '0.4rem 0.85rem',
    cursor: 'pointer',
    border: 'none',
    borderRadius: '5px',
    background: '#334155',
    color: '#ffffff',
  },
  exampleBox: {
    background: '#fff3cd',
    padding: '1rem',
    borderRadius: '4px',
  },
  noteText: {
    marginTop: '0.6rem',
    color: '#374151',
  },
  quickLinks: {
    marginTop: '0.5rem',
  },
};

export default ParamsDemo;