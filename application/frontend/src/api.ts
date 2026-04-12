import axios, { AxiosError } from 'axios';

const api = axios.create({ baseURL: '/api' });

// Extract a human-readable error message from Axios errors.
// Prefers the backend's `error` or `details` field over generic status text.
export function extractApiError(err: unknown, fallback = 'An unexpected error occurred'): string {
    if (axios.isAxiosError(err)) {
        const axErr = err as AxiosError<{ error?: string; details?: string; message?: string }>;
        const data = axErr.response?.data;
        if (data) {
            const raw = data.error ?? data.details ?? data.message ?? '';
            // Extract the chaincode detail from Fabric Gateway gRPC errors
            // Format 1: "chaincode response 500, <message>"
            const ccDetail = raw.match(/chaincode response \d+, (.+)/)?.[1];
            if (ccDetail) return ccDetail;
            // Format 2: "10 ABORTED: ... — <per-peer details>"
            const fabricDetail = raw.match(/\d+ [A-Z]+:.*? — (.+)/)?.[1];
            if (fabricDetail) return fabricDetail;
            // Format 3: Long gRPC error — extract after last colon
            const afterDash = raw.match(/ — (.+)/)?.[1];
            if (afterDash) return afterDash;
            if (raw) return raw;
        }
        if (axErr.response) return `Server error (${axErr.response.status})`;
        if (axErr.request) return 'No response from server — check network connection';
    }
    if (err instanceof Error) return err.message;
    return fallback;
}

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
