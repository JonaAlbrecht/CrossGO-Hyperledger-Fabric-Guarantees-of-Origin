import { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import api from '../api';
import type { UserSession } from '../types';

interface AuthContextType {
    user: UserSession | null;
    login: (orgName: string, userName: string) => Promise<void>;
    logout: () => void;
    isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<UserSession | null>(() => {
        const stored = localStorage.getItem('go_user');
        return stored ? JSON.parse(stored) : null;
    });

    const login = useCallback(async (orgName: string, userName: string) => {
        const { data } = await api.post('/auth/login', { orgName, userName });
        const session: UserSession = {
            token: data.token,
            mspId: data.mspId,
            orgName,
            userName,
            role: data.role,
        };
        localStorage.setItem('go_token', data.token);
        localStorage.setItem('go_user', JSON.stringify(session));
        setUser(session);
    }, []);

    const logout = useCallback(() => {
        localStorage.removeItem('go_token');
        localStorage.removeItem('go_user');
        setUser(null);
    }, []);

    return (
        <AuthContext.Provider value={{ user, login, logout, isAuthenticated: !!user }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextType {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error('useAuth must be used within AuthProvider');
    return ctx;
}
