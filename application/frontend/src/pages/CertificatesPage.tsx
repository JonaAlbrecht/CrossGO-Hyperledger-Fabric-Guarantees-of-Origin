import { useState, FormEvent } from 'react';
import api, { extractApiError } from '../api';
import { ENERGY_CARRIERS } from '../types';
import type { EnergyCarrier } from '../types';
import Tooltip from '../components/Tooltip';

const AMOUNT_UNITS: Record<EnergyCarrier, string> = {
    electricity: 'MWh',
    hydrogen: 'kg',
    biogas: 'm³',
    heating_cooling: 'MWh',
};

const AMOUNT_API_KEY: Record<EnergyCarrier, string> = {
    electricity: 'amountMWh',
    hydrogen: 'kilos',
    biogas: 'cubicMeters',
    heating_cooling: 'amountMWh',
};

export default function CancellationsPage() {
    const [carrier, setCarrier] = useState<EnergyCarrier>('electricity');
    const [assetID, setAssetID] = useState('');
    const [amount, setAmount] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleCancel = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setMessage(''); setLoading(true);
        try {
            const payload: Record<string, unknown> = { goAssetID: assetID };
            if (amount) payload[AMOUNT_API_KEY[carrier]] = parseFloat(amount);

            const { data } = await api.post(`/cancellations/${carrier}`, payload);
            setMessage(data.message);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Cancellation failed'));
        } finally {
            setLoading(false);
        }
    };

    const activeCarrier = ENERGY_CARRIERS.find((c) => c.value === carrier)!;

    return (
        <div>
            <div className="mb-6">
                <h2 className="text-2xl font-bold">Cancellation Statements</h2>
                <p className="text-sm text-gray-500 mt-1">
                    Cancel a Guarantee of Origin to claim its renewable energy attributes.
                    Cancelling creates an immutable Cancellation Statement on the blockchain.
                </p>
            </div>

            {/* Carrier selector */}
            <div className="mb-6">
                <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">Energy Carrier</label>
                <div className="flex gap-2">
                    {ENERGY_CARRIERS.map((c) => (
                        <button key={c.value} onClick={() => setCarrier(c.value)}
                            className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                                carrier === c.value
                                    ? `${c.bgColor} ${c.color} ring-2 ring-offset-1 ring-gray-300`
                                    : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                            }`}>
                            {c.label}
                        </button>
                    ))}
                </div>
            </div>

            <form onSubmit={handleCancel} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                <h3 className="font-semibold text-gray-700">Cancel {activeCarrier.label} GO</h3>
                <p className="text-sm text-gray-500">
                    Cancelling this GO proves that you consumed renewable {activeCarrier.label.toLowerCase()} energy.
                    The Cancellation Statement can be verified by any auditor on the blockchain.
                </p>

                <div>
                    <Tooltip text="The unique identifier of the Guarantee of Origin you want to cancel" position="right">
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                    </Tooltip>
                    <input value={assetID} onChange={(e) => setAssetID(e.target.value)}
                        className="w-full border rounded-lg px-3 py-2" required />
                </div>
                <div>
                    <Tooltip text="Leave empty to cancel the full GO. Specify an amount for partial cancellation." position="right">
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            Amount ({AMOUNT_UNITS[carrier]}) — leave empty for full cancellation
                        </label>
                    </Tooltip>
                    <input type="number" step="0.001" value={amount} onChange={(e) => setAmount(e.target.value)}
                        className="w-full border rounded-lg px-3 py-2" />
                </div>
                {error && (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p className="text-red-700 text-sm font-medium">Error</p>
                        <p className="text-red-600 text-sm mt-1">{error}</p>
                    </div>
                )}
                {message && (
                    <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                        <p className="text-green-700 text-sm">{message}</p>
                    </div>
                )}
                <button type="submit" disabled={loading}
                    className="bg-purple-600 hover:bg-purple-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                    {loading ? 'Processing...' : 'Cancel GO & Create Statement'}
                </button>
            </form>
        </div>
    );
}
