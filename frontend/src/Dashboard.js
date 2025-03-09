import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { useAuth } from './AuthContext';
import './styles.css';
import { format } from 'date-fns';

const Dashboard = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const [showAddAccount, setShowAddAccount] = useState(false);
    const [banks, setBanks] = useState([]);
    const [bankAccounts, setBankAccounts] = useState([]);
    const [bankDetails, setBankDetails] = useState({});
    const [formData, setFormData] = useState({
        bank_id: '',
        agency_number: '',
        account_number: '',
        balance: 0.0,
    });
    const [displayBalance, setDisplayBalance] = useState('R$ 0,00');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [isLoading, setIsLoading] = useState(true);
    // Add near other state declarations
    const [showTransactionForm, setShowTransactionForm] = useState(false);
    const [transactionCategories, setTransactionCategories] = useState([]);
    const [transactions, setTransactions] = useState([]);
    const [isLoadingTransactions, setIsLoadingTransactions] = useState(true);

    // Load banks and user's bank accounts
    useEffect(() => {
        const fetchBanks = axios.get('http://localhost:8080/api/banks');
        const fetchAccounts = axios.get('http://localhost:8080/api/bank-accounts', {
            headers: {
                Authorization: localStorage.getItem('token'),
            },
        });

        // Execute both requests concurrently
        Promise.all([fetchBanks, fetchAccounts])
            .then(([banksResponse, accountsResponse]) => {
                const banksData = banksResponse.data;
                setBanks(banksData);

                // Create a lookup object for bank details
                const bankMap = {};
                banksData.forEach(bank => {
                    bankMap[bank.id.toLowerCase()] = bank; // Convert to lowercase to match BankID
                });
                setBankDetails(bankMap);

                // Set the bank accounts
                setBankAccounts(accountsResponse.data);
                setIsLoading(false);
            })
            .catch(err => {
                setError('Falha ao carregar dados. Tente novamente.');
                setIsLoading(false);
                console.error('Erro ao carregar dados:', err);
            });
    }, []);

    // Refresh accounts after adding a new one
    const refreshAccounts = () => {
        setIsLoading(true);
        axios.get('http://localhost:8080/api/bank-accounts', {
            headers: {
                Authorization: localStorage.getItem('token'),
            },
        })
            .then(response => {
                setBankAccounts(response.data);
                setIsLoading(false);
            })
            .catch(err => {
                setError('Falha ao atualizar contas:');
                setIsLoading(false);
            });
    };

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

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
        setError('');
        try {
            await axios.post('http://localhost:8080/api/bank-accounts', formData, {
                headers: {
                    Authorization: localStorage.getItem('token'),
                },
            });
            setSuccess('Conta bancária adicionada com sucesso!');

            // Reset form
            setFormData({
                bank_id: '',
                agency_number: '',
                account_number: '',
                balance: 0.0,
            });
            setDisplayBalance('R$ 0,00');

            // Refresh the accounts list
            refreshAccounts();

            // Hide form after successful submission
            setTimeout(() => {
                setShowAddAccount(false);
                setSuccess('');
            }, 3000);

        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao criar conta');
        }
    };
    // Add with other state declarations
    const [transactionData, setTransactionData] = useState({
        category_id: '',
        from_account_id: '',
        to_account_id: '',
        amount: 0,
        description: ''
    });

    const [transactionDisplayAmount, setTransactionDisplayAmount] = useState('R$ 0,00');

// Add this useEffect to load transaction categories
    useEffect(() => {
        axios.get('http://localhost:8080/api/transaction-categories')
            .then(response => setTransactionCategories(response.data))
            .catch(() => setError('Falha ao carregar categorias'));
    }, []);

// Add this handler
    const handleTransactionAmountChange = (e) => {
        const rawValue = e.target.value.replace(/[R$\s.,]/g, '');
        const numericValue = parseFloat(rawValue) / 100;

        if (!isNaN(numericValue)) {
            setTransactionData({ ...transactionData, amount: numericValue });
            setTransactionDisplayAmount(new Intl.NumberFormat('pt-BR', {
                style: 'currency',
                currency: 'BRL'
            }).format(numericValue));
        } else {
            setTransactionData({ ...transactionData, amount: 0 });
            setTransactionDisplayAmount('R$ 0,00');
        }
    };

