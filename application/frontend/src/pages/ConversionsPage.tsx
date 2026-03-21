import { useState, FormEvent } from 'react';
import api from '../api';

export default function ConversionsPage() {
    const [tab, setTab] = useState<'backlog' | 'issue'>('backlog');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    // Backlog form
    const [kilos, setKilos] = useState('');
    const [hMethod, setHMethod] = useState('electrolysis');
    const [mwh, setMwh] = useState('');

    const handleAddBacklog = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setMessage(''); setLoading(true);
        try {
            const { data } = await api.post('/conversions/backlog', {
                kilosHydrogen: parseFloat(kilos),
                hydrogenProductionMethod: hMethod,
                mwhElectricity: parseFloat(mwh),
            });
            setMessage(data.message);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed');
        } finally {
            setLoading(false);
        }
    };

    const handleIssue = async () => {
        setError(''); setMessage(''); setLoading(true);
        try {
            const { data } = await api.post('/conversions/issue');
            setMessage(data.message);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2 className="text-2xl font-bold mb-6">Electricity → Hydrogen Conversion</h2>

            <div className="flex gap-2 mb-6">
                <button onClick={() => setTab('backlog')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium ${
                        tab === 'backlog' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Add to Backlog
                </button>
                <button onClick={() => setTab('issue')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium ${
                        tab === 'issue' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Issue Hydrogen GO
                </button>
            </div>

            {tab === 'backlog' ? (
                <form onSubmit={handleAddBacklog} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Add Hydrogen Production to Backlog</h3>
                    <p className="text-sm text-gray-500">
                        Record hydrogen production that consumed electricity. The platform will match
                        this against your electricity GOs to issue hydrogen GOs.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Hydrogen Produced (kg)</label>
                        <input type="number" step="0.001" value={kilos} onChange={(e) => setKilos(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Production Method</label>
                        <select value={hMethod} onChange={(e) => setHMethod(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2">
                            <option value="electrolysis">Electrolysis</option>
                            <option value="smr">Steam Methane Reforming</option>
                            <option value="biomass_gasification">Biomass Gasification</option>
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Electricity Used (MWh)</label>
                        <input type="number" step="0.001" value={mwh} onChange={(e) => setMwh(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    {message && <p className="text-green-600 text-sm">{message}</p>}
                    <button type="submit" disabled={loading}
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Submitting...' : 'Add to Backlog'}
                    </button>
                </form>
            ) : (
                <div className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Issue Hydrogen GO from Backlog</h3>
                    <p className="text-sm text-gray-500">
                        This will consume electricity GOs from your collection to cover the hydrogen
                        backlog, then mint a new hydrogen GO. Consumption declarations are created automatically.
                    </p>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    {message && <p className="text-green-600 text-sm">{message}</p>}
                    <button onClick={handleIssue} disabled={loading}
                        className="bg-blue-600 hover:bg-blue-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Processing...' : 'Issue Hydrogen GO'}
                    </button>
                </div>
            )}
        </div>
    );
}
