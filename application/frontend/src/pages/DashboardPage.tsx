import { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { orgDisplayName } from '../types';
import Tooltip from '../components/Tooltip';
import api from '../api';
import { Zap, Droplets, Flame, Thermometer, Cpu, FileCheck, TrendingUp, BarChart3, Activity } from 'lucide-react';

interface DashboardStats {
    electricityGOs: number;
    hydrogenGOs: number;
    biogasGOs: number;
    heatingCoolingGOs: number;
    devices: number;
    totalGOs: number;
}

function StatCard({ icon: Icon, label, value, color, tooltip }: {
    icon: typeof Zap; label: string; value: number; color: string; tooltip: string;
}) {
    return (
        <Tooltip text={tooltip} position="bottom">
            <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 w-full">
                <div className="flex items-center gap-4">
                    <div className={`p-3 rounded-lg ${color}`}>
                        <Icon size={24} className="text-white" />
                    </div>
                    <div>
                        <p className="text-2xl font-bold">{value}</p>
                        <p className="text-sm text-gray-500">{label}</p>
                    </div>
                </div>
            </div>
        </Tooltip>
    );
}

export default function DashboardPage() {
    const { user } = useAuth();
    const [stats, setStats] = useState<DashboardStats>({
        electricityGOs: 0, hydrogenGOs: 0, biogasGOs: 0, heatingCoolingGOs: 0, devices: 0, totalGOs: 0,
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchStats() {
            try {
                const [goRes, devRes] = await Promise.all([
                    api.get('/guarantees'),
                    api.get('/devices'),
                ]);
                const eCount = goRes.data.electricityGOs?.length ?? 0;
                const hCount = goRes.data.hydrogenGOs?.length ?? 0;
                const bCount = goRes.data.biogasGOs?.length ?? 0;
                const hcCount = goRes.data.heatingCoolingGOs?.length ?? 0;
                setStats({
                    electricityGOs: eCount,
                    hydrogenGOs: hCount,
                    biogasGOs: bCount,
                    heatingCoolingGOs: hcCount,
                    devices: Array.isArray(devRes.data) ? devRes.data.length : 0,
                    totalGOs: eCount + hCount + bCount + hcCount,
                });
            } catch {
                // Stats not available — show zeros
            } finally {
                setLoading(false);
            }
        }
        fetchStats();
    }, []);

    return (
        <div>
            <h2 className="text-2xl font-bold mb-1">Dashboard</h2>
            <p className="text-gray-500 mb-8">
                Welcome, <span className="font-medium text-gray-700">{user?.userName}</span> —{' '}
                <span className="capitalize">{user?.role}</span> at {orgDisplayName(user?.orgName ?? '')}
            </p>

            {loading ? (
                <p className="text-gray-400">Loading statistics...</p>
            ) : (
                <>
                    {/* Summary row */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                        <StatCard icon={TrendingUp} label="Total Guarantees of Origin" value={stats.totalGOs}
                            color="bg-primary-600" tooltip="Total number of active Guarantees of Origin across all energy carriers" />
                        <StatCard icon={Cpu} label="Registered Devices" value={stats.devices}
                            color="bg-gray-600" tooltip="Metering devices registered on the network that attest energy production" />
                        <StatCard icon={BarChart3} label="Energy Carriers" value={4}
                            color="bg-indigo-500" tooltip="Supported energy carriers: electricity, hydrogen, biogas, heating & cooling" />
                    </div>

                    {/* Per-carrier row */}
                    <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
                        <Activity size={18} />
                        Guarantees of Origin by Energy Carrier
                    </h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
                        <StatCard icon={Zap} label="Electricity GOs" value={stats.electricityGOs}
                            color="bg-yellow-500" tooltip="Guarantees of Origin from electricity production (solar, wind, hydro, biomass, geothermal)" />
                        <StatCard icon={Droplets} label="Hydrogen GOs" value={stats.hydrogenGOs}
                            color="bg-blue-500" tooltip="Guarantees of Origin from hydrogen production (electrolysis, SMR, biomass gasification)" />
                        <StatCard icon={Flame} label="Biogas GOs" value={stats.biogasGOs}
                            color="bg-green-500" tooltip="Guarantees of Origin from biogas production (anaerobic digestion, landfill gas)" />
                        <StatCard icon={Thermometer} label="Heating & Cooling GOs" value={stats.heatingCoolingGOs}
                            color="bg-orange-500" tooltip="Guarantees of Origin from district heating/cooling, heat pumps, geothermal heating" />
                    </div>
                </>
            )}

            <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <FileCheck size={20} />
                    Quick Actions
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                    {user?.role === 'issuer' && (
                        <>
                            <ActionButton href="/devices" label="Register Device"
                                tooltip="Register a new metering device for a producer organization" />
                            <ActionButton href="/organizations" label="Register Organization"
                                tooltip="Onboard a new producer or buyer organization to the network" />
                        </>
                    )}
                    {user?.role === 'producer' && (
                        <>
                            <ActionButton href="/guarantees" label="Create Guarantee of Origin"
                                tooltip="Issue a new Guarantee of Origin based on metered production data" />
                            <ActionButton href="/conversions" label="Convert Energy Carrier"
                                tooltip="Convert Guarantees of Origin from one energy carrier to another" />
                        </>
                    )}
                    {(user?.role === 'producer' || user?.role === 'consumer') && (
                        <ActionButton href="/transfers" label="Transfer GO"
                            tooltip="Transfer Guarantees of Origin to another organization" />
                    )}
                    <ActionButton href="/cancellations" label="Cancellation Statements"
                        tooltip="Cancel Guarantees of Origin to claim renewable attributes" />
                    <ActionButton href="/verification" label="Verify on Blockchain"
                        tooltip="Verify Cancellation Statements and audit GO lifecycle" />
                </div>
            </div>
        </div>
    );
}

function ActionButton({ href, label, tooltip }: { href: string; label: string; tooltip: string }) {
    return (
        <Tooltip text={tooltip} position="bottom">
            <a
                href={href}
                className="block text-center bg-gray-50 hover:bg-primary-50 border border-gray-200 hover:border-primary-300 rounded-lg px-4 py-3 text-sm font-medium text-gray-700 hover:text-primary-700 transition-colors w-full"
            >
                {label}
            </a>
        </Tooltip>
    );
}
