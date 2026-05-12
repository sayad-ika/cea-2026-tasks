import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

function ProtectedRoute({ children }) {
  const { isLoggedIn } = useAuth();

  if (!isLoggedIn) {
    // Redirect to the demo page where the user can log in
    return <Navigate to="/demo" replace />;
  }

  return children;
}

export default ProtectedRoute;