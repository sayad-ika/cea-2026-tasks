import { NavLink } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const navLinkStyle = ({ isActive }) => ({
  color: 'white',
  textDecoration: 'none',
  fontWeight: '500',
  padding: '0.25rem 0.5rem',
  borderRadius: '4px',
  background: isActive ? '#1abc9c' : 'transparent',
  outline: isActive ? '2px solid #1abc9c' : 'none',
});

function Navbar() {
  const { isLoggedIn } = useAuth();
  return (
    <nav style={styles.nav}>
      <NavLink to="/" end style={navLinkStyle}>Intro</NavLink>
      <NavLink to="/routing" style={navLinkStyle}>Routing</NavLink>
      <NavLink to="/params/42" style={navLinkStyle}>Params/Query</NavLink>
      <NavLink to="/protected" style={navLinkStyle}>Protected</NavLink>
      <NavLink to="/demo" style={navLinkStyle}>Demo App</NavLink>
      <NavLink to="/xss" style={navLinkStyle}>XSS+CSRF</NavLink>
      <NavLink to="/storage" style={navLinkStyle}>Storage</NavLink>
      <NavLink to="/admin" style={navLinkStyle}>Admin (protected)</NavLink>
      <span style={styles.authBadge}>
        {isLoggedIn ? '✅ Logged in' : '🔒 Not logged in'}
      </span>
    </nav>
  );
}

const styles = {
  nav: {
    background: '#2c3e50',
    padding: '0.75rem 1rem',
    display: 'flex',
    gap: '1rem',
    flexWrap: 'wrap',
    alignItems: 'center',
    borderBottom: '2px solid #ecf0f1',
  },
  authBadge: {
    marginLeft: 'auto',
    background: '#1abc9c',
    padding: '0.25rem 0.75rem',
    borderRadius: '20px',
    fontSize: '0.8rem',
    color: '#fff',
  },
};

export default Navbar;