// Add this submit handler
    const handleTransactionSubmit = async (e) => {
        e.preventDefault();
        try {
            await axios.post('http://localhost:8080/api/transactions', transactionData, {
                headers: { Authorization: localStorage.getItem('token') }
            });
            refreshAccounts(); // Update balances
            refreshTransactions(); // Update balances
            setShowTransactionForm(false);
            setTransactionData({
                category_id: '',
                from_account_id: '',
                to_account_id: '',
                amount: 0,
                description: ''
            });
            setTransactionDisplayAmount('R$ 0,00');
            setSuccess('Transação registrada com sucesso!');
        } catch (err) {
            setError(err.response?.data?.error || 'Falha na transação');
        }
    };

    useEffect(() => {
        axios.get('http://localhost:8080/api/transactions', {
            headers: {
                Authorization: localStorage.getItem('token'),
            }
        })
            .then(response => {
                setTransactions(response.data);
                setIsLoadingTransactions(false);
            })
            .catch(err => {
                setError('Falha ao carregar transações');
                setIsLoadingTransactions(false);
                console.error('Erro ao carregar transações:', err);
            });
    }, []);

// Add this function to refresh transactions after creating a new one
    const refreshTransactions = () => {
        setIsLoadingTransactions(true);
        axios.get('http://localhost:8080/api/transactions', {
            headers: {
                Authorization: localStorage.getItem('token'),
            }
        })
            .then(response => {
                setTransactions(response.data);
                setIsLoadingTransactions(false);
            })
            .catch(err => {
                setIsLoadingTransactions(false);
            });
    };

    // Calculate total balance across all accounts
    const totalBalance = bankAccounts.reduce((sum, account) => sum + account.Balance, 0);

    return (
        <div className="dashboard">
            <div className="dashboard-header">
                <h1>Bem vindo, {user?.username}</h1>
                <button className="logout-button" onClick={handleLogout}>
                    Logout
                </button>
            </div>

            <div className="game-stats">
                <div className="stat-card">
                    <h3>💰 Saldo Total</h3>
                    <p>{formatCurrency(totalBalance)}</p>
                </div>
                <div className="stat-card">
                    <h3>🏦 Contas Bancárias</h3>
                    <p>{bankAccounts.length}</p>
                </div>
            </div>

            {/* Bank Account Section */}
            <div className="dashboard-section">
                <div className="section-header">
                    <h2>Contas Bancárias</h2>
                    <button
                        className="action-button"
                        onClick={() => setShowAddAccount(!showAddAccount)}
                    >
                        {showAddAccount ? 'Cancelar' : 'Adicionar Conta'}
                    </button>
                </div>

                {/* Add Bank Account Form */}
                {showAddAccount && (
                    <div className="form-container">
                        {success && <div className="success-message">{success}</div>}
                        {error && <div className="error-message">{error}</div>}

                        <form className="dashboard-form" onSubmit={handleSubmit}>
                            <select
                                value={formData.bank_id}
                                onChange={(e) => setFormData({ ...formData, bank_id: e.target.value })}
                                required
                            >
                                <option value="">Selecione o banco</option>
                                {banks.map((bank) => (
                                    <option key={bank.id} value={bank.id}>
                                        {bank.name}
                                    </option>
                                ))}
                            </select>

                            <div className="form-row">
                                <input
                                    type="text"
                                    placeholder="Agência"
                                    value={formData.agency_number}
                                    onChange={(e) => setFormData({ ...formData, agency_number: e.target.value })}
                                    required
                                />

                                <input
                                    type="text"
                                    placeholder="Número da conta"
                                    value={formData.account_number}
                                    onChange={(e) => setFormData({ ...formData, account_number: e.target.value })}
                                    required
                                />
                            </div>

                            <input
                                type="text"
                                placeholder="Saldo Inicial"
                                value={displayBalance}
                                onChange={handleBalanceChange}
                                onFocus={(e) => e.target.select()}
                                required
                            />

                            <button type="submit" className="submit-button">Adicionar</button>
                        </form>
                    </div>
                )}

                {/* List of Bank Accounts */}
                <div className="accounts-list">
                    {isLoading ? (
                        <p className="loading-text">Carregando suas contas...</p>
                    ) : bankAccounts.length === 0 ? (
                        <p className="placeholder-text">
                            {!showAddAccount && "Nenhuma conta bancária, clique em Adicionar Conta para começar."}
                        </p>
                    ) : (
                        <div className="accounts-grid">
                            {bankAccounts.map(account => (
                                <div key={account.ID} className="account-card">
                                    <div className="account-header">
                                        <h3>{bankDetails[account.BankID.toLowerCase()]?.name || 'Bank'}</h3>
                                        <span className="account-date">
                                            {format(new Date(account.CreatedAt), 'dd/MM/yyyy HH:mm')}
                                        </span>
                                    </div>
                                    <div className="account-details">
                                        <p><strong>Agência:</strong> {account.AgencyNumber}</p>
                                        <p><strong>Conta:</strong> {account.AccountNumber}</p>
                                        <p className={`account-balance ${account.Balance < 0 ? 'negative' : ''}`}>
                                            {formatCurrency(account.Balance)}
                                        </p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
            {/* Transaction Section */}
            <div className="dashboard-section">
                <div className="section-header">
                    <h2>Transações Financeiras</h2>
                    <button
                        className="action-button"
                        onClick={() => setShowTransactionForm(!showTransactionForm)}
                    >
                        {showTransactionForm ? 'Cancelar' : 'Nova Transação'}
                    </button>
                </div>

                {/* Transaction Form */}
                {showTransactionForm && (
                    <div className="form-container">
                        <form className="dashboard-form" onSubmit={handleTransactionSubmit}>
                            <select
                                value={transactionData.category_id}
                                onChange={(e) => setTransactionData({ ...transactionData, category_id: e.target.value })}
                                required
                            >
                                <option value="">Selecione o Tipo</option>
                                {transactionCategories.map(cat => (
                                    <option key={cat.ID} value={cat.ID}>{cat.Name}</option>
                                ))}
                            </select>

                            {/* From Account (for transfers/debits) */}
                            {transactionData.category_id &&
                                transactionCategories.find(c => c.ID === transactionData.category_id)?.Type !== 'credit' && (
                                    <select
                                        value={transactionData.from_account_id}
                                        onChange={(e) => setTransactionData({ ...transactionData, from_account_id: e.target.value })}
                                        required
                                    >
                                        <option value="">Conta de Origem</option>
                                        {bankAccounts.map(acc => (
                                            <option key={acc.ID} value={acc.ID}>
                                                {bankDetails[acc.BankID]?.name} - {acc.AccountNumber}
                                            </option>
                                        ))}
                                    </select>
                                )}

                            {/* To Account (for transfers/credits) */}
                            {transactionData.category_id &&
                                transactionCategories.find(c => c.ID === transactionData.category_id)?.Type !== 'debit' && (
                                    <select
                                        value={transactionData.to_account_id}
                                        onChange={(e) => setTransactionData({ ...transactionData, to_account_id: e.target.value })}
                                        required
                                    >
                                        <option value="">Conta de Destino</option>
                                        {bankAccounts.map(acc => (
                                            <option key={acc.ID} value={acc.ID}>
                                                {bankDetails[acc.BankID]?.name} - {acc.AccountNumber}
                                            </option>
                                        ))}
                                    </select>
                                )}

                            <input
                                type="text"
                                placeholder="Valor (R$)"
                                value={transactionDisplayAmount}
                                onChange={handleTransactionAmountChange}
                                required
                            />

                            <input
                                type="text"
                                placeholder="Descrição"
                                value={transactionData.description}
                                onChange={(e) => setTransactionData({ ...transactionData, description: e.target.value })}
                            />

                            <button type="submit" className="submit-button">
                                Registrar Transação
                            </button>
                        </form>
                    </div>
                )}
            </div>
            <div className="dashboard-section">
                <div className="section-header">
                    <h2>Histórico de Transações</h2>
                </div>

                <div className="transactions-list">
                    {isLoadingTransactions ? (
                        <p className="loading-text">Carregando transações...</p>
                    ) : transactions.length === 0 ? (
                        <p className="placeholder-text">Nenhuma transação encontrada.</p>
                    ) : (
                        <div className="transactions-table-container">
                            <table className="transactions-table">
                                <thead>
                                <tr>
                                    <th>Data</th>
                                    <th>Categoria</th>
                                    <th>De</th>
                                    <th>Para</th>
                                    <th>Valor</th>
                                    <th>Descrição</th>
                                </tr>
                                </thead>
                                <tbody>
                                {transactions.map(transaction => {
                                    const fromAccount = bankAccounts.find(acc => acc.ID === transaction.FromAccountID);
                                    const toAccount = bankAccounts.find(acc => acc.ID === transaction.ToAccountID);

                                    return (
                                        <tr key={transaction.ID}>
                                            <td>{new Date(transaction.transaction_date).toLocaleDateString()}</td>
                                            <td>{transaction.Category ? transaction.Category.Name : '-'}</td>
                                            <td>
                                                {fromAccount ?
                                                    `${bankDetails[fromAccount.BankID.toLowerCase()]?.name || 'Banco'} (${fromAccount.AccountNumber})` :
                                                    '-'}
                                            </td>
                                            <td>
                                                {toAccount ?
                                                    `${bankDetails[toAccount.BankID.toLowerCase()]?.name || 'Banco'} (${toAccount.AccountNumber})` :
                                                    '-'}
                                            </td>
                                            <td className={transaction.Category?.Type === 'debit' ? 'amount-negative' : 'amount-positive'}>
                                                {formatCurrency(transaction.Amount)}
                                            </td>
                                            <td>{transaction.Description || '-'}</td>
                                        </tr>
                                    );
                                })}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Dashboard;