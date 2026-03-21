import { useState, FormEvent } from 'react';
import api from '../api';

export default function TransfersPage() {
    const [mode, setMode] = useState<'single' | 'eAmount' | 'hAmount'>('single');
    const [recipientMSP, setRecipientMSP] = useState('consumer1MSP');
    const [goAssetID, setGoAssetID] = useState('');
    const [amount, setAmount] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');
        setLoading(true);
        try {
            if (mode === 'single') {
                const { data } = await api.post('/transfers', { goAssetID, recipientMSP });
                setMessage(data.message);
            } else if (mode === 'eAmount') {
                const { data } = await api.post('/transfers/electricity-by-amount', {
                    recipientMSP,
                    amountMWh: parseFloat(amount),
                });
                setMessage(data.message);
            } else {
                const { data } = await api.post('/transfers/hydrogen-by-amount', {
                    recipientMSP,
                    kilos: parseFloat(amount),
                });
                setMessage(data.message);
            }
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Transfer failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2 className="text-2xl font-bold mb-6">Transfer GOs</h2>

            {/* Mode selector */}
            <div className="flex gap-2 mb-6">
                {([
                    { key: 'single', label: 'Single GO' },
                    { key: 'eAmount', label: 'Electricity by Amount' },
                    { key: 'hAmount', label: 'Hydrogen by Amount' },
                ] as const).map(({ key, label }) => (
                    <button key={key} onClick={() => setMode(key)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            mode === key ? 'bg-primary-100 text-primary-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        {label}
                    </button>
                ))}
            </div>

            <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Recipient Organization</label>
                    <select value={recipientMSP} onChange={(e) => setRecipientMSP(e.target.value)}
                        className="w-full border rounded-lg px-3 py-2">
                        <option value="producer1MSP">Producer 1</option>
                        <option value="consumer1MSP">Consumer 1</option>
                        <option value="issuer1MSP">Issuer 1</option>
                    </select>
                </div>

                {mode === 'single' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                        <input value={goAssetID} onChange={(e) => setGoAssetID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" placeholder="e.g. eGO_1" required />
                    </div>
                )}

                {mode !== 'single' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {mode === 'eAmount' ? 'Amount (MWh)' : 'Amount (kg)'}
                        </label>
                        <input type="number" step="0.001" value={amount} onChange={(e) => setAmount(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                )}

                {error && <p className="text-red-600 text-sm">{error}</p>}
                {message && <p className="text-green-600 text-sm">{message}</p>}

                <button type="submit" disabled={loading}
                    className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                    {loading ? 'Submitting...' : 'Transfer'}
                </button>
            </form>
        </div>
    );
}
