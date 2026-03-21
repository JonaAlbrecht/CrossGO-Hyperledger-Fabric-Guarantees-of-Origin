import { useState, FormEvent } from 'react';
import api from '../api';

export default function CertificatesPage() {
    const [tab, setTab] = useState<'cancel' | 'verify'>('cancel');
    const [goType, setGoType] = useState<'electricity' | 'hydrogen'>('electricity');
    const [assetID, setAssetID] = useState('');
    const [amount, setAmount] = useState('');
    const [verifyID, setVerifyID] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [verifyResult, setVerifyResult] = useState<Record<string, unknown> | null>(null);
    const [loading, setLoading] = useState(false);

    const handleCancel = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setMessage(''); setLoading(true);
        try {
            const payload: Record<string, unknown> = { goAssetID: assetID };
            if (goType === 'electricity' && amount) payload.amountMWh = parseFloat(amount);
            if (goType === 'hydrogen' && amount) payload.kilos = parseFloat(amount);

            const { data } = await api.post(`/cancellations/${goType}`, payload);
            setMessage(data.message);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Cancellation failed');
        } finally {
            setLoading(false);
        }
    };

    const handleVerify = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setVerifyResult(null); setLoading(true);
        try {
            const { data } = await api.post('/cancellations/verify', { cancellationStatementID: verifyID });
            setVerifyResult(data);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Verification failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2 className="text-2xl font-bold mb-6">Certificates & Cancellations</h2>

            <div className="flex gap-2 mb-6">
                <button onClick={() => setTab('cancel')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium ${
                        tab === 'cancel' ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Cancel GO
                </button>
                <button onClick={() => setTab('verify')}
                    className={`px-4 py-2 rounded-lg text-sm font-medium ${
                        tab === 'verify' ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}>
                    Verify Certificate
                </button>
            </div>

            {tab === 'cancel' ? (
                <form onSubmit={handleCancel} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Claim Renewable Attributes</h3>
                    <p className="text-sm text-gray-500">
                        Cancelling a GO creates a cancellation statement — an immutable certificate proving
                        your renewable energy consumption.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Type</label>
                        <div className="flex gap-2">
                            <button type="button" onClick={() => setGoType('electricity')}
                                className={`px-3 py-1.5 rounded text-sm ${
                                    goType === 'electricity' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-600'
                                }`}>Electricity</button>
                            <button type="button" onClick={() => setGoType('hydrogen')}
                                className={`px-3 py-1.5 rounded text-sm ${
                                    goType === 'hydrogen' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600'
                                }`}>Hydrogen</button>
                        </div>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                        <input value={assetID} onChange={(e) => setAssetID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            Amount ({goType === 'electricity' ? 'MWh' : 'kg'}) — leave empty for full cancellation
                        </label>
                        <input type="number" step="0.001" value={amount} onChange={(e) => setAmount(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" />
                    </div>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    {message && <p className="text-green-600 text-sm">{message}</p>}
                    <button type="submit" disabled={loading}
                        className="bg-purple-600 hover:bg-purple-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Processing...' : 'Cancel GO & Create Certificate'}
                    </button>
                </form>
            ) : (
                <form onSubmit={handleVerify} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Verify Cancellation Statement</h3>
                    <p className="text-sm text-gray-500">
                        Enter a cancellation statement ID to verify its authenticity on the blockchain.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Statement ID</label>
                        <input value={verifyID} onChange={(e) => setVerifyID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                    {error && <p className="text-red-600 text-sm">{error}</p>}
                    <button type="submit" disabled={loading}
                        className="bg-purple-600 hover:bg-purple-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Verifying...' : 'Verify'}
                    </button>

                    {verifyResult && (
                        <div className="mt-4 p-4 bg-green-50 rounded-lg border border-green-200">
                            <p className="text-green-700 font-medium mb-2">Certificate Verified</p>
                            <pre className="text-xs text-gray-700 overflow-auto">{JSON.stringify(verifyResult, null, 2)}</pre>
                        </div>
                    )}
                </form>
            )}
        </div>
    );
}
