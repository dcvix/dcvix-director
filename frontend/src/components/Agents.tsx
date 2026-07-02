import { useEffect, useState, useContext } from 'react';
import { faSun, faMoon } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { getAgents, approveAgent, denyAgent, revokeAgent, logout } from '../services/api';
import { ThemeContext } from '../contexts/ThemeContext';
import { useNavigate } from 'react-router-dom';
import Logo from './Logo';
import type { Agent } from '../types';

const Agents = () => {
    const { theme, toggleTheme } = useContext(ThemeContext);
    const navigate = useNavigate();
    const [agents, setAgents] = useState<Agent[]>([]);
    const [stateFilter, setStateFilter] = useState('');
    const [error, setError] = useState('');

    const fetchAgents = async () => {
        try {
            const data = await getAgents(stateFilter || undefined);
            setAgents(data || []);
            setError('');
        } catch (err: any) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchAgents();
    }, [stateFilter]);

    const handleApprove = async (guid: string) => {
        try {
            await approveAgent(guid);
            fetchAgents();
        } catch (err: any) {
            setError(err.response?.data || err.message);
        }
    };

    const handleDeny = async (guid: string) => {
        try {
            await denyAgent(guid);
            fetchAgents();
        } catch (err: any) {
            setError(err.response?.data || err.message);
        }
    };

    const handleRevoke = async (guid: string) => {
        try {
            await revokeAgent(guid);
            fetchAgents();
        } catch (err: any) {
            setError(err.response?.data || err.message);
        }
    };

    const handleLogout = async () => {
        await logout();
        window.location.href = '/login';
    };

    return (
        <div>
            <nav className="navbar has-shadow" role="navigation" aria-label="main navigation">
                <div className="navbar-brand">
                    <a className="navbar-item" href="#">
                        <Logo style={{ maxWidth: '200px', width: '100%' }} />
                    </a>
                </div>
                <div className="navbar-start">
                    <a className="navbar-item" onClick={() => navigate('/')}>Dashboard</a>
                    <a className="navbar-item is-active">Agents</a>
                </div>
                <div className="navbar-end">
                    <div className="navbar-item">
                        <button className="button is-ghost" onClick={toggleTheme} title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}>
                            <FontAwesomeIcon icon={theme === 'light' ? faMoon : faSun} />
                        </button>
                    </div>
                    <div className="navbar-item">
                        <button className="button is-danger" onClick={handleLogout}>Logout</button>
                    </div>
                </div>
            </nav>

            <section className="section">
                <div className="container">
                    <div className="level">
                        <div className="level-left">
                            <h1 className="title">Agent Management</h1>
                        </div>
                        <div className="level-right">
                            <div className="field is-grouped">
                                <label className="mr-2 is-size-7" style={{ alignSelf: 'center' }}>State:</label>
                                <span className="select is-small">
                                    <select value={stateFilter} onChange={e => setStateFilter(e.target.value)}>
                                        <option value="">All</option>
                                        <option value="pending">Pending</option>
                                        <option value="registered">Approved</option>
                                        <option value="revoked">Revoked</option>
                                    </select>
                                </span>
                            </div>
                        </div>
                    </div>

                    {error && <div className="notification is-danger is-light">{error}</div>}

                    <div className="table-container">
                        <table className="table is-fullwidth is-striped">
                            <thead>
                                <tr>
                                    <th>GUID</th>
                                    <th>Hostname</th>
                                    <th>State</th>
                                    <th>Created</th>
                                    <th>Registered</th>
                                    <th>Last Seen</th>
                                    <th>Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {agents.map(agent => (
                                    <tr key={agent.guid}>
                                        <td><code>{agent.guid.substring(0, 8)}...</code></td>
                                        <td>{agent.hostname}</td>
                                        <td>
                                            <span className={`tag ${agent.state === 'registered' ? 'is-success' : agent.state === 'pending' ? 'is-warning' : 'is-danger'}`}>
                                                {agent.state}
                                            </span>
                                        </td>
                                        <td>{new Date(agent.created_at).toLocaleString()}</td>
                                        <td>{agent.registered_at ? new Date(agent.registered_at).toLocaleString() : '-'}</td>
                                        <td>{agent.last_seen_at ? new Date(agent.last_seen_at).toLocaleString() : '-'}</td>
                                        <td>
                                            <div className="buttons are-small">
                                                {agent.state === 'pending' && (
                                                    <>
                                                        <button className="button is-success" onClick={() => handleApprove(agent.guid)}>Approve</button>
                                                        <button className="button is-danger" onClick={() => handleDeny(agent.guid)}>Deny</button>
                                                    </>
                                                )}
                                                {agent.state === 'registered' && (
                                                    <button className="button is-warning" onClick={() => handleRevoke(agent.guid)}>Revoke</button>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                                {agents.length === 0 && (
                                    <tr>
                                        <td colSpan={7} className="has-text-centered">No agents found</td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </section>
        </div>
    );
};

export default Agents;
