import { useState, FormEvent } from 'react';
import api, { extractApiError } from '../api';
import { ENERGY_CARRIERS } from '../types';
import type { EnergyCarrier } from '../types';
import Tooltip from '../components/Tooltip';

/** Per-carrier production methods — what conversion does this carrier involve? */
const CONVERSION_METHODS: Record<EnergyCarrier, { value: string; label: string }[]> = {
    electricity: [
        { value: 'wind', label: 'Wind Turbine' },
        { value: 'solar', label: 'Solar PV' },
        { value: 'hydro', label: 'Hydroelectric' },
    ],
    hydrogen: [
        { value: 'electrolysis', label: 'Electrolysis' },
        { value: 'smr', label: 'Steam Methane Reforming' },
        { value: 'biomass_gasification', label: 'Biomass Gasification' },
    ],
    biogas: [
        { value: 'anaerobic_digestion', label: 'Anaerobic Digestion' },
        { value: 'landfill_capture', label: 'Landfill Gas Capture' },
    ],
    heating_cooling: [
        { value: 'heat_pump', label: 'Heat Pump' },
        { value: 'solar_thermal', label: 'Solar Thermal' },
        { value: 'geothermal', label: 'Geothermal' },
        { value: 'district_heating', label: 'District Heating' },
    ],
};

/** Per-carrier amount units and API field names */
const CARRIER_UNITS: Record<EnergyCarrier, { unit: string; apiField: string }> = {
    electricity: { unit: 'MWh', apiField: 'amountMWh' },
    hydrogen: { unit: 'kg', apiField: 'kilosHydrogen' },
    biogas: { unit: 'm³', apiField: 'cubicMetersBiogas' },
    heating_cooling: { unit: 'MWh', apiField: 'amountMWhThermal' },
};

