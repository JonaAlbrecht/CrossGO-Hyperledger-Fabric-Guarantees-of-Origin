import { useEffect, useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import api, { extractApiError } from '../api';
import { ENERGY_CARRIERS, carrierStyle } from '../types';
import type { ElectricityGO, HydrogenGO, EnergyCarrier } from '../types';
import Tooltip from '../components/Tooltip';

const PRODUCTION_METHODS: Record<EnergyCarrier, { value: string; label: string }[]> = {
    electricity: [
        { value: 'solar', label: 'Solar' },
        { value: 'wind', label: 'Wind' },
        { value: 'hydro', label: 'Hydro' },
        { value: 'biomass', label: 'Biomass' },
        { value: 'geothermal', label: 'Geothermal' },
    ],
    hydrogen: [
        { value: 'electrolysis', label: 'Electrolysis' },
        { value: 'smr', label: 'Steam Methane Reforming' },
        { value: 'biomass_gasification', label: 'Biomass Gasification' },
    ],
    biogas: [
        { value: 'anaerobic_digestion', label: 'Anaerobic Digestion' },
        { value: 'landfill_gas', label: 'Landfill Gas' },
        { value: 'sewage_gas', label: 'Sewage Gas' },
    ],
    heating_cooling: [
        { value: 'heat_pump', label: 'Heat Pump' },
        { value: 'geothermal_heating', label: 'Geothermal Heating' },
        { value: 'district_heating', label: 'District Heating' },
        { value: 'solar_thermal', label: 'Solar Thermal' },
        { value: 'absorption_cooling', label: 'Absorption Cooling' },
    ],
};

const AMOUNT_UNITS: Record<EnergyCarrier, string> = {
    electricity: 'MWh',
    hydrogen: 'kg',
    biogas: 'Nm³',
    heating_cooling: 'MWh',
};

interface GORecord {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export default function GuaranteesPage() {
    const { user } = useAuth();
    const [gosByCarrier, setGosByCarrier] = useState<Record<string, GORecord[]>>({});
    const [loading, setLoading] = useState(true);
    const [tab, setTab] = useState<EnergyCarrier>('electricity');
    const [showForm, setShowForm] = useState(false);
    const [error, setError] = useState('');
    const [formCarrier, setFormCarrier] = useState<EnergyCarrier>('electricity');

    // Form fields
    const [amount, setAmount] = useState('');
    const [emissions, setEmissions] = useState('');
    const [elapsed, setElapsed] = useState('');
    const [method, setMethod] = useState('solar');

    const fetchGOs = async () => {
        try {
            const { data } = await api.get('/guarantees');
            setGosByCarrier({
                electricity: data.electricityGOs ?? [],
                hydrogen: data.hydrogenGOs ?? [],
                biogas: data.biogasGOs ?? [],
                heating_cooling: data.heatingCoolingGOs ?? [],
            });
        } catch {
            setGosByCarrier({});
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchGOs(); }, []);

    const handleCreateGO = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        try {
            // Map the URL path for heating_cooling to the backend route
            const routeCarrier = formCarrier === 'heating_cooling' ? 'heating-cooling' : formCarrier;
            await api.post(`/guarantees/${routeCarrier}`, {
                amount: parseFloat(amount),
                emissions: parseFloat(emissions),
                elapsedSeconds: parseFloat(elapsed),
                productionMethod: method,
                energyCarrier: formCarrier,
            });
            setShowForm(false);
            await fetchGOs();
        } catch (err: unknown) {
            setError(extractApiError(err, 'Failed to create Guarantee of Origin'));
        }
    };

    const formatDate = (ts: number) => new Date(ts * 1000).toLocaleString();
    const currentGOs = gosByCarrier[tab] ?? [];

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h2 className="text-2xl font-bold">Guarantees of Origin</h2>
                    <p className="text-sm text-gray-500 mt-1">
                        View and create Guarantees of Origin for all energy carriers
                    </p>
                </div>
                {user?.role === 'producer' && (
                    <Tooltip text="Issue a new Guarantee of Origin based on metered production data from a registered device" position="left">
                        <button onClick={() => setShowForm(!showForm)}
                            className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors">
                            {showForm ? 'Cancel' : '+ Create Guarantee of Origin'}
                        </button>
                    </Tooltip>
                )}
            </div>

            {showForm && (
                <form onSubmit={handleCreateGO} className="bg-white rounded-xl shadow-sm border p-6 mb-6 space-y-4">
                    <h3 className="font-semibold text-gray-700">Create Guarantee of Origin</h3>
                    <p className="text-sm text-gray-500">
                        A Guarantee of Origin (GO) is an electronic document that certifies the origin of energy produced from renewable sources,
                        as defined by EU Directive 2018/2001 (RED II).
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Energy Carrier</label>
                            <select value={formCarrier} onChange={(e) => {
                                const c = e.target.value as EnergyCarrier;
                                setFormCarrier(c);
                                setMethod(PRODUCTION_METHODS[c][0].value);
                            }}
                                className="w-full border rounded-lg px-3 py-2">
                                {ENERGY_CARRIERS.map((c) => (
                                    <option key={c.value} value={c.value}>{c.label}</option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Amount ({AMOUNT_UNITS[formCarrier]})
                            </label>
                            <input type="number" step="0.001" value={amount} onChange={(e) => setAmount(e.target.value)}
                                className="w-full border rounded-lg px-3 py-2" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Emissions (kg CO₂)</label>
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
                                {PRODUCTION_METHODS[formCarrier].map((m) => (
                                    <option key={m.value} value={m.value}>{m.label}</option>
                                ))}
                            </select>
                        </div>
                    </div>
                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                            <p className="text-red-700 text-sm font-medium">Failed to create Guarantee of Origin</p>
                            <p className="text-red-600 text-sm mt-1">{error}</p>
                        </div>
                    )}
                    <button type="submit"
                        className="bg-primary-600 hover:bg-primary-700 text-white rounded-lg px-6 py-2 text-sm font-medium">
                        Submit to Blockchain
                    </button>
                </form>
            )}

            {/* Carrier tabs */}
            <div className="flex gap-2 mb-4 flex-wrap">
                {ENERGY_CARRIERS.map((c) => {
                    const count = (gosByCarrier[c.value] ?? []).length;
                    return (
                        <button key={c.value} onClick={() => setTab(c.value)}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                                tab === c.value ? `${c.bgColor} ${c.color}` : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                            }`}>
                            {c.label} ({count})
                        </button>
                    );
                })}
            </div>

            {loading ? (
                <p className="text-gray-400">Loading Guarantees of Origin...</p>
            ) : (
                <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-600">
                            <tr>
                                <th className="text-left px-6 py-3">Asset ID</th>
                                <th className="text-left px-6 py-3">Energy Carrier</th>
                                <th className="text-left px-6 py-3">Created</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y">
                            {currentGOs.map((go) => {
                                const style = carrierStyle(go.GOType || tab);
                                return (
                                    <tr key={go.AssetID} className="hover:bg-gray-50">
                                        <td className="px-6 py-3 font-mono text-xs">{go.AssetID}</td>
                                        <td className="px-6 py-3">
                                            <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${style.bgColor} ${style.color}`}>
                                                {style.label}
                                            </span>
                                        </td>
                                        <td className="px-6 py-3 text-gray-500">{formatDate(go.CreationDateTime)}</td>
                                    </tr>
                                );
                            })}
                            {currentGOs.length === 0 && (
                                <tr><td colSpan={3} className="px-6 py-8 text-center text-gray-400">
                                    No {carrierStyle(tab).label} Guarantees of Origin found
                                </td></tr>
                            )}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
