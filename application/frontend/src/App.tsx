import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import DevicesPage from './pages/DevicesPage';
import GuaranteesPage from './pages/GuaranteesPage';
import TransfersPage from './pages/TransfersPage';
import ConversionsPage from './pages/ConversionsPage';
import CancellationsPage from './pages/CertificatesPage';
import BridgePage from './pages/BridgePage';
import VerificationPage from './pages/VerificationPage';
import OrganizationsPage from './pages/OrganizationsPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const { isAuthenticated } = useAuth();
    return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

export default function App() {
    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
                path="/"
                element={
                    <ProtectedRoute>
                        <Layout />
                    </ProtectedRoute>
                }
            >
                <Route index element={<DashboardPage />} />
                <Route path="devices" element={<DevicesPage />} />
                <Route path="guarantees" element={<GuaranteesPage />} />
                <Route path="transfers" element={<TransfersPage />} />
                <Route path="conversions" element={<ConversionsPage />} />
                <Route path="cancellations" element={<CancellationsPage />} />
                <Route path="bridge" element={<BridgePage />} />
                <Route path="organizations" element={<OrganizationsPage />} />
                <Route path="verification" element={<VerificationPage />} />
                {/* Legacy route redirect */}
                <Route path="certificates" element={<Navigate to="/cancellations" replace />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
}
