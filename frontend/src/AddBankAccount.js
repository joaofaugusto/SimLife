import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useAuth } from './AuthContext';

const AddBankAccount = () => {
    const { user } = useAuth();
    const [banks, setBanks] = useState([]);
    const [formData, setFormData] = useState({
        bank_id: '',
        agency_number: '',
        account_number: '',
        balance: 0.0,
    });
    const [displayBalance, setDisplayBalance] = useState('R$ 0,00');
    const [error, setError] = useState('');

    useEffect(() => {
        axios.get('http://localhost:8080/api/banks')
            .then((response) => setBanks(response.data))
            .catch(() => setError('Failed to load banks'));
    }, []);

    // Format the value as Brazilian currency
    const formatCurrency = (value) => {
        return `R$ ${value.toFixed(2).replace('.', ',')}`;
    };

    // Handle balance input changes
    const handleBalanceChange = (e) => {
        // Get the raw input value (remove R$, spaces, and commas)
        const rawValue = e.target.value.replace(/[R$\s.,]/g, '');

        // Convert to a number (divide by 100 to handle decimals correctly)
        const numericValue = parseFloat(rawValue) / 100;

        if (!isNaN(numericValue)) {
            // Update the actual value in formData
            setFormData({ ...formData, balance: numericValue });
            // Format for display
            setDisplayBalance(formatCurrency(numericValue));
        } else {
            // Handle empty or invalid input
            setFormData({ ...formData, balance: 0 });
            setDisplayBalance('R$ 0,00');
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await axios.post('http://localhost:8080/api/bank-accounts', formData, {
                headers: {
                    Authorization: localStorage.getItem('token'),
                },
            });
            alert('Bank account added!');
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to create account');
        }
    };

    return (
        <div className="auth-container">
            <form className="auth-form" onSubmit={handleSubmit}>
                <h1>Add Bank Account</h1>
                {error && <p className="error-message">{error}</p>}
                <select
                    value={formData.bank_id}
                    onChange={(e) => setFormData({ ...formData, bank_id: e.target.value })}
                    required
                >
                    <option value="">Select a Bank</option>
                    {banks.map((bank) => (
                        <option key={bank.id} value={bank.id}>
                            {bank.name}
                        </option>
                    ))}
                </select>
                <input
                    type="text"
                    placeholder="Agency Number"
                    value={formData.agency_number}
                    onChange={(e) => setFormData({ ...formData, agency_number: e.target.value })}
                    required
                />
                <input
                    type="text"
                    placeholder="Account Number"
                    value={formData.account_number}
                    onChange={(e) => setFormData({ ...formData, account_number: e.target.value })}
                    required
                />
                <input
                    type="text"
                    placeholder="Initial Balance"
                    value={displayBalance}
                    onChange={handleBalanceChange}
                    onFocus={(e) => {
                        // Optional: Select all text on focus for easier editing
                        e.target.select();
                    }}
                    required
                />
                <button type="submit">Add Account</button>
            </form>
        </div>
    );
};

export default AddBankAccount;