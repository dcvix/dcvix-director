import { useEffect, useState, useRef, useContext } from 'react';
import { faSun, faMoon } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useNavigate } from 'react-router-dom';
import { closeSession, getServers, getSessions, logout } from '../services/api';
import { ThemeContext } from '../contexts/ThemeContext';
import Logo from './Logo';
import type { Server, Session } from '../types';


const Dashboard = () => {
    const { theme, toggleTheme } = useContext(ThemeContext);
    const navigate = useNavigate();
    const [sessions, setSessions] = useState<Session[]>([]);
    const [servers, setServers] = useState<Server[]>([]);
    const [activeTab, setActiveTab] = useState('sessions');
    const [serverIdFilter, setServerIdFilter] = useState('');
    const [ownerFilter, setOwnerFilter] = useState('');
    const [connectionsFilter, setConnectionsFilter] = useState('');
    const [isModalActive, setIsModalActive] = useState(false);
    const [closeSessionResult, setCloseSessionResult] = useState('');
    const [isConfirmModalActive, setIsConfirmModalActive] = useState(false);
    const [sessionToClose, setSessionToClose] = useState<{sessionId: string, serverId: string, owner: string} | null>(null);
    const [sessionSortConfig, setSessionSortConfig] = useState<{key: keyof Session, direction: string}>({ key: 'owner', direction: 'asc' });
    const [serverSortConfig, setServerSortConfig] = useState<{key: keyof Server, direction: string}>({ key: 'hostname', direction: 'asc' });
    const [refreshInterval, setRefreshInterval] = useState(5000);
    const intervalRef = useRef<number | null>(null);

    const fetchData = async () => {
        try {
            const sessionsData = await getSessions();
            const serversData = await getServers();
            setSessions(sessionsData || []);
            setServers(serversData || []);
        } catch (error) {
            console.error('Error fetching data:', error);
        }
    };

    useEffect(() => {
        fetchData();
        if (intervalRef.current) {
            clearInterval(intervalRef.current);
        }
        if (refreshInterval > 0) {
            intervalRef.current = window.setInterval(fetchData, refreshInterval);
        }
        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        };
    }, [refreshInterval]);

    const handleLogout = async () => {
        await logout();
        window.location.href = '/login';
    };

    const handleCloseSession = (sessionId: string, serverId: string, owner: string) => {
        setSessionToClose({ sessionId, serverId, owner });
        setIsConfirmModalActive(true);
    };

    const confirmCloseSession = async () => {
        if (!sessionToClose) return;

        try {
            const result = await closeSession(sessionToClose.sessionId, sessionToClose.serverId);
            setCloseSessionResult(JSON.stringify(result, null, 2));
            setIsModalActive(true);
            fetchData();
        } catch (err: any) {
            setCloseSessionResult('Error: ' + err.message);
            setIsModalActive(true);
            console.error('Error closing session:', err);
        } finally {
            setIsConfirmModalActive(false);
            setSessionToClose(null);
        }
    };

    const cancelCloseSession = () => {
        setIsConfirmModalActive(false);
        setSessionToClose(null);
    };

    const handleSessionSort = (key: keyof Session) => {
        let direction = 'asc';
        if (sessionSortConfig.key === key && sessionSortConfig.direction === 'asc') {
            direction = 'desc';
        }
        setSessionSortConfig({ key, direction });
    };

    const handleServerSort = (key: keyof Server) => {
        let direction = 'asc';
        if (serverSortConfig.key === key && serverSortConfig.direction === 'asc') {
            direction = 'desc';
        }
        setServerSortConfig({ key, direction });
    };

    const sortedSessions = () => {
        if (!sessions) return [];
        let sortableItems = [...sessions];
        if (sessionSortConfig.key !== null) {
            sortableItems.sort((a, b) => {
                if (a[sessionSortConfig.key] < b[sessionSortConfig.key]) {
                    return sessionSortConfig.direction === 'asc' ? -1 : 1;
                }
                if (a[sessionSortConfig.key] > b[sessionSortConfig.key]) {
                    return sessionSortConfig.direction === 'asc' ? 1 : -1;
                }
                return 0;
            });
        }
        return sortableItems;
    };

    const filteredSessions = sortedSessions().filter(session =>
        (serverIdFilter === '' || session.server_id?.toLowerCase().includes(serverIdFilter.toLowerCase())) &&
        (ownerFilter === '' || session.owner?.toLowerCase().includes(ownerFilter.toLowerCase())) &&
        (connectionsFilter === '' || session["num-of-connections"] >= parseInt(connectionsFilter, 10))
    );

    const sortedServers = () => {
        if (!servers) return [];
        let sortableItems = [...servers];
        if (serverSortConfig.key !== null) {
            sortableItems.sort((a, b) => {
                if (a[serverSortConfig.key] < b[serverSortConfig.key]) {
                    return serverSortConfig.direction === 'asc' ? -1 : 1;
                }
                if (a[serverSortConfig.key] > b[serverSortConfig.key]) {
                    return serverSortConfig.direction === 'asc' ? 1 : -1;
                }
                return 0;
            });
        }
        return sortableItems;
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
                    <a className="navbar-item is-active">Dashboard</a>
                    <a className="navbar-item" onClick={() => navigate('/agents')}>Agents</a>
                </div>
                <div className="navbar-end">
                    <div className="navbar-item">
                        <button className="button is-ghost" onClick={toggleTheme} title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}>
                            <FontAwesomeIcon icon={theme === 'light' ? faMoon : faSun} />
                        </button>
                    </div>
                    <div className="navbar-item">
                        <div className="field is-grouped is-flex is-align-items-center">
                            <label className="mr-2 is-size-7">Auto-refresh:</label>
                            <span className="select is-small">
                                <select
                                    value={refreshInterval}
                                    onChange={(e) => setRefreshInterval(parseInt(e.target.value, 10))}
                                >
                                    <option value={5000}>5 seconds</option>
                                    <option value={10000}>10 seconds</option>
                                    <option value={30000}>30 seconds</option>
                                    <option value={60000}>1 minute</option>
                                    <option value={0}>Never</option>
                                </select>
                            </span>
                        </div>
                    </div>
                    <div className="navbar-item">
                        <button className="button is-danger" onClick={handleLogout}>
                            Logout
                        </button>
                    </div>
                </div>
            </nav>

            {/* Modals */}
            <div className={`modal ${isModalActive ? 'is-active' : ''}`}>
                <div className="modal-background" onClick={() => setIsModalActive(false)}></div>
                <div className="modal-card">
                    <header className="modal-card-head">
                        <p className="modal-card-title">Close Request Sent</p>
                        <button className="delete" aria-label="close" onClick={() => setIsModalActive(false)}></button>
                    </header>
                    <section className="modal-card-body">
                        <pre>{closeSessionResult}</pre>
                    </section>
                    <footer className="modal-card-foot">
                        <button className="button is-danger" onClick={() => setIsModalActive(false)}>Close</button>
                    </footer>
                </div>
            </div>

            <div className={`modal ${isConfirmModalActive ? 'is-active' : ''}`}>
                <div className="modal-background" onClick={cancelCloseSession}></div>
                <div className="modal-card">
                    <header className="modal-card-head">
                        <p className="modal-card-title">Confirm Session Close</p>
                        <button className="delete" aria-label="close" onClick={cancelCloseSession}></button>
                    </header>
                    <section className="modal-card-body">
                        <p>Are you sure you want to close this session?</p>
                        {sessionToClose && (
                            <div className="content">
                                <ul>
                                    <li><strong>Session ID:</strong> {sessionToClose.sessionId}</li>
                                    <li><strong>Server ID:</strong> {sessionToClose.serverId}</li>
                                    <li><strong>Owner:</strong> {sessionToClose.owner}</li>
                                </ul>
                            </div>
                        )}
                    </section>
                    <footer className="modal-card-foot">
                        <button className="button is-danger" onClick={confirmCloseSession}>Confirm Close</button>
                        <button className="button" onClick={cancelCloseSession}>Cancel</button>
                    </footer>
                </div>
            </div>

            <section className="section">
                <div className="container">
                    <div className="tabs">
                        <ul>
                            <li className={activeTab === 'sessions' ? 'is-active' : ''}>
                                <a onClick={() => setActiveTab('sessions')}>Sessions</a>
                            </li>
                            <li className={activeTab === 'servers' ? 'is-active' : ''}>
                                <a onClick={() => setActiveTab('servers')}>Servers</a>
                            </li>
                        </ul>
                    </div>

                    {activeTab === 'sessions' && (
                        <div>
                            <div className="columns">
                                <div className="column is-3">
                                    <input className="input" type="text" placeholder="Filter by Server ID" value={serverIdFilter} onChange={e => setServerIdFilter(e.target.value)} />
                                </div>
                                <div className="column is-3">
                                    <input className="input" type="text" placeholder="Filter by Owner" value={ownerFilter} onChange={e => setOwnerFilter(e.target.value)} />
                                </div>
                                <div className="column is-3">
                                    <input className="input" type="number" placeholder="Min Connections" value={connectionsFilter} onChange={e => setConnectionsFilter(e.target.value)} />
                                </div>
                                <div className="column is-3">
                                    <button className="button is-success is-fullwidth" onClick={() => { setServerIdFilter(''); setOwnerFilter(''); setConnectionsFilter(''); }}>
                                        Clear Filters
                                    </button>
                                </div>
                            </div>
                            <div className="table-container">
                                <table className="table is-fullwidth is-striped">
                                    <thead>
                                        <tr>
                                            <th onClick={() => handleSessionSort('id')}>ID</th>
                                            <th onClick={() => handleSessionSort('owner')}>Owner</th>
                                            <th onClick={() => handleSessionSort('last-seen')}>Last update</th>
                                            <th onClick={() => handleSessionSort('creation-time')}>Creation Time</th>
                                            <th onClick={() => handleSessionSort('status')}>Status</th>
                                            <th onClick={() => handleSessionSort('type')}>Type</th>
                                            <th onClick={() => handleSessionSort('num-of-connections')}>Connections</th>
                                            <th onClick={() => handleSessionSort('server_id')}>Server ID</th>
                                            <th>Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {filteredSessions.map(session => (
                                            <tr key={session.uid}>
                                                <td>{session.id}</td>
                                                <td>{session.owner}</td>
                                                <td>{new Date(session["last-seen"]).toLocaleString()}</td>
                                                <td>{new Date(session['creation-time']).toLocaleString()}</td>
                                                <td>{session.status}</td>
                                                <td>{session.type}</td>
                                                <td>{session["num-of-connections"]}</td>
                                                <td>{session.server_id}</td>
                                                <td>
                                                    <button className="button is-small is-danger" onClick={() => handleCloseSession(session.id, session.server_id, session.owner)}>
                                                        Close
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}

                    {activeTab === 'servers' && (
                        <div className="table-container">
                            <table className="table is-fullwidth is-striped">
                                <thead>
                                    <tr>
                                        <th onClick={() => handleServerSort('hostname')}>Hostname</th>
                                        <th onClick={() => handleServerSort('last-seen')}>Last update</th>
                                        <th onClick={() => handleServerSort('sessions')}>Sessions</th>
                                        <th onClick={() => handleServerSort('cores')}>Cores</th>
                                        <th onClick={() => handleServerSort('free_memory')}>Memory</th>
                                        <th onClick={() => handleServerSort('cpu_usage')}>CPU Usage</th>
                                        <th onClick={() => handleServerSort('load1')}>Load (1/5/15)</th>
                                        <th onClick={() => handleServerSort('tags')}>Tags</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {sortedServers().map(server => (
                                        <tr key={server.hostname}>
                                            <td>{server.hostname}</td>
                                            <td>{new Date(server["last-seen"]).toLocaleString()}</td>
                                            <td>{server.sessions.length}</td>
                                            <td>{server.cores}</td>
                                            <td>{Math.round(server.free_memory / 1024 / 1024 / 1024)}GB / {Math.round(server.total_memory / 1024 / 1024 / 1024)}GB</td>
                                            <td>{server.cpu_usage.toFixed(2)}%</td>
                                            <td>{server.load1.toFixed(2)} / {server.load5.toFixed(2)} / {server.load15.toFixed(2)}</td>
                                            <td><div className="tags">{server.tags.map(tag => <span key={tag} className="tag">{tag}</span>)}</div></td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </section>
        </div>
    );
};

export default Dashboard;