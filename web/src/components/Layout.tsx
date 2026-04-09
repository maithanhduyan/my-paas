import { useEffect, useState } from 'react'
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { Box, LayoutDashboard, Plus, Database, LogOut, Archive, Users, Package, Server, Menu, X, PanelLeftClose, PanelLeftOpen } from 'lucide-react'

export function Layout() {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem('sidebar_collapsed') === 'true')

  const toggleCollapsed = () => {
    setCollapsed((prev) => {
      localStorage.setItem('sidebar_collapsed', String(!prev))
      return !prev
    })
  }

  // Close sidebar on route change (mobile)
  useEffect(() => {
    setSidebarOpen(false)
  }, [pathname])

  useEffect(() => {
    const token = localStorage.getItem('mypaas_token')
    if (!token) {
      navigate('/login')
    }
  }, [navigate])

  const handleLogout = () => {
    const token = localStorage.getItem('mypaas_token')
    if (token) {
      fetch('/api/auth/logout', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {})
    }
    localStorage.removeItem('mypaas_token')
    navigate('/login')
  }

  const sidebarWidth = collapsed ? 'w-[60px]' : 'w-56'

  return (
    <div className="flex h-screen">
      {/* Mobile header */}
      <div className="fixed top-0 left-0 right-0 h-14 bg-surface-50 border-b border-surface-300 flex items-center px-4 z-30 lg:hidden">
        <button onClick={() => setSidebarOpen(!sidebarOpen)} className="p-1 text-gray-400 hover:text-gray-200">
          {sidebarOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
        <Link to="/" className="flex items-center gap-2 ml-3 text-lg font-bold">
          <Box className="w-5 h-5 text-accent" />
          <span>My PaaS</span>
        </Link>
      </div>

      {/* Overlay */}
      {sidebarOpen && (
        <div className="fixed inset-0 bg-black/60 z-30 lg:hidden" onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed top-0 left-0 bottom-0 ${sidebarWidth} bg-surface-50 border-r border-surface-300 flex flex-col z-40
        transition-all duration-200 ease-in-out
        lg:static lg:translate-x-0
        ${sidebarOpen ? 'translate-x-0 !w-56' : '-translate-x-full'}
      `}>
        {/* Logo */}
        <div className="p-4 border-b border-surface-300 flex items-center justify-between">
          <Link to="/" className={`flex items-center gap-2 text-lg font-bold ${collapsed ? 'justify-center w-full' : ''}`}>
            <Box className="w-6 h-6 text-accent shrink-0" />
            {!collapsed && <span>My PaaS</span>}
          </Link>
        </div>

        {/* Nav */}
        <nav className="flex-1 p-2 space-y-1 overflow-y-auto">
          <NavItem to="/" icon={<LayoutDashboard className="w-4 h-4" />} active={pathname === '/'} collapsed={collapsed}>
            Dashboard
          </NavItem>
          <NavItem to="/projects/new" icon={<Plus className="w-4 h-4" />} active={pathname === '/projects/new'} collapsed={collapsed}>
            New Project
          </NavItem>
          <NavItem to="/services" icon={<Database className="w-4 h-4" />} active={pathname === '/services'} collapsed={collapsed}>
            Services
          </NavItem>
          <NavItem to="/marketplace" icon={<Package className="w-4 h-4" />} active={pathname === '/marketplace'} collapsed={collapsed}>
            Marketplace
          </NavItem>
          <NavItem to="/backups" icon={<Archive className="w-4 h-4" />} active={pathname === '/backups'} collapsed={collapsed}>
            Backups
          </NavItem>
          <NavItem to="/users" icon={<Users className="w-4 h-4" />} active={pathname === '/users'} collapsed={collapsed}>
            Team
          </NavItem>
          <NavItem to="/swarm" icon={<Server className="w-4 h-4" />} active={pathname === '/swarm'} collapsed={collapsed}>
            Swarm
          </NavItem>
        </nav>

        {/* Footer */}
        <div className="border-t border-surface-300">
          {/* Collapse toggle (desktop only) */}
          <button
            onClick={toggleCollapsed}
            className="hidden lg:flex w-full items-center justify-center gap-2 px-3 py-2 text-gray-500 hover:text-gray-300 hover:bg-surface-200 transition-colors text-xs"
            title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? <PanelLeftOpen className="w-4 h-4" /> : <PanelLeftClose className="w-4 h-4" />}
            {!collapsed && <span>Collapse</span>}
          </button>
          <div className={`p-3 flex items-center ${collapsed ? 'justify-center' : 'justify-between'}`}>
            {!collapsed && <span className="text-xs text-gray-500">My PaaS v0.4.0</span>}
            <button onClick={handleLogout} className="text-gray-500 hover:text-gray-300" title="Logout">
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto bg-surface pt-14 lg:pt-0">
        <Outlet />
      </main>
    </div>
  )
}

function NavItem({ to, icon, active, collapsed, children }: {
  to: string
  icon: React.ReactNode
  active: boolean
  collapsed: boolean
  children: React.ReactNode
}) {
  return (
    <Link
      to={to}
      title={collapsed ? String(children) : undefined}
      className={`flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors ${
        collapsed ? 'justify-center' : ''
      } ${
        active
          ? 'bg-accent/15 text-accent-hover'
          : 'text-gray-400 hover:bg-surface-200 hover:text-gray-200'
      }`}
    >
      {icon}
      {!collapsed && children}
    </Link>
  )
}
