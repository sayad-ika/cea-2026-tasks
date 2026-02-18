import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "react-hot-toast";
import { AuthProvider } from "./contexts/AuthContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import { ProtectedRoute, ErrorBoundary } from "./components";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Home } from "./pages/Home";
import { HeadcountDashboard } from "./pages/HeadcountDashboard";
import { OverridePanel } from "./pages/OverridePanel";
import "./App.css";
import { TeamParticipation } from "./pages/TeamParticipation";

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <ErrorBoundary>
          <BrowserRouter>
            <Toaster
              position="top-right"
              toastOptions={{
                duration: 3000,
                style: {
                  borderRadius: "16px",
                  background: "var(--color-background-light)",
                  color: "var(--color-text-main)",
                  boxShadow: "10px 10px 20px #e6dccf, -10px -10px 20px #ffffff",
                },
              }}
            />
            <Routes>
              {/* Public */}
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />

              {/* Authenticated — any role */}
              <Route
                path="/home"
                element={
                  <ProtectedRoute>
                    <Home />
                  </ProtectedRoute>
                }
              />

              {/* Admin / Logistics only */}
              <Route
                path="/headcount"
                element={
                  <ProtectedRoute allowedRoles={["admin", "logistics"]}>
                    <HeadcountDashboard />
                  </ProtectedRoute>
                }
              />

              {/* Admin / Team Lead only */}
              <Route
                path="/override"
                element={
                  <ProtectedRoute allowedRoles={["admin", "team_lead"]}>
                    <OverridePanel />
                  </ProtectedRoute>
                }
              />

              <Route
                path="/team"
                element={
                  <ProtectedRoute allowedRoles={["admin", "team_lead", "logistics"]}>
                    <TeamParticipation />
                  </ProtectedRoute>
                }
              />

              {/* <Route path="/showcase" element={<ComponentShowcase />} /> */}

              {/* Catch-all */}
              <Route path="*" element={<Navigate to="/home" replace />} />
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
