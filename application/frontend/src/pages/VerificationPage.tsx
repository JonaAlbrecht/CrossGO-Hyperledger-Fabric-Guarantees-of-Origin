import { useState, FormEvent } from 'react';
import api, { extractApiError } from '../api';
import Tooltip from '../components/Tooltip';

type VerifyTab = 'cancellation' | 'bridge' | 'audit';

export default function VerificationPage() {
    const [tab, setTab] = useState<VerifyTab>('cancellation');
    const [statementID, setStatementID] = useState('');
    const [bridgeTxID, setBridgeTxID] = useState('');
    const [auditGoID, setAuditGoID] = useState('');
    const [result, setResult] = useState<Record<string, unknown> | null>(null);
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleVerifyCancellation = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setResult(null); setLoading(true);
        try {
            const { data } = await api.post('/cancellations/verify', { cancellationStatementID: statementID });
            setResult(data);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Verification failed'));
        } finally {
            setLoading(false);
        }
    };

    const handleVerifyBridge = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setResult(null); setLoading(true);
        try {
            const { data } = await api.post('/bridge/verify', { bridgeTransactionID: bridgeTxID });
            setResult(data);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Bridge verification failed'));
        } finally {
            setLoading(false);
        }
    };

    const handleAuditGO = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setResult(null); setLoading(true);
        try {
            const { data } = await api.get(`/guarantees/${auditGoID}/history`);
            setResult(data);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Audit lookup failed'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <div className="mb-6">
                <h2 className="text-2xl font-bold">Verification</h2>
                <p className="text-sm text-gray-500 mt-1">
                    Verify Cancellation Statements, cross-channel bridge proofs, and audit the full lifecycle of any Guarantee of Origin on the blockchain.
                </p>
            </div>

            {/* Tab selector */}
            <div className="flex gap-2 mb-6">
                <Tooltip text="Verify the authenticity of a Cancellation Statement" position="bottom">
                    <button onClick={() => { setTab('cancellation'); setResult(null); setError(''); }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            tab === 'cancellation' ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        Cancellation Statement
                    </button>
                </Tooltip>
                <Tooltip text="Verify a cross-channel bridge transfer proof, checking dual-issuer consensus" position="bottom">
                    <button onClick={() => { setTab('bridge'); setResult(null); setError(''); }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            tab === 'bridge' ? 'bg-indigo-100 text-indigo-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        Bridge Proof
                    </button>
                </Tooltip>
                <Tooltip text="View the complete lifecycle of a GO: issuance, transfers, conversions, and cancellation" position="bottom">
                    <button onClick={() => { setTab('audit'); setResult(null); setError(''); }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                            tab === 'audit' ? 'bg-teal-100 text-teal-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        GO Lifecycle Audit
                    </button>
                </Tooltip>
            </div>

            {/* Cancellation verification */}
            {tab === 'cancellation' && (
                <form onSubmit={handleVerifyCancellation} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Verify Cancellation Statement</h3>
                    <p className="text-sm text-gray-500">
                        Enter a Cancellation Statement ID to verify its authenticity and integrity on the blockchain.
                        This confirms the renewable energy attributes were properly claimed and retired.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Statement ID</label>
                        <input value={statementID} onChange={(e) => setStatementID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" placeholder="e.g. CS_eGO_1_1710000000" required />
                    </div>
                    {renderFeedback(error, null, loading)}
                    <button type="submit" disabled={loading}
                        className="bg-purple-600 hover:bg-purple-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Verifying...' : 'Verify Statement'}
                    </button>
                </form>
            )}

            {/* Bridge proof verification */}
            {tab === 'bridge' && (
                <form onSubmit={handleVerifyBridge} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Verify Cross-Channel Bridge Proof</h3>
                    <p className="text-sm text-gray-500">
                        Verify that a cross-channel bridge transfer was properly executed with dual-issuer consensus.
                        Both the source and target channel issuing authorities must have signed the proof.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Bridge Transaction ID</label>
                        <input value={bridgeTxID} onChange={(e) => setBridgeTxID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" placeholder="e.g. bridge_tx_abc123" required />
                    </div>
                    {renderFeedback(error, null, loading)}
                    <button type="submit" disabled={loading}
                        className="bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Verifying...' : 'Verify Bridge Proof'}
                    </button>
                </form>
            )}

            {/* GO lifecycle audit */}
            {tab === 'audit' && (
                <form onSubmit={handleAuditGO} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">GO Lifecycle Audit</h3>
                    <p className="text-sm text-gray-500">
                        Enter a GO Asset ID to view its complete lifecycle: issuance, transfers, conversions, and cancellation.
                        All events are immutably recorded on the Hyperledger Fabric ledger.
                    </p>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                        <input value={auditGoID} onChange={(e) => setAuditGoID(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" placeholder="e.g. eGO_1" required />
                    </div>
                    {renderFeedback(error, null, loading)}
                    <button type="submit" disabled={loading}
                        className="bg-teal-600 hover:bg-teal-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Loading...' : 'View Lifecycle'}
                    </button>
                </form>
            )}

            {/* Results display */}
            {result && (
                <div className="mt-6 bg-white rounded-xl shadow-sm border p-6 max-w-2xl">
                    <h3 className="font-semibold text-gray-700 mb-3">
                        {tab === 'cancellation' && '✓ Cancellation Statement Verified'}
                        {tab === 'bridge' && '✓ Bridge Proof Verified'}
                        {tab === 'audit' && 'GO Lifecycle History'}
                    </h3>
                    <pre className="text-xs text-gray-700 bg-gray-50 rounded-lg p-4 overflow-auto max-h-96">
                        {JSON.stringify(result, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
}

function renderFeedback(error: string, message: string | null, _loading: boolean) {
    return (
        <>
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
        </>
    );
}
