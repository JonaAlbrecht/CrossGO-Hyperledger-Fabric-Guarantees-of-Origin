import { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../api';
import { Zap, Droplets, Cpu, FileCheck } from 'lucide-react';

interface DashboardStats {
    electricityGOs: number;
    hydrogenGOs: number;
    devices: number;
}

function StatCard({ icon: Icon, label, value, color }: { icon: typeof Zap; label: string; value: number; color: string }) {
    return (
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
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
    );
}

export default function DashboardPage() {
    const { user } = useAuth();
    const [stats, setStats] = useState<DashboardStats>({ electricityGOs: 0, hydrogenGOs: 0, devices: 0 });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchStats() {
            try {
                const [goRes, devRes] = await Promise.all([
                    api.get('/guarantees'),
                    api.get('/devices'),
                ]);
                setStats({
                    electricityGOs: goRes.data.electricityGOs?.length ?? 0,
                    hydrogenGOs: goRes.data.hydrogenGOs?.length ?? 0,
                    devices: Array.isArray(devRes.data) ? devRes.data.length : 0,
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
                <span className="capitalize">{user?.role}</span> at {user?.orgName}
            </p>

            {loading ? (
                <p className="text-gray-400">Loading statistics...</p>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <StatCard icon={Zap} label="Electricity GOs" value={stats.electricityGOs} color="bg-yellow-500" />
                    <StatCard icon={Droplets} label="Hydrogen GOs" value={stats.hydrogenGOs} color="bg-blue-500" />
                    <StatCard icon={Cpu} label="Devices" value={stats.devices} color="bg-primary-600" />
                </div>
            )}

            <div className="mt-10 bg-white rounded-xl shadow-sm border border-gray-100 p-6">
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <FileCheck size={20} />
                    Quick Actions
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                    {user?.role === 'issuer' && (
                        <ActionButton href="/devices" label="Register Device" />
                    )}
                    {(user?.role === 'producer') && (
                        <>
                            <ActionButton href="/guarantees" label="Create GO" />
                            <ActionButton href="/conversions" label="Manage Backlog" />
                        </>
                    )}
                    {(user?.role === 'producer' || user?.role === 'consumer') && (
                        <ActionButton href="/transfers" label="Transfer GO" />
                    )}
                    <ActionButton href="/certificates" label="View Certificates" />
                </div>
            </div>
        </div>
    );
}

function ActionButton({ href, label }: { href: string; label: string }) {
    return (
        <a
            href={href}
            className="block text-center bg-gray-50 hover:bg-primary-50 border border-gray-200 hover:border-primary-300 rounded-lg px-4 py-3 text-sm font-medium text-gray-700 hover:text-primary-700 transition-colors"
        >
            {label}
        </a>
    );
}
