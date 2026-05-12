import Navbar from './Navbar';

function Layout({ children }) {
  return (
    <div style={styles.container}>
      <Navbar />
      <main style={styles.main}>{children}</main>
      <footer style={styles.footer}>
        <small>Training app: routing, security & best practices</small>
      </footer>
    </div>
  );
}

const styles = {
  container: {
    fontFamily: 'system-ui, -apple-system, sans-serif',
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
  },
  main: {
    flex: 1,
    padding: '2rem',
    maxWidth: '1200px',
    margin: '0 auto',
    width: '100%',
  },
  footer: {
    textAlign: 'center',
    padding: '1rem',
    backgroundColor: '#ecf0f1',
    fontSize: '0.8rem',
  },
};

export default Layout;