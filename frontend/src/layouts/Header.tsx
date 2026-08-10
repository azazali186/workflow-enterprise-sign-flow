import { useLocation, useNavigate } from 'react-router-dom'
import { ChevronRight, LogOut, Menu, User as UserIcon } from 'lucide-react'
import { useAppDispatch, useAppSelector } from '@/store'
import { setMobileNav } from '@/store/slices/uiSlice'
import { sessionEnd } from '@/store/slices/authSlice'
import { authService } from '@/services/auth.service'
import { tokenStore } from '@/services/api/client'
import { Dropdown } from '@/components/ui/Dropdown'
import { useToast } from '@/hooks/useToast'
import { humanize } from '@/utils/format'

export function Header() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAppSelector((s) => s.auth.user)
  const toast = useToast()

  const crumbs = location.pathname
    .replace('/app', '')
    .split('/')
    .filter(Boolean)
    .map(humanize)

  const onLogout = async () => {
    try {
      await authService.logout()
    } catch {
      // Token may already be invalid — still clear the local session.
    } finally {
      tokenStore.clear()
      dispatch(sessionEnd())
      navigate('/login', { replace: true })
      toast.info('Signed out', 'Your session has ended.')
    }
  }

  const initials = (user?.name ?? '?')
    .split(' ')
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between gap-4 border-b border-slate-200/80 bg-white/80 px-4 backdrop-blur-md sm:px-6">
      <div className="flex items-center gap-3">
        <button
          onClick={() => dispatch(setMobileNav(true))}
          className="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 lg:hidden"
          aria-label="Open navigation"
        >
          <Menu className="h-5 w-5" />
        </button>
        <nav aria-label="Breadcrumb" className="hidden items-center gap-1.5 text-sm sm:flex">
          <span className="text-slate-400">Console</span>
          {crumbs.map((c, i) => (
            <span key={i} className="flex items-center gap-1.5">
              <ChevronRight className="h-3.5 w-3.5 text-slate-300" aria-hidden />
              <span className={i === crumbs.length - 1 ? 'font-medium text-slate-900' : 'text-slate-500'}>
                {c}
              </span>
            </span>
          ))}
        </nav>
      </div>

      <Dropdown
        trigger={
          <button className="flex items-center gap-2.5 rounded-lg p-1.5 pr-2 transition-colors hover:bg-slate-100">
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
              {initials}
            </span>
            <span className="hidden text-left sm:block">
              <span className="block text-sm font-medium text-slate-900">{user?.name ?? '—'}</span>
              <span className="block text-[11px] text-slate-500">{user?.email ?? ''}</span>
            </span>
          </button>
        }
        items={[
          {
            key: 'me',
            label: 'My account',
            icon: <UserIcon className="h-4 w-4" />,
            onSelect: () => navigate('/app/dashboard'),
          },
          {
            key: 'logout',
            label: 'Sign out',
            icon: <LogOut className="h-4 w-4" />,
            danger: true,
            onSelect: onLogout,
          },
        ]}
      />
    </header>
  )
}
