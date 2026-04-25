import { useState, useEffect, FormEvent } from 'react';
import api, { extractApiError } from '../api';
import { useAuth } from '../context/AuthContext';
import Tooltip from '../components/Tooltip';
import { Lock, Unlock, CheckCircle, ArrowRight } from 'lucide-react';

type BridgeTab = 'initiate' | 'locks' | 'approve';

interface Lock {
    lockId: string;
    goAssetId: string;
    goType: string;
    sourceChannel: string;
    destinationChannel: string;
    status: string;
    ownerMSP: string;
    sourceIssuerMSP: string;
    targetIssuerMSP?: string;
    lockReceiptHash: string;
    lockedAt: number;
    amountMWh?: number;
}

export default function BridgePage() {
    const { user } = useAuth();
    const [tab, setTab] = useState<BridgeTab>('initiate');
    const [locks, setLocks] = useState<Lock[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');

    // Initiate bridge form
    const [goAssetID, setGoAssetID] = useState('');
    const [destinationChannel, setDestinationChannel] = useState('');
    const [ownerMSP, setOwnerMSP] = useState(user?.mspId || '');

    // Finalize form
    const [selectedLockID, setSelectedLockID] = useState('');
    const [mintedAssetID, setMintedAssetID] = useState('');

    useEffect(() => {
        if (tab === 'locks') {
            loadLocks();
        }
    }, [tab]);

    const loadLocks = async () => {
        setLoading(true);
        setError('');
        try {
            const { data } = await api.get('/bridge/locks');
            setLocks(data.records || []);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to load locks'));
        } finally {
            setLoading(false);
        }
    };

    const handleInitiateLock = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');
        setLoading(true);
        try {
            const { data } = await api.post('/bridge/lock', {
                goAssetID,
                destinationChannel,
                ownerMSP,
            });
            setMessage(data.message || 'GO locked successfully — awaiting destination issuer approval');
            setGoAssetID('');
            setDestinationChannel('');
            if (tab === 'locks') loadLocks();
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to lock GO'));
        } finally {
            setLoading(false);
        }
    };

    const handleApprove = async (lockID: string) => {
        setError('');
        setMessage('');
        setLoading(true);
        try {
            const { data } = await api.post('/bridge/approve', { bridgeTransactionID: lockID });
            setMessage(data.message || 'Bridge transfer approved');
            loadLocks();
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to approve bridge transfer'));
        } finally {
            setLoading(false);
        }
    };

    const handleFinalize = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');
        setLoading(true);
        try {
            const { data } = await api.post('/bridge/finalize', {
                lockID: selectedLockID,
                mintedAssetID,
                ownerMSP,
            });
            setMessage(data.message || 'Bridge transfer finalized');
            setSelectedLockID('');
            setMintedAssetID('');
            if (tab === 'locks') loadLocks();
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to finalize bridge transfer'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <div className="mb-6">
                <h2 className="text-2xl font-bold">Cross-Channel Bridge</h2>
                <p className="text-sm text-gray-500 mt-1">
                    Transfer Guarantees of Origin across sovereign national registries with tri-party endorsement (owner + source issuer + destination issuer).
                </p>
            </div>

            {/* Tab selector */}
            <div className="flex gap-2 mb-6">
                <Tooltip text="Initiate a new cross-channel bridge transfer (requires owner consent)" position="bottom">
                    <button
                        onClick={() => { setTab('initiate'); setError(''); setMessage(''); }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 ${
                            tab === 'initiate' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}
                    >
                        <Lock size={16} />
                        Initiate Lock
                    </button>
                </Tooltip>
                <Tooltip text="View all pending and completed bridge locks" position="bottom">
                    <button
                        onClick={() => { setTab('locks'); setError(''); setMessage(''); }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 ${
                            tab === 'locks' ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}
                    >
                        <Unlock size={16} />
                        View Locks
                    </button>
                </Tooltip>
                {user?.role === 'issuer' && (
                    <Tooltip text="Approve pending bridge transfers (issuer only)" position="bottom">
                        <button
                            onClick={() => { setTab('approve'); setError(''); setMessage(''); }}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 ${
                                tab === 'approve' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                            }`}
                        >
                            <CheckCircle size={16} />
                            Approve
                        </button>
                    </Tooltip>
                )}
            </div>

            {/* Feedback messages */}
            {error && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
                    <p className="text-red-700 text-sm font-medium">Error</p>
                    <p className="text-red-600 text-sm mt-1">{error}</p>
                </div>
            )}
            {message && (
                <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                    <p className="text-green-700 text-sm">{message}</p>
                </div>
            )}

            {/* Initiate Lock Tab */}
            {tab === 'initiate' && (
                <div className="bg-white rounded-xl shadow-sm border p-6 max-w-2xl">
                    <h3 className="font-semibold text-gray-700 mb-4 flex items-center gap-2">
                        <Lock size={20} className="text-blue-600" />
                        Phase 1: Lock GO on Source Channel
                    </h3>
                    <p className="text-sm text-gray-500 mb-6">
                        Lock a Guarantee of Origin for cross-channel transfer. This requires <strong>tri-party endorsement</strong>:
                        the GO owner and source channel issuer must both sign this transaction. The GO status will transition to "locked"
                        and cannot be transferred or cancelled until the bridge is finalized or revoked.
                    </p>

                    <form onSubmit={handleInitiateLock} className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">GO Asset ID</label>
                            <input
                                value={goAssetID}
                                onChange={(e) => setGoAssetID(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2"
                                placeholder="e.g. eGO_1"
                                required
                            />
                            <p className="text-xs text-gray-500 mt-1">The GO you want to transfer across channels</p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Destination Channel</label>
                            <input
                                value={destinationChannel}
                                onChange={(e) => setDestinationChannel(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2"
                                placeholder="e.g. hydrogen-de"
                                required
                            />
                            <p className="text-xs text-gray-500 mt-1">The target channel where the GO will be minted</p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Owner MSP</label>
                            <input
                                value={ownerMSP}
                                onChange={(e) => setOwnerMSP(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2 bg-gray-50"
                                placeholder="e.g. eproducer1MSP"
                                required
                            />
                            <p className="text-xs text-gray-500 mt-1">
                                Your organization's MSP ID (defaults to your current identity). Both you and the source issuer must endorse.
                            </p>
                        </div>

                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                            <p className="text-sm text-blue-800 font-medium">⚠️ Tri-Party Endorsement Required</p>
                            <p className="text-xs text-blue-700 mt-1">
                                This transaction must be co-signed by both the GO owner ({ownerMSP}) and the source channel issuer.
                                The backend will automatically configure the endorsement policy to require both signatures.
                            </p>
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="bg-blue-600 hover:bg-blue-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50"
                        >
                            {loading ? 'Locking...' : 'Lock GO for Bridge Transfer'}
                        </button>
                    </form>
                </div>
            )}

            {/* View Locks Tab */}
            {tab === 'locks' && (
                <div className="bg-white rounded-xl shadow-sm border p-6">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="font-semibold text-gray-700 flex items-center gap-2">
                            <Unlock size={20} className="text-purple-600" />
                            Bridge Locks
                        </h3>
                        <button
                            onClick={loadLocks}
                            className="text-sm text-blue-600 hover:text-blue-700 font-medium"
                        >
                            ↻ Refresh
                        </button>
                    </div>

                    {loading && <p className="text-sm text-gray-500">Loading locks...</p>}

                    {!loading && locks.length === 0 && (
                        <p className="text-sm text-gray-500">No bridge locks found.</p>
                    )}

                    {!loading && locks.length > 0 && (
                        <div className="space-y-4">
                            {locks.map((lock) => (
                                <div key={lock.lockId} className="border rounded-lg p-4 hover:shadow-sm transition-shadow">
                                    <div className="flex items-start justify-between">
                                        <div className="flex-1">
                                            <div className="flex items-center gap-3 mb-2">
                                                <span className="text-sm font-mono text-gray-700">{lock.lockId}</span>
                                                <StatusBadge status={lock.status} />
                                            </div>
                                            <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                                                <div>
                                                    <span className="text-gray-500">GO Asset:</span>
                                                    <span className="ml-2 font-medium">{lock.goAssetId}</span>
                                                </div>
                                                <div>
                                                    <span className="text-gray-500">Type:</span>
                                                    <span className="ml-2 font-medium">{lock.goType}</span>
                                                </div>
                                                <div>
                                                    <span className="text-gray-500">Owner:</span>
                                                    <span className="ml-2 font-medium text-xs">{lock.ownerMSP}</span>
                                                </div>
                                                <div>
                                                    <span className="text-gray-500">Source Issuer:</span>
                                                    <span className="ml-2 font-medium text-xs">{lock.sourceIssuerMSP}</span>
                                                </div>
                                                <div className="col-span-2 flex items-center gap-2">
                                                    <span className="text-gray-500">{lock.sourceChannel}</span>
                                                    <ArrowRight size={14} className="text-gray-400" />
                                                    <span className="text-gray-500">{lock.destinationChannel}</span>
                                                </div>
                                            </div>
                                        </div>

                                        {user?.role === 'issuer' && lock.status === 'locked' && (
                                            <button
                                                onClick={() => handleApprove(lock.lockId)}
                                                disabled={loading}
                                                className="ml-4 bg-green-600 hover:bg-green-700 text-white rounded-lg px-4 py-2 text-sm font-medium disabled:opacity-50"
                                            >
                                                Approve
                                            </button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}

                    {/* Finalize Section */}
                    <div className="mt-8 pt-8 border-t">
                        <h4 className="font-semibold text-gray-700 mb-4 flex items-center gap-2">
                            <CheckCircle size={18} className="text-green-600" />
                            Phase 3: Finalize Lock (After Minting on Destination)
                        </h4>
                        <p className="text-sm text-gray-500 mb-4">
                            After the destination issuer has minted the GO on the target channel, finalize the lock
                            to complete the bridge transfer. This requires owner + source issuer endorsement.
                        </p>

                        <form onSubmit={handleFinalize} className="space-y-4 max-w-xl">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Lock ID</label>
                                <select
                                    value={selectedLockID}
                                    onChange={(e) => setSelectedLockID(e.target.value)}
                                    className="w-full border rounded-lg px-3 py-2"
                                    required
                                >
                                    <option value="">Select a lock to finalize</option>
                                    {locks.filter(l => l.status === 'approved').map(lock => (
                                        <option key={lock.lockId} value={lock.lockId}>
                                            {lock.lockId} — {lock.goAssetId}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Minted Asset ID</label>
                                <input
                                    value={mintedAssetID}
                                    onChange={(e) => setMintedAssetID(e.target.value)}
                                    className="w-full border rounded-lg px-3 py-2"
                                    placeholder="Asset ID from destination channel"
                                    required
                                />
                            </div>

                            <button
                                type="submit"
                                disabled={loading || !selectedLockID}
                                className="bg-green-600 hover:bg-green-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50"
                            >
                                {loading ? 'Finalizing...' : 'Finalize Bridge Transfer'}
                            </button>
                        </form>
                    </div>
                </div>
            )}

            {/* Approve Tab (Issuer Only) */}
            {tab === 'approve' && user?.role === 'issuer' && (
                <div className="bg-white rounded-xl shadow-sm border p-6">
                    <h3 className="font-semibold text-gray-700 mb-4 flex items-center gap-2">
                        <CheckCircle size={20} className="text-green-600" />
                        Approve Pending Bridge Transfers
                    </h3>
                    <p className="text-sm text-gray-500 mb-6">
                        As a target channel issuer, approve pending bridge locks to allow minting on this channel.
                        This is part of the dual-issuer consensus mechanism.
                    </p>

                    {loading && <p className="text-sm text-gray-500">Loading pending locks...</p>}

                    {!loading && locks.filter(l => l.status === 'locked').length === 0 && (
                        <p className="text-sm text-gray-500">No pending bridge locks require approval.</p>
                    )}

                    {!loading && locks.filter(l => l.status === 'locked').length > 0 && (
                        <div className="space-y-4">
                            {locks.filter(l => l.status === 'locked').map((lock) => (
                                <div key={lock.lockId} className="border rounded-lg p-4 flex items-center justify-between">
                                    <div>
                                        <p className="text-sm font-medium text-gray-700">{lock.goAssetId}</p>
                                        <p className="text-xs text-gray-500 mt-1">
                                            {lock.sourceChannel} → {lock.destinationChannel}
                                        </p>
                                        <p className="text-xs text-gray-500">Owner: {lock.ownerMSP}</p>
                                    </div>
                                    <button
                                        onClick={() => handleApprove(lock.lockId)}
                                        disabled={loading}
                                        className="bg-green-600 hover:bg-green-700 text-white rounded-lg px-4 py-2 text-sm font-medium disabled:opacity-50"
                                    >
                                        Approve
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

function StatusBadge({ status }: { status: string }) {
    const colors = {
        locked: 'bg-yellow-100 text-yellow-800',
        approved: 'bg-blue-100 text-blue-800',
        bridged: 'bg-green-100 text-green-800',
        expired: 'bg-gray-100 text-gray-800',
    };
    return (
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800'}`}>
            {status}
        </span>
    );
}
