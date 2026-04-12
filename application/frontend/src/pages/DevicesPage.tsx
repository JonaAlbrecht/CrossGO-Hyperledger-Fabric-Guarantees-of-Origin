import { useEffect, useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import api, { extractApiError } from '../api';
import type { Device, EnergyCarrier } from '../types';
import { ENERGY_CARRIERS, orgDisplayName } from '../types';
import Tooltip from '../components/Tooltip';

const OWNER_ORGS = [
    { value: 'eproducer1MSP', label: 'Alpha WindFarm GmbH' },
    { value: 'hproducer1MSP', label: 'Beta Electrolyser B.V.' },
    { value: 'buyer1MSP', label: 'Gamma-Town EnergySupplier Ltd' },
];

const DEVICE_TYPES = [
    { value: 'SmartMeter', label: 'Smart Meter', tooltip: 'Measures energy input or production output' },
    { value: 'OutputMeter', label: 'Output Meter', tooltip: 'Measures the output of a conversion device' },
    { value: 'ConversionDevice', label: 'Conversion Device', tooltip: 'Converts one energy carrier into another (e.g. electrolyser)' },
];

export default function DevicesPage() {
    const { user } = useAuth();
    const [devices, setDevices] = useState<Device[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    const [error, setError] = useState('');

    // Form state
    const [deviceType, setDeviceType] = useState('SmartMeter');
    const [ownerOrg, setOwnerOrg] = useState('eproducer1MSP');
    const [carriers, setCarriers] = useState<string[]>(['electricity']);
    const [actionError, setActionError] = useState('');

    // Conversion efficiency fields — one per target carrier
    const [convEfficiencies, setConvEfficiencies] = useState<Record<string, string>>({});

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
            // Build attributes including conversion efficiencies
            const attributes: Record<string, string> = {};
            for (const [carrier, eff] of Object.entries(convEfficiencies)) {
                if (eff && parseFloat(eff) > 0) {
                    attributes[`conversionEfficiencyTo_${carrier}`] = eff;
                }
            }

            await api.post('/devices', {
                deviceType,
                ownerOrgMSP: ownerOrg,
                energyCarriers: carriers,
                attributes,
            });
            setShowForm(false);
            setConvEfficiencies({});
            await fetchDevices();
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to register device'));
        }
    };

    const handleAction = async (id: string, action: 'revoke' | 'suspend' | 'reactivate') => {
        setActionError('');
        try {
            await api.put(`/devices/${id}/${action}`);
            await fetchDevices();
        } catch (err: unknown) {
            setActionError(extractApiError(err, `Failed to ${action} device`));
        }
    };

    const toggleCarrier = (c: string) => {
        setCarriers((prev) =>
            prev.includes(c) ? prev.filter((x) => x !== c) : [...prev, c],
        );
    };

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h2 className="text-2xl font-bold">Devices</h2>
                    <p className="text-sm text-gray-500 mt-1">
                        Register and manage metering devices that attest energy production for Guarantees of Origin
                    </p>
                </div>
                {user?.role === 'issuer' && (
                    <Tooltip text="Register a new metering or conversion device for a producer organization" position="left">
                        <button
                            onClick={() => setShowForm(!showForm)}
                            className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors"
                        >
                            {showForm ? 'Cancel' : '+ Register Device'}
                        </button>
                    </Tooltip>
                )}
            </div>

            {actionError && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
                    <p className="text-red-700 text-sm font-medium">Action failed</p>
                    <p className="text-red-600 text-sm mt-1">{actionError}</p>
                </div>
            )}

            {showForm && (
                <form onSubmit={handleRegister} className="bg-white rounded-xl shadow-sm border p-6 mb-6 space-y-4">
                    <h3 className="font-semibold text-gray-700">Register New Device</h3>
                    <p className="text-sm text-gray-500">
                        A device is a registered metering instrument that cryptographically attests energy production data.
                        Only the issuer can register devices.
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Device Type</label>
                            <select value={deviceType} onChange={(e) => setDeviceType(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2">
                                {DEVICE_TYPES.map((dt) => (
                                    <option key={dt.value} value={dt.value}>{dt.label}</option>
                                ))}
                            </select>
                            <p className="text-xs text-gray-400 mt-1">
                                {DEVICE_TYPES.find((dt) => dt.value === deviceType)?.tooltip}
                            </p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Owner Organization</label>
                            <select value={ownerOrg} onChange={(e) => setOwnerOrg(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2">
                                {OWNER_ORGS.map((org) => (
                                    <option key={org.value} value={org.value}>{org.label}</option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Energy Carriers</label>
                            <div className="flex flex-wrap gap-2 mt-1">
                                {ENERGY_CARRIERS.map((c) => (
                                    <button key={c.value} type="button" onClick={() => toggleCarrier(c.value)}
                                        className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                                            carriers.includes(c.value)
                                                ? `${c.bgColor} ${c.color} ring-2 ring-offset-1 ring-gray-300`
                                                : 'bg-gray-100 text-gray-500'
                                        }`}>
                                        {c.label}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Conversion efficiency fields */}
                    <div>
                        <Tooltip text="Conversion efficiency defines how much of one energy carrier this device can produce from another. E.g., an electrolyser might convert electricity to hydrogen at 65% efficiency." position="right">
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Conversion Efficiencies (%)
                            </label>
                        </Tooltip>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                            {ENERGY_CARRIERS.map((c) => (
                                <div key={c.value}>
                                    <label className="block text-xs text-gray-500 mb-1">→ {c.label}</label>
                                    <input
                                        type="number" step="0.01" min="0" max="100"
                                        placeholder="—"
                                        value={convEfficiencies[c.value] ?? ''}
                                        onChange={(e) => setConvEfficiencies((prev) => ({ ...prev, [c.value]: e.target.value }))}
                                        className="w-full border rounded-lg px-3 py-1.5 text-sm"
                                    />
                                </div>
                            ))}
                        </div>
                        <p className="text-xs text-gray-400 mt-1">
                            Leave empty if this device cannot convert to that carrier.
                        </p>
                    </div>

                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                            <p className="text-red-700 text-sm font-medium">Registration failed</p>
                            <p className="text-red-600 text-sm mt-1">{error}</p>
                        </div>
                    )}
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
                                <tr key={d.deviceID} className="hover:bg-gray-50">
                                    <td className="px-6 py-3 font-mono text-xs">{d.deviceID}</td>
                                    <td className="px-6 py-3">{d.deviceType}</td>
                                    <td className="px-6 py-3">{orgDisplayName(d.ownerOrgMSP)}</td>
                                    <td className="px-6 py-3">
                                        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                                            d.status === 'active' ? 'bg-green-100 text-green-700' :
                                            d.status === 'suspended' ? 'bg-yellow-100 text-yellow-700' :
                                            'bg-red-100 text-red-700'
                                        }`}>
                                            {d.status}
                                        </span>
                                    </td>
                                    <td className="px-6 py-3">{d.energyCarriers?.join(', ')}</td>
                                    {user?.role === 'issuer' && (
                                        <td className="px-6 py-3 space-x-2">
                                            {d.status === 'active' && (
                                                <>
                                                    <button onClick={() => handleAction(d.deviceID, 'suspend')}
                                                        className="text-yellow-600 hover:underline text-xs">Suspend</button>
                                                    <button onClick={() => handleAction(d.deviceID, 'revoke')}
                                                        className="text-red-600 hover:underline text-xs">Revoke</button>
                                                </>
                                            )}
                                            {d.status === 'suspended' && (
                                                <button onClick={() => handleAction(d.deviceID, 'reactivate')}
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
