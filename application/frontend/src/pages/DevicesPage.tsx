import { useEffect, useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../api';
import type { Device } from '../types';

export default function DevicesPage() {
    const { user } = useAuth();
    const [devices, setDevices] = useState<Device[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    const [error, setError] = useState('');

    // Form state
    const [deviceType, setDeviceType] = useState('SmartMeter');
    const [ownerOrg, setOwnerOrg] = useState('producer1MSP');
    const [carriers, setCarriers] = useState('electricity');

    const fetchDevices = async () => {
        try {
            const { data } = await api.get('/devices');
            setDevices(Array.isArray(data) ? data : []);
        } catch {
            setDevices([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchDevices(); }, []);

    const handleRegister = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        try {
            await api.post('/devices', {
                deviceType,
                ownerOrgMSP: ownerOrg,
                energyCarriers: carriers.split(',').map((c) => c.trim()),
                attributes: {},
            });
            setShowForm(false);
            await fetchDevices();
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed to register device');
        }
    };

    const handleAction = async (id: string, action: 'revoke' | 'suspend' | 'reactivate') => {
        try {
            await api.put(`/devices/${id}/${action}`);
            await fetchDevices();
        } catch (err: unknown) {
            alert(err instanceof Error ? err.message : `Failed to ${action} device`);
        }
    };

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold">Devices</h2>
                {user?.role === 'issuer' && (
                    <button
                        onClick={() => setShowForm(!showForm)}
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors"
                    >
                        {showForm ? 'Cancel' : '+ Register Device'}
                    </button>
                )}
            </div>

            {showForm && (
                <form onSubmit={handleRegister} className="bg-white rounded-xl shadow-sm border p-6 mb-6 space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Device Type</label>
                            <select value={deviceType} onChange={(e) => setDeviceType(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2">
                                <option value="SmartMeter">Smart Meter</option>
                                <option value="OutputMeter">Output Meter</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Owner Org MSP</label>
                            <select value={ownerOrg} onChange={(e) => setOwnerOrg(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2">
                                <option value="producer1MSP">Producer 1</option>
                                <option value="consumer1MSP">Consumer 1</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Energy Carriers</label>
                            <input value={carriers} onChange={(e) => setCarriers(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" placeholder="electricity, hydrogen" />
                        </div>
                    </div>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    <button type="submit"
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium">
                        Register
                    </button>
                </form>
            )}

            {loading ? (
                <p className="text-gray-400">Loading devices...</p>
            ) : devices.length === 0 ? (
                <p className="text-gray-400">No devices registered yet.</p>
            ) : (
                <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-600">
                            <tr>
                                <th className="text-left px-6 py-3">ID</th>
                                <th className="text-left px-6 py-3">Type</th>
                                <th className="text-left px-6 py-3">Owner</th>
                                <th className="text-left px-6 py-3">Status</th>
                                <th className="text-left px-6 py-3">Carriers</th>
                                {user?.role === 'issuer' && <th className="text-left px-6 py-3">Actions</th>}
                            </tr>
                        </thead>
                        <tbody className="divide-y">
                            {devices.map((d) => (
                                <tr key={d.DeviceID} className="hover:bg-gray-50">
                                    <td className="px-6 py-3 font-mono text-xs">{d.DeviceID}</td>
                                    <td className="px-6 py-3">{d.DeviceType}</td>
                                    <td className="px-6 py-3">{d.OwnerOrgMSP}</td>
                                    <td className="px-6 py-3">
                                        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                                            d.Status === 'active' ? 'bg-green-100 text-green-700' :
                                            d.Status === 'suspended' ? 'bg-yellow-100 text-yellow-700' :
                                            'bg-red-100 text-red-700'
                                        }`}>
                                            {d.Status}
                                        </span>
                                    </td>
                                    <td className="px-6 py-3">{d.EnergyCarriers?.join(', ')}</td>
                                    {user?.role === 'issuer' && (
                                        <td className="px-6 py-3 space-x-2">
                                            {d.Status === 'active' && (
                                                <>
                                                    <button onClick={() => handleAction(d.DeviceID, 'suspend')}
                                                        className="text-yellow-600 hover:underline text-xs">Suspend</button>
                                                    <button onClick={() => handleAction(d.DeviceID, 'revoke')}
                                                        className="text-red-600 hover:underline text-xs">Revoke</button>
                                                </>
                                            )}
                                            {d.Status === 'suspended' && (
                                                <button onClick={() => handleAction(d.DeviceID, 'reactivate')}
                                                    className="text-green-600 hover:underline text-xs">Reactivate</button>
                                            )}
                                        </td>
                                    )}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
