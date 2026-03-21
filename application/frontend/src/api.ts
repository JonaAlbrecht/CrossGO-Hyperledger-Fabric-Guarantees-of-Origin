import axios from 'axios';

const api = axios.create({ baseURL: '/api' });

// Attach JWT to every request
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('go_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Auto-logout on 401
api.interceptors.response.use(
    (res) => res,
    (err) => {
        if (err.response?.status === 401) {
            localStorage.removeItem('go_token');
            localStorage.removeItem('go_user');
            window.location.href = '/login';
        }
        return Promise.reject(err);
    }
);

export default api;
