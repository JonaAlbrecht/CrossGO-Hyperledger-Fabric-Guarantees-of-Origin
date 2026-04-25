import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { orgDisplayName } from '../types';
import Tooltip from './Tooltip';
import {
    Zap, ArrowRightLeft, FlaskConical, FileX2, Cpu, LayoutDashboard, LogOut,
    ShieldCheck, Building2, ArrowLeftRight,
} from 'lucide-react';

const NAV_SECTIONS = [
    {
        heading: null,
        items: [
            {
                to: '/', label: 'Dashboard', icon: LayoutDashboard,
                roles: ['issuer', 'producer', 'consumer'],
                tooltip: 'Overview of your Guarantees of Origin, devices and key metrics',
            },
        ],
    },
    {
        heading: 'Management',
        items: [
            {
                to: '/devices', label: 'Devices', icon: Cpu,
                roles: ['issuer', 'producer'],
                tooltip: 'Register and manage metering devices that attest energy production',
            },
            {
                to: '/organizations', label: 'Organizations', icon: Building2,
                roles: ['issuer'],
                tooltip: 'Register new producer or buyer organizations on the network',
            },
            {
                to: '/guarantees', label: 'Guarantees of Origin', icon: Zap,
                roles: ['issuer', 'producer', 'consumer'],
                tooltip: 'View, create and manage Guarantees of Origin for all energy carriers',
            },
        ],
    },
    {
        heading: 'Operations',
        items: [
            {
                to: '/transfers', label: 'Transfers', icon: ArrowRightLeft,
                roles: ['producer', 'consumer'],
                tooltip: 'Transfer Guarantees of Origin between organizations',
            },
            {
                to: '/conversions', label: 'Conversions', icon: FlaskConical,
                roles: ['producer'],
                tooltip: 'Convert Guarantees of Origin from one energy carrier to another (e.g. electricity → hydrogen)',
            },
            {
                to: '/cancellations', label: 'Cancellation Statements', icon: FileX2,
                roles: ['issuer', 'producer', 'consumer'],
                tooltip: 'Cancel Guarantees of Origin to claim renewable attributes and generate Cancellation Statements',
            },
            {
                to: '/bridge', label: 'Cross-Channel Bridge', icon: ArrowLeftRight,
                roles: ['issuer', 'producer', 'consumer'],
                tooltip: 'Transfer GOs across sovereign national registries with tri-party endorsement (owner + source issuer + dest issuer)',
            },
        ],
    },
    {
        heading: 'Verification',
        items: [
            {
                to: '/verification', label: 'Verify', icon: ShieldCheck,
                roles: ['issuer', 'producer', 'consumer'],
                tooltip: 'Verify Cancellation Statements, cross-channel bridge proofs, and audit GO lifecycle on the blockchain',
            },
        ],
    },
];

export default function Layout() {
    const { user, logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="flex h-screen">
            {/* Sidebar */}
            <aside className="w-64 bg-primary-800 text-white flex flex-col">
                <div className="p-6 border-b border-primary-700">
                    <h1 className="text-xl font-bold">GO Platform</h1>
                    <p className="text-primary-100 text-sm mt-1">Guarantees of Origin</p>
                </div>

                <nav className="flex-1 py-2 overflow-y-auto">
                    {NAV_SECTIONS.map((section, si) => {
                        const visibleItems = section.items.filter((item) =>
                            item.roles.includes(user?.role ?? ''),
                        );
                        if (visibleItems.length === 0) return null;
                        return (
                            <div key={si}>
                                {section.heading && (
                                    <p className="px-6 pt-4 pb-1 text-xs font-semibold uppercase tracking-wider text-primary-400">
                                        {section.heading}
                                    </p>
                                )}
                                {visibleItems.map(({ to, label, icon: Icon, tooltip }) => (
                                    <Tooltip key={to} text={tooltip} position="right">
                                        <NavLink
                                            to={to}
                                            end={to === '/'}
                                            className={({ isActive }) =>
                                                `flex items-center gap-3 px-6 py-2.5 text-sm transition-colors w-full ${
                                                    isActive
                                                        ? 'bg-primary-700 text-white font-medium'
                                                        : 'text-primary-100 hover:bg-primary-700/50'
                                                }`
                                            }
                                        >
                                            <Icon size={18} />
                                            {label}
                                        </NavLink>
                                    </Tooltip>
                                ))}
                            </div>
                        );
                    })}
                </nav>

                <div className="p-4 border-t border-primary-700">
                    <div className="text-sm text-primary-100 mb-2">
                        <span className="font-medium text-white">{user?.userName}</span>
                        <br />
                        <span className="capitalize">{user?.role}</span> · {orgDisplayName(user?.orgName ?? '')}
                    </div>
                    <button
                        onClick={handleLogout}
                        className="flex items-center gap-2 text-sm text-primary-200 hover:text-white transition-colors"
                    >
                        <LogOut size={16} />
                        Sign out
                    </button>
                </div>
            </aside>

            {/* Main content */}
            <main className="flex-1 overflow-auto p-8">
                <Outlet />
            </main>
        </div>
    );
}
