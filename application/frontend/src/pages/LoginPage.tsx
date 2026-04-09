import { useState, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const ORGS = [
    { value: 'issuer1', label: 'Issuer 1 (Certification Body)' },
    { value: 'eproducer1', label: 'E-Producer 1 (Electricity Producer)' },
    { value: 'hproducer1', label: 'H-Producer 1 (Hydrogen Producer)' },
    { value: 'buyer1', label: 'Buyer 1 (Energy Consumer)' },
];

export default function LoginPage() {
    const { login } = useAuth();
    const navigate = useNavigate();
    const [orgName, setOrgName] = useState('producer1');
    const [userName, setUserName] = useState('Admin');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);
        try {
            await login(orgName, userName);
            navigate('/');
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : 'Login failed';
            setError(msg);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-800 to-primary-900">
            <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
                <div className="text-center mb-8">
                    <h1 className="text-3xl font-bold text-primary-800">GO Platform</h1>
                    <p className="text-gray-500 mt-2">Guarantee of Origin Management System</p>
                </div>

                <form onSubmit={handleSubmit} className="space-y-6">
                    <div>
                        <label htmlFor="org" className="block text-sm font-medium text-gray-700 mb-1">
                            Organization
                        </label>
                        <select
                            id="org"
                            value={orgName}
                            onChange={(e) => setOrgName(e.target.value)}
                            className="w-full border border-gray-300 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        >
                            {ORGS.map((org) => (
                                <option key={org.value} value={org.value}>
                                    {org.label}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div>
                        <label htmlFor="user" className="block text-sm font-medium text-gray-700 mb-1">
                            User Name
                        </label>
                        <input
                            id="user"
                            type="text"
                            value={userName}
                            onChange={(e) => setUserName(e.target.value)}
                            className="w-full border border-gray-300 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                            placeholder="e.g. Admin"
                            required
                        />
                    </div>

                    {error && (
                        <div className="bg-red-50 text-red-700 text-sm rounded-lg px-4 py-3">{error}</div>
                    )}

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg px-4 py-2.5 transition-colors disabled:opacity-50"
                    >
                        {loading ? 'Connecting to Fabric...' : 'Sign In'}
                    </button>
                </form>

                <p className="text-center text-xs text-gray-400 mt-6">
                    Hyperledger Fabric 2.x · Tiered GO Network
                </p>
            </div>
        </div>
    );
}
