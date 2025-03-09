import React, { useState } from 'react';
import axios from 'axios';
import './styles.css'; // Import shared styles

const Register = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (password.length < 5) {
            setError('Password must be at least 5 characters');
            return;
        }

        try {
            const response = await axios.post('http://localhost:8080/api/register', {
                username,
                password_hash: password,
            });
            alert(response.data.message);
        } catch (err) {
            setError(err.response?.data?.error || 'Registration failed');
        }
    };

    return (
        <div className="auth-container">
            <form className="auth-form" onSubmit={handleSubmit}>
                <h1>Register</h1>
                {error && <p className="error-message">{error}</p>}
                <input
                    type="text"
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                />
                <input
                    type="password"
                    placeholder="Password (min 5 characters)"
                    value={password}
                    onChange={(e) => {
                        setPassword(e.target.value);
                        setError('');
                    }}
                    required
                />
                <button type="submit">Register</button>
            </form>
        </div>
    );
};

export default Register;