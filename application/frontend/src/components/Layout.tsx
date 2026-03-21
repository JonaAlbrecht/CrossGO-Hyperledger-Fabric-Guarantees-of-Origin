import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Zap, ArrowRightLeft, FlaskConical, FileCheck, Cpu, LayoutDashboard, LogOut } from 'lucide-react';

const NAV_ITEMS = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard, roles: ['issuer', 'producer', 'consumer'] },
    { to: '/devices', label: 'Devices', icon: Cpu, roles: ['issuer', 'producer'] },
    { to: '/guarantees', label: 'Guarantees', icon: Zap, roles: ['issuer', 'producer', 'consumer'] },
    { to: '/transfers', label: 'Transfers', icon: ArrowRightLeft, roles: ['producer', 'consumer'] },
    { to: '/conversions', label: 'Conversions', icon: FlaskConical, roles: ['producer'] },
    { to: '/certificates', label: 'Certificates', icon: FileCheck, roles: ['issuer', 'producer', 'consumer'] },
];

export default function Layout() {
    const { user, logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const visibleNav = NAV_ITEMS.filter((item) => item.roles.includes(user?.role ?? ''));

    return (
        <div className="flex h-screen">
            {/* Sidebar */}
            <aside className="w-64 bg-primary-800 text-white flex flex-col">
                <div className="p-6 border-b border-primary-700">
                    <h1 className="text-xl font-bold">GO Platform</h1>
                    <p className="text-primary-100 text-sm mt-1">Guarantee of Origin</p>
                </div>

                <nav className="flex-1 py-4">
                    {visibleNav.map(({ to, label, icon: Icon }) => (
                        <NavLink
                            key={to}
                            to={to}
                            end={to === '/'}
                            className={({ isActive }) =>
                                `flex items-center gap-3 px-6 py-3 text-sm transition-colors ${
                                    isActive
                                        ? 'bg-primary-700 text-white font-medium'
                                        : 'text-primary-100 hover:bg-primary-700/50'
                                }`
                            }
                        >
                            <Icon size={18} />
                            {label}
                        </NavLink>
                    ))}
                </nav>

                <div className="p-4 border-t border-primary-700">
                    <div className="text-sm text-primary-100 mb-2">
                        <span className="font-medium text-white">{user?.userName}</span>
                        <br />
                        <span className="capitalize">{user?.role}</span> · {user?.orgName}
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
