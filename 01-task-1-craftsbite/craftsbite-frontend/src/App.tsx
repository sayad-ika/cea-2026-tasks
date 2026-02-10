import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './contexts/AuthContext';
import { ProtectedRoute } from './components';
import { Login } from './pages/Login';
import { Home } from './pages/Home';
import { HeadcountDashboard } from './pages/HeadcountDashboard';
import { OverridePanel } from './pages/OverridePanel';
import ComponentShowcase from './pages/ComponentShowcase';
import './App.css';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 3000,
            style: {
              borderRadius: '16px',
              background: 'var(--color-background-light)',
              color: 'var(--color-text-main)',
              boxShadow: '10px 10px 20px #e6dccf, -10px -10px 20px #ffffff',
            },
          }}
        />
        <Routes>
          {/* Public */}
          <Route path="/login" element={<Login />} />

          {/* Authenticated — any role */}
          <Route path="/home" element={
            <ProtectedRoute><Home /></ProtectedRoute>
          } />

          {/* Admin / Logistics only */}
          <Route path="/headcount" element={
            <ProtectedRoute allowedRoles={['admin', 'logistics']}>
              <HeadcountDashboard />
            </ProtectedRoute>
          } />

          {/* Admin / Team Lead only */}
          <Route path="/override" element={
            <ProtectedRoute allowedRoles={['admin', 'team_lead']}>
              <OverridePanel />
            </ProtectedRoute>
          } />

          {/* Showcase (dev only) */}
          <Route path="/showcase" element={<ComponentShowcase />} />

          {/* Catch-all */}
          <Route path="*" element={<Navigate to="/home" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
