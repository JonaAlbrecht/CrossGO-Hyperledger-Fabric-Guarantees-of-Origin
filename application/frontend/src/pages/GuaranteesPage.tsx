import { useEffect, useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../api';
import type { ElectricityGO, HydrogenGO } from '../types';

export default function GuaranteesPage() {
    const { user } = useAuth();
    const [egos, setEgos] = useState<ElectricityGO[]>([]);
    const [hgos, setHgos] = useState<HydrogenGO[]>([]);
    const [loading, setLoading] = useState(true);
    const [tab, setTab] = useState<'electricity' | 'hydrogen'>('electricity');
    const [showForm, setShowForm] = useState(false);
    const [error, setError] = useState('');

    // Electricity form
    const [amountMWh, setAmountMWh] = useState('');
    const [emissions, setEmissions] = useState('');
    const [elapsed, setElapsed] = useState('');
    const [method, setMethod] = useState('solar');

    const fetchGOs = async () => {
        try {
            const { data } = await api.get('/guarantees');
            setEgos(data.electricityGOs ?? []);
            setHgos(data.hydrogenGOs ?? []);
        } catch {
            setEgos([]);
            setHgos([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchGOs(); }, []);

    const handleCreateEGO = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        try {
            await api.post('/guarantees/electricity', {
                amountMWh: parseFloat(amountMWh),
                emissions: parseFloat(emissions),
                elapsedSeconds: parseFloat(elapsed),
                electricityProductionMethod: method,
            });
            setShowForm(false);
            await fetchGOs();
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed to create GO');
        }
    };

    const formatDate = (ts: number) => new Date(ts * 1000).toLocaleString();

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold">Guarantees of Origin</h2>
                {user?.role === 'producer' && (
                    <button onClick={() => setShowForm(!showForm)}
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors">
                        {showForm ? 'Cancel' : '+ Create GO'}
                    </button>
                )}
            </div>

            {showForm && (
                <form onSubmit={handleCreateEGO} className="bg-white rounded-xl shadow-sm border p-6 mb-6 space-y-4">
                    <h3 className="font-semibold text-gray-700">Create Electricity GO</h3>
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Amount (MWh)</label>
                            <input type="number" step="0.001" value={amountMWh} onChange={(e) => setAmountMWh(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Emissions (kg CO2)</label>
                            <input type="number" step="0.001" value={emissions} onChange={(e) => setEmissions(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Elapsed (seconds)</label>
                            <input type="number" step="1" value={elapsed} onChange={(e) => setElapsed(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Production Method</label>
                            <select value={method} onChange={(e) => setMethod(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2">
                                <option value="solar">Solar</option>
                                <option value="wind">Wind</option>
                                <option value="hydro">Hydro</option>
                                <option value="biomass">Biomass</option>
                                <option value="geothermal">Geothermal</option>
                            </select>
                        </div>
                    </div>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    <button type="submit"
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium">
                        Submit to Blockchain
                    </button>
                </form>
            )}

            {/* Tabs */}
            <div className="flex gap-2 mb-4">
                <button onClick={() => setTab('electricity')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                        tab === 'electricity' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Electricity ({egos.length})
                </button>
                <button onClick={() => setTab('hydrogen')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                        tab === 'hydrogen' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Hydrogen ({hgos.length})
                </button>
            </div>

            {loading ? (
                <p className="text-gray-400">Loading guarantees...</p>
            ) : (
                <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-600">
                            <tr>
                                <th className="text-left px-6 py-3">Asset ID</th>
                                <th className="text-left px-6 py-3">Type</th>
                                <th className="text-left px-6 py-3">Created</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y">
                            {(tab === 'electricity' ? egos : hgos).map((go) => (
                                <tr key={go.AssetID} className="hover:bg-gray-50">
                                    <td className="px-6 py-3 font-mono text-xs">{go.AssetID}</td>
                                    <td className="px-6 py-3">
                                        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                                            go.GOType === 'electricity' ? 'bg-yellow-100 text-yellow-700' : 'bg-blue-100 text-blue-700'
                                        }`}>
                                            {go.GOType}
                                        </span>
                                    </td>
                                    <td className="px-6 py-3 text-gray-500">{formatDate(go.CreationDateTime)}</td>
                                </tr>
                            ))}
                            {(tab === 'electricity' ? egos : hgos).length === 0 && (
                                <tr><td colSpan={3} className="px-6 py-8 text-center text-gray-400">No {tab} GOs found</td></tr>
                            )}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
