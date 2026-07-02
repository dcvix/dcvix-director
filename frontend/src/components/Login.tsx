import { faEye, faEyeSlash, faSun, faMoon } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useState, useContext } from 'react';
import axios from 'axios';
import { login } from '../services/api';
import { ThemeContext } from '../contexts/ThemeContext';
import Logo from './Logo';

const Login = () => {
    const { theme, toggleTheme } = useContext(ThemeContext);
    const [userID, setUserID] = useState('');
    const [password, setPassword] = useState('');
    const [otp, setOtp] = useState('');
    const [error, setError] = useState('');
    const [showPassword, setShowPassword] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await login(userID, password, otp);
            window.location.href = '/';
        } catch (err: unknown) {
            if (axios.isAxiosError(err)) {
                if (!err.response) {
                    setError('Could not connect to the server');
                } else if (typeof err.response.data === 'string') {
                    setError(err.response.data);
                } else {
                    setError('Invalid credentials');
                }
            } else {
                setError('An unexpected error occurred');
            }
        }
    };

    return (
        <section className="section">
            <div className="container">
                <div className="columns is-centered">
                    <div className="column is-one-third">
                        <div className="box" style={{ position: 'relative' }}>
                            <button className="button is-ghost" onClick={toggleTheme}
                                title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}
                                style={{ position: 'absolute', top: '0.75rem', right: '0.75rem' }}>
                                <FontAwesomeIcon icon={theme === 'light' ? faMoon : faSun} />
                            </button>
                            <div style={{ textAlign: 'center', marginBottom: '1rem' }}>
                                <Logo style={{ maxWidth: '200px', width: '100%' }} />
                            </div>
                            {error && (
                                <div className="notification is-danger">
                                    {error}
                                </div>
                            )}
                            <form onSubmit={handleSubmit}>
                                <div className="field">
                                    <label className="label">Username</label>
                                    <div className="control">
                                        <input
                                            className="input"
                                            type="text"
                                            value={userID}
                                            onChange={(e) => setUserID(e.target.value)}
                                            required
                                        />
                                    </div>
                                </div>
                                <div className="field">
                                    <label className="label">Password</label>
                                    <div className="control has-icons-right">
                                        <input
                                            className="input"
                                            type={showPassword ? "text" : "password"}
                                            value={password}
                                            onChange={(e) => setPassword(e.target.value)}
                                            required
                                        />
                                        <span
                                            className="icon is-small is-right is-clickable"
                                            onClick={() => setShowPassword(!showPassword)}
                                            style={{ pointerEvents: 'auto' }}
                                        >
                                            <FontAwesomeIcon icon={showPassword ? faEyeSlash : faEye} />
                                        </span>
                                    </div>
                                </div>
                                <div className="field">
                                    <label className="label">OTP</label>
                                    <div className="control">
                                        <input
                                            className="input"
                                            type="text"
                                            value={otp}
                                            onChange={(e) => setOtp(e.target.value)}
                                        />
                                    </div>
                                </div>
                                <button className="button is-primary is-fullwidth" type="submit">
                                    Login
                                </button>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
};

export default Login;