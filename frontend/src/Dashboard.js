import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';
import './styles.css';

const Dashboard = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="dashboard">
            <div className="dashboard-header">
                <h1>Welcome, {user?.username}</h1>
                <button className="logout-button" onClick={handleLogout}>
                    Logout
                </button>
            </div>

            <div className="game-stats">
                <div className="stat-card">
                    <h3>💰 Balance</h3>
                    <p>$10,000</p>
                </div>
                <div className="stat-card">
                    <h3>📈 Total Investments</h3>
                    <p>$2,500</p>
                </div>
                <div className="stat-card">
                    <h3>📅 Days Active</h3>
                    <p>3</p>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;