export default function ConversionsPage() {
    const [tab, setTab] = useState<'backlog' | 'issue'>('backlog');
    const [targetCarrier, setTargetCarrier] = useState<EnergyCarrier>('hydrogen');
    const [sourceCarrier, setSourceCarrier] = useState<EnergyCarrier>('electricity');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    // Backlog form
    const [outputAmount, setOutputAmount] = useState('');
    const [method, setMethod] = useState(CONVERSION_METHODS.hydrogen[0].value);
    const [inputAmount, setInputAmount] = useState('');

    const handleCarrierChange = (c: EnergyCarrier) => {
        setTargetCarrier(c);
        setMethod(CONVERSION_METHODS[c][0]?.value ?? '');
        // Auto-set a sensible source carrier
        if (c === 'hydrogen') setSourceCarrier('electricity');
        else if (c === 'biogas') setSourceCarrier('electricity');
        else if (c === 'heating_cooling') setSourceCarrier('electricity');
        else setSourceCarrier('hydrogen');
    };

    const handleAddBacklog = async (e: FormEvent) => {
        e.preventDefault();
        setError(''); setMessage(''); setLoading(true);
        try {
            const targetUnit = CARRIER_UNITS[targetCarrier];
            const sourceUnit = CARRIER_UNITS[sourceCarrier];
            const { data } = await api.post('/conversions/backlog', {
                targetCarrier,
                sourceCarrier,
                [targetUnit.apiField]: parseFloat(outputAmount),
                productionMethod: method,
                [`sourceAmount_${sourceUnit.apiField}`]: parseFloat(inputAmount),
            });
            setMessage(data.message);
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to add to backlog'));
        } finally {
            setLoading(false);
        }
    };

    const handleIssue = async () => {
        setError(''); setMessage(''); setLoading(true);
        try {
            const { data } = await api.post('/conversions/issue', { targetCarrier });
            setMessage(data.message);
        } catch (err: unknown) {
            setError(extractApiError(err, `Failed to issue ${targetCarrier} GOs`));
        } finally {
            setLoading(false);
        }
    };

    const targetInfo = CARRIER_UNITS[targetCarrier];
    const sourceInfo = CARRIER_UNITS[sourceCarrier];
    const targetLabel = ENERGY_CARRIERS.find((c) => c.value === targetCarrier)!.label;
    const sourceLabel = ENERGY_CARRIERS.find((c) => c.value === sourceCarrier)!.label;

    return (
        <div>
            <div className="mb-6">
                <h2 className="text-2xl font-bold">Energy Carrier Conversion</h2>
                <p className="text-sm text-gray-500 mt-1">
                    Record production of one energy carrier from another and issue new Guarantees of Origin.
                    Source GOs are consumed (cancelled) to cover the conversion.
                </p>
            </div>

            {/* Target carrier selector */}
            <div className="mb-4">
                <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">Target Carrier</label>
                <div className="flex gap-2">
                    {ENERGY_CARRIERS.map((c) => (
                        <button key={c.value} onClick={() => handleCarrierChange(c.value)}
                            className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                                targetCarrier === c.value
                                    ? `${c.bgColor} ${c.color} ring-2 ring-offset-1 ring-gray-300`
                                    : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                            }`}>
                            {c.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Tab selector */}
            <div className="flex gap-2 mb-6">
                <Tooltip text="Record production output and match it against source carrier GOs" position="bottom">
                    <button onClick={() => setTab('backlog')}
                        className={`px-4 py-2 rounded-lg text-sm font-medium ${
                            tab === 'backlog' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        Add to Backlog
                    </button>
                </Tooltip>
                <Tooltip text="Consume source GOs and mint new target carrier GO" position="bottom">
                    <button onClick={() => setTab('issue')}
                        className={`px-4 py-2 rounded-lg text-sm font-medium ${
                            tab === 'issue' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}>
                        Issue {targetLabel} GO
                    </button>
                </Tooltip>
            </div>

            {tab === 'backlog' ? (
                <form onSubmit={handleAddBacklog} className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Add {targetLabel} Production to Backlog</h3>
                    <p className="text-sm text-gray-500">
                        Record {targetLabel.toLowerCase()} production that consumed {sourceLabel.toLowerCase()}.
                        The platform will match this against your {sourceLabel.toLowerCase()} GOs to issue {targetLabel.toLowerCase()} GOs.
                    </p>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Source Carrier</label>
                        <select value={sourceCarrier} onChange={(e) => setSourceCarrier(e.target.value as EnergyCarrier)}
                            className="w-full border rounded-lg px-3 py-2">
                            {ENERGY_CARRIERS.filter((c) => c.value !== targetCarrier).map((c) => (
                                <option key={c.value} value={c.value}>{c.label}</option>
                            ))}
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {targetLabel} Produced ({targetInfo.unit})
                        </label>
                        <input type="number" step="0.001" value={outputAmount} onChange={(e) => setOutputAmount(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Production Method</label>
                        <select value={method} onChange={(e) => setMethod(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2">
                            {CONVERSION_METHODS[targetCarrier].map((m) => (
                                <option key={m.value} value={m.value}>{m.label}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {sourceLabel} Used ({sourceInfo.unit})
                        </label>
                        <input type="number" step="0.001" value={inputAmount} onChange={(e) => setInputAmount(e.target.value)}
                            className="w-full border rounded-lg px-3 py-2" required />
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
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Submitting...' : 'Add to Backlog'}
                    </button>
                </form>
            ) : (
                <div className="bg-white rounded-xl shadow-sm border p-6 max-w-lg space-y-4">
                    <h3 className="font-semibold text-gray-700">Issue {targetLabel} GO from Backlog</h3>
                    <p className="text-sm text-gray-500">
                        This will consume source carrier GOs from your collection to cover the {targetLabel.toLowerCase()} backlog,
                        then mint a new {targetLabel.toLowerCase()} Guarantee of Origin.
                        Consumption declarations are created automatically.
                    </p>
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
                    <button onClick={handleIssue} disabled={loading}
                        className="bg-blue-600 hover:bg-blue-700 text-white rounded-lg px-6 py-2 text-sm font-medium disabled:opacity-50">
                        {loading ? 'Processing...' : `Issue ${targetLabel} GO`}
                    </button>
                </div>
            )}
        </div>
    );
}
