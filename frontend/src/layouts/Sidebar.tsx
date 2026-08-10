import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { ChevronDown, ChevronsLeft, ChevronsRight, PenTool, X } from 'lucide-react'
import { useAppDispatch, useAppSelector } from '@/store'
import { setMobileNav, toggleSidebar } from '@/store/slices/uiSlice'
import { navigation, type NavItem } from './navigation'
import { usePermission } from '@/hooks/usePermission'
import { hasRoutePermission } from '@/utils/permissions'
import { cn } from '@/utils/cn'

function visible(items: NavItem[], canSee: (p?: string) => boolean): NavItem[] {
  return items
    .filter((i) => canSee(i.permission))
    .map((i) =>
      i.children ? { ...i, children: i.children.filter((c) => canSee(c.permission)) } : i,
    )
    .filter((i) => !i.children || i.children.length > 0)
}

export function Sidebar() {
  const collapsed = useAppSelector((s) => s.ui.sidebarCollapsed)
  const mobileOpen = useAppSelector((s) => s.ui.mobileNavOpen)
  const dispatch = useAppDispatch()
  const { user } = usePermission()

  const items = visible(navigation, (p) =>
    !p ? true : hasRoutePermission(user, p.split(' ')[0], p.split(' ')[1]),
  )

  const inner = (
    <div className="flex h-full flex-col">
      <Brand collapsed={collapsed} />
      <nav className="flex-1 space-y-0.5 overflow-y-auto px-3 py-4" aria-label="Main">
        {items.map((item) => (
          <NavItemRow key={item.path} item={item} collapsed={collapsed} />
        ))}
      </nav>
      <div className="border-t border-slate-800/60 px-4 py-3">
        <p className="text-[11px] text-slate-500">
          {collapsed ? 'v1.0' : 'SignFlow console · v1.0'}
        </p>
      </div>
    </div>
  )

  return (
    <>
      {/* Desktop sidebar */}
      <aside
        className={cn(
          'hidden lg:flex h-dvh shrink-0 flex-col bg-slate-900 text-slate-300 transition-all duration-200 ease-out',
          collapsed ? 'w-[72px]' : 'w-64',
        )}
      >
        {inner}
        <button
          onClick={() => dispatch(toggleSidebar())}
          className="absolute bottom-16 hidden h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-800 hover:text-slate-200 lg:inline-flex"
          style={{ left: collapsed ? '20px' : '224px' }}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronsRight className="h-4 w-4" /> : <ChevronsLeft className="h-4 w-4" />}
        </button>
      </aside>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 animate-fade-in bg-slate-950/50"
            onClick={() => dispatch(setMobileNav(false))}
            aria-hidden
          />
          <aside className="absolute inset-y-0 left-0 w-72 animate-slide-in-right bg-slate-900 text-slate-300 shadow-modal">
            <button
              onClick={() => dispatch(setMobileNav(false))}
              className="absolute right-3 top-4 rounded-lg p-1.5 text-slate-400 hover:bg-slate-800 hover:text-slate-200"
              aria-label="Close navigation"
            >
              <X className="h-5 w-5" />
            </button>
            {inner}
          </aside>
        </div>
      )}
    </>
  )
}

function Brand({ collapsed }: { collapsed: boolean }) {
  return (
    <div className={cn('flex h-16 items-center gap-2.5 border-b border-slate-800/60 px-4', collapsed && 'justify-center px-2')}>
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-600 text-white shadow-sm shadow-primary-600/40">
        <PenTool className="h-4.5 w-4.5" aria-hidden />
      </div>
      {!collapsed && (
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-white">SignFlow</p>
          <p className="text-[11px] text-slate-500">Operations console</p>
        </div>
      )}
    </div>
  )
}

function NavItemRow({ item, collapsed }: { item: NavItem; collapsed: boolean }) {
  const location = useLocation()
  const [open, setOpen] = useState(() => location.pathname.startsWith(item.path))
  const hasChildren = Boolean(item.children?.length)

  if (hasChildren) {
    const activeChild = item.children!.some((c) => location.pathname.startsWith(c.path))
    return (
      <div>
        <button
          onClick={() => setOpen((v) => !v)}
          className={cn(
            'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150',
            activeChild ? 'bg-slate-800 text-white' : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-100',
            collapsed && 'justify-center px-0',
          )}
          aria-expanded={open}
        >
          <item.icon className="h-4.5 w-4.5 shrink-0" aria-hidden />
          {!collapsed && (
            <>
              <span className="flex-1 truncate text-left">{item.label}</span>
              <ChevronDown
                className={cn('h-3.5 w-3.5 transition-transform duration-200', open && 'rotate-180')}
                aria-hidden
              />
            </>
          )}
        </button>
        {open && !collapsed && (
          <div className="ml-4 mt-0.5 space-y-0.5 border-l border-slate-800 pl-3 animate-fade-in">
            {item.children!.map((child) => (
              <NavLink
                key={child.path}
                to={child.path}
                className={({ isActive }) =>
                  cn(
                    'flex items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-[13px] transition-colors duration-150',
                    isActive
                      ? 'bg-primary-600/15 text-primary-300'
                      : 'text-slate-500 hover:bg-slate-800/60 hover:text-slate-200',
                  )
                }
              >
                <child.icon className="h-4 w-4 shrink-0" aria-hidden />
                {child.label}
              </NavLink>
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <NavLink
      to={item.path}
      className={({ isActive }) =>
        cn(
          'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150',
          isActive
            ? 'bg-primary-600/20 font-medium text-primary-200'
            : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-100',
          collapsed && 'justify-center px-0',
        )
      }
    >
      <item.icon className="h-4.5 w-4.5 shrink-0" aria-hidden />
      {!collapsed && <span className="truncate">{item.label}</span>}
    </NavLink>
  )
}
