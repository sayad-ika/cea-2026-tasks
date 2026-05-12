import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';

import Intro from './routes/Intro';
import RoutingBasics from './routes/RoutingBasics';
import ParamsDemo from './routes/ParamsDemo';
import ProtectedRoutesExplanation from './routes/ProtectedRoutesExplanation';
import DemoApp from './routes/DemoApp';
import XSSDemo from './routes/XSSDemo';
import StorageDemo from './routes/StorageDemo';
import AdminPage from './routes/AdminPage';

function NotFound() {
  return (
    <div style={{ textAlign: 'center', marginTop: '4rem' }}>
      <h1>404 — Page Not Found</h1>
      <p>This route doesn't exist. Use the navbar to navigate.</p>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Intro />} />
          <Route path="/routing" element={<RoutingBasics />} />
          <Route path="/params/:userId" element={<ParamsDemo />} />
          <Route path="/protected" element={<ProtectedRoutesExplanation />} />
          <Route path="/demo" element={<DemoApp />} />
          <Route path="/xss" element={<XSSDemo />} />
          <Route path="/storage" element={<StorageDemo />} />
          {/* Protected route example: /admin requires login */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute>
                <AdminPage />
              </ProtectedRoute>
            }
          />
          {/* Catch-all: any unknown URL shows 404 */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;