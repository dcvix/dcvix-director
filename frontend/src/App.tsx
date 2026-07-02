import { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import AgentsPage from './pages/AgentsPage';
import { checkAuth } from './services/api';

function App() {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const verifyAuth = async () => {
            const auth = await checkAuth();
            setIsAuthenticated(auth);
            setLoading(false);
        };
        verifyAuth();
    }, []);

    if (loading) {
        return null;
    }

    return (
        <Router>
            <Routes>
                <Route path="/login" element={!isAuthenticated ? <LoginPage /> : <Navigate to="/" />} />
                <Route 
                    path="/agents" 
                    element={isAuthenticated ? <AgentsPage /> : <Navigate to="/login" />}
                />
                <Route 
                    path="/*" 
                    element={isAuthenticated ? <DashboardPage /> : <Navigate to="/login" />}
                />
            </Routes>
        </Router>
    );
}

export default App;