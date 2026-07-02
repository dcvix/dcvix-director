import axios from 'axios';
import type { Agent, Server, Session } from '../types';

const apiClient = axios.create({
    baseURL: '/v1',
    withCredentials: true,  // Needed for cookies
});

// apiClient.interceptors.request.use(config => {
//     const token = localStorage.getItem('token');
//     if (token) {
//         config.headers.Authorization = `Bearer ${token}`;
//     }
//     return config;
// });
// Response interceptor to handle 401 (Unauthorized)
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        const url = error.config?.url || '';
        const isAuthHealthCheck = url === '/auth-health' || url.endsWith('/auth-health');
        const isLoginEndpoint = url.includes('/login');
        if (error.response?.status === 401 && !isAuthHealthCheck && !isLoginEndpoint) {
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export const login = async (userID: string, password: string, otp: string) => {
    await apiClient.post('/admin/login', { userID, password, otp });
};

export const checkAuth = async (): Promise<boolean> => {
    try {
        await apiClient.get('/auth-health');
        return true;
    } catch {
        return false;
    }
};

// Remove session cookie
export const logout = async (): Promise<void> => {
    try {
        await apiClient.post('/logout');
    } catch (error) {
        console.error('Logout failed:', error);
    }
};

export const getSessions = async (): Promise<Session[]> => {
    const response = await apiClient.get('/sessions');
    return response.data;
};

export const getServers = async (): Promise<Server[]> => {
    const response = await apiClient.get('/servers');
    return response.data;
};

export const closeSession = async (sessionId: string, serverId: string) => {
    const response = await apiClient.delete(`/sessions/${sessionId}`, {
        data: { ServerID: serverId },
    });
    return response.data;
};

export const getAgents = async (state?: string): Promise<Agent[]> => {
    const params = state ? { state } : {};
    const response = await apiClient.get('/admin/agents', { params });
    return response.data;
};

export const approveAgent = async (guid: string): Promise<void> => {
    await apiClient.post(`/admin/agents/${guid}/approve`);
};

export const denyAgent = async (guid: string): Promise<void> => {
    await apiClient.post(`/admin/agents/${guid}/deny`);
};

export const revokeAgent = async (guid: string): Promise<void> => {
    await apiClient.post(`/admin/agents/${guid}/revoke`);
};