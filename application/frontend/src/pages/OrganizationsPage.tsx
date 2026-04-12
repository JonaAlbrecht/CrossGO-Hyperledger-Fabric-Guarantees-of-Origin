import { useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import api, { extractApiError } from '../api';
import Tooltip from '../components/Tooltip';
import { orgDisplayName } from '../types';

interface RegisteredOrg {
    orgMSP: string;
    displayName: string;
    role: string;
    registeredAt: string;
}

export default function OrganizationsPage() {
    const { user } = useAuth();
    const [showForm, setShowForm] = useState(false);
    const [orgName, setOrgName] = useState('');
    const [orgMSP, setOrgMSP] = useState('');
    const [orgRole, setOrgRole] = useState<'producer' | 'buyer'>('producer');
    const [country, setCountry] = useState('DE');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    // Known orgs (read from chain in production)
    const [registeredOrgs] = useState<RegisteredOrg[]>([
        { orgMSP: 'issuer1MSP', displayName: 'German Issuing Authority (UBA)', role: 'issuer', registeredAt: '2024-01-01' },
        { orgMSP: 'eproducer1MSP', displayName: 'Alpha WindFarm GmbH', role: 'producer', registeredAt: '2024-01-15' },
        { orgMSP: 'hproducer1MSP', displayName: 'Beta Electrolyser B.V.', role: 'producer', registeredAt: '2024-02-01' },
        { orgMSP: 'buyer1MSP', displayName: 'Gamma-Town EnergySupplier Ltd', role: 'buyer', registeredAt: '2024-03-01' },
    ]);

    const handleRegister = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setMessage(''); setLoading(true);
        try {
            const { data } = await api.post('/organizations', {
                orgMSP,
                displayName: orgName,
                role: orgRole,
                country,
                energyCarriers: [],
            });
            setMessage(data.message || 'Organization registration submitted');
            setShowForm(false);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to register organization'));
        } finally {
            setLoading(false);
        }
    };

    const roleColor = (role: string) => {
        switch (role) {
            case 'issuer': return 'bg-purple-100 text-purple-700';
            case 'producer': return 'bg-green-100 text-green-700';
            case 'buyer': return 'bg-blue-100 text-blue-700';
            default: return 'bg-gray-100 text-gray-700';
        }
    };

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h2 className="text-2xl font-bold">Organizations</h2>
                    <p className="text-sm text-gray-500 mt-1">
                        View registered organizations and register new participants in the GO network.
                        Only issuing authorities can register new organizations.
                    </p>
                </div>
                {user?.role === 'issuer' && (
                    <Tooltip text="Register a new producer or buyer organization in the network" position="left">
                        <button onClick={() => setShowForm(!showForm)}
                            className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors">
                            {showForm ? 'Cancel' : '+ Register Organization'}
                        </button>
                    </Tooltip>
                )}
            </div>

            {message && (
                <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
                    <p className="text-green-700 text-sm">{message}</p>
                </div>
            )}

            {showForm && (
                <form onSubmit={handleRegister} className="bg-white rounded-xl shadow-sm border p-6 mb-6 space-y-4">
                    <h3 className="font-semibold text-gray-700">Register New Organization</h3>
                    <p className="text-sm text-gray-500">
                        Adding a new organization provisions their MSP identity in the Hyperledger Fabric network
                        and assigns them a role (producer or buyer).
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Organization Name</label>
                            <input value={orgName} onChange={(e) => setOrgName(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" placeholder="e.g. Delta Solar Park AG" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">MSP Identifier</label>
                            <input value={orgMSP} onChange={(e) => setOrgMSP(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" placeholder="e.g. producer2MSP" required />
                            <p className="text-xs text-gray-400 mt-1">Must match the Fabric MSP ID</p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
                            <select value={orgRole} onChange={(e) => setOrgRole(e.target.value as 'producer' | 'buyer')}
                                className="w-full border rounded-lg px-3 py-2">
                                <option value="producer">Producer</option>
                                <option value="buyer">Buyer</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Country</label>
                            <input value={country} onChange={(e) => setCountry(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" placeholder="ISO 3166-1 alpha-2" />
                            <p className="text-xs text-gray-400 mt-1">ISO 3166-1 code (e.g. DE, NL)</p>
                        </div>
                    </div>
                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                            <p className="text-red-700 text-sm font-medium">Registration failed</p>
                            <p className="text-red-600 text-sm mt-1">{error}</p>
                        </div>
                    )}
                    <button type="submit" disabled={loading}
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Registering...' : 'Register'}
                    </button>
                </form>
            )}

            {/* Org table */}
            <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
                <table className="w-full text-sm">
                    <thead className="bg-gray-50 text-gray-600">
                        <tr>
                            <th className="text-left px-6 py-3">Organization</th>
                            <th className="text-left px-6 py-3">MSP ID</th>
                            <th className="text-left px-6 py-3">Role</th>
                            <th className="text-left px-6 py-3">Registered</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y">
                        {registeredOrgs.map((org) => (
                            <tr key={org.orgMSP} className="hover:bg-gray-50">
                                <td className="px-6 py-3 font-medium">{org.displayName}</td>
                                <td className="px-6 py-3 font-mono text-xs text-gray-500">{org.orgMSP}</td>
                                <td className="px-6 py-3">
                                    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${roleColor(org.role)}`}>
                                        {org.role}
                                    </span>
                                </td>
                                <td className="px-6 py-3 text-gray-500">{org.registeredAt}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
