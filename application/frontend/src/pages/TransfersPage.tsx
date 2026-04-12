import { useState, FormEvent } from 'react';
import api, { extractApiError } from '../api';
import { ENERGY_CARRIERS, orgDisplayName } from '../types';
import type { EnergyCarrier } from '../types';
import Tooltip from '../components/Tooltip';

const RECIPIENT_ORGS = [
    { value: 'eproducer1MSP', label: 'Alpha WindFarm GmbH' },
    { value: 'hproducer1MSP', label: 'Beta Electrolyser B.V.' },
    { value: 'buyer1MSP', label: 'Gamma-Town EnergySupplier Ltd' },
    { value: 'issuer1MSP', label: 'German Issuing Authority (UBA)' },
];

const AMOUNT_UNITS: Record<EnergyCarrier, string> = {
    electricity: 'MWh',
    hydrogen: 'kg',
    biogas: 'm³',
    heating_cooling: 'MWh',
};

const AMOUNT_ENDPOINTS: Record<EnergyCarrier, string> = {
    electricity: '/transfers/electricity-by-amount',
    hydrogen: '/transfers/hydrogen-by-amount',
    biogas: '/transfers/biogas-by-amount',
    heating_cooling: '/transfers/heating-cooling-by-amount',
};

const AMOUNT_BODY_KEY: Record<EnergyCarrier, string> = {
    electricity: 'amountMWh',
    hydrogen: 'kilos',
    biogas: 'cubicMeters',
    heating_cooling: 'amountMWh',
};

type TransferMode = 'single' | 'byAmount';

export default function TransfersPage() {
    const [mode, setMode] = useState<TransferMode>('single');
    const [carrier, setCarrier] = useState<EnergyCarrier>('electricity');
    const [recipientMSP, setRecipientMSP] = useState('buyer1MSP');
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
            } else {
                const { data } = await api.post(AMOUNT_ENDPOINTS[carrier], {
                    recipientMSP,
                    [AMOUNT_BODY_KEY[carrier]]: parseFloat(amount),
                });
                setMessage(data.message);
            }
        } catch (err: unknown) {
            setError(extractApiError(err, 'Transfer failed'));
        } finally {
            setLoading(false);
        }
    };

    const activeCarrier = ENERGY_CARRIERS.find((c) => c.value === carrier)!;

    return (
        <div>
            <div className="mb-6">
                <h2 className="text-2xl font-bold">Transfer Guarantees of Origin</h2>
                <p className="text-sm text-gray-500 mt-1">
                    Transfer ownership of Guarantees of Origin to another registered organization
                </p>
            </div>

            {/* Mode selector */}
            <div className="flex gap-2 mb-4">
                <Tooltip text="Transfer a single GO by its asset ID" position="bottom">
                    <button onClick={() => setMode('single')}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            mode === 'single' ? 'bg-primary-100 text-primary-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        Single GO
                    </button>
                </Tooltip>
                <Tooltip text="Transfer a batch of GOs by specifying a total amount for a given energy carrier" position="bottom">
                    <button onClick={() => setMode('byAmount')}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            mode === 'byAmount' ? 'bg-primary-100 text-primary-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        By Amount
                    </button>
                </Tooltip>
            </div>

            {/* Carrier selector — only in byAmount mode */}
            {mode === 'byAmount' && (
                <div className="flex gap-2 mb-6">
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
            )}

            <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Recipient Organization</label>
                    <select value={recipientMSP} onChange={(e) => setRecipientMSP(e.target.value)}
                        className="w-full border rounded-lg px-3 py-2">
                        {RECIPIENT_ORGS.map((org) => (
                            <option key={org.value} value={org.value}>{org.label}</option>
                        ))}
                    </select>
                </div>

                {mode === 'single' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                        <input value={goAssetID} onChange={(e) => setGoAssetID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" placeholder="e.g. eGO_1" required />
                    </div>
                )}

                {mode === 'byAmount' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            Amount ({AMOUNT_UNITS[carrier]})
                        </label>
                        <input type="number" step="0.001" value={amount} onChange={(e) => setAmount(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                        <p className="text-xs text-gray-400 mt-1">
                            {activeCarrier.label} GOs will be selected from your balance in FIFO order.
                        </p>
                    </div>
                )}

                {error && (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p className="text-red-700 text-sm font-medium">Transfer failed</p>
                        <p className="text-red-600 text-sm mt-1">{error}</p>
                    </div>
                )}
                {message && (
                    <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                        <p className="text-green-700 text-sm">{message}</p>
                    </div>
                )}

                <button type="submit" disabled={loading}
                    className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                    {loading ? 'Submitting...' : 'Transfer'}
                </button>
            </form>
        </div>
    );
}
