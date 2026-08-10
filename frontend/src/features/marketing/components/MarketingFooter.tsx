import { Link } from 'react-router-dom'
import { PenTool } from 'lucide-react'

const columns = [
  {
    title: 'Product',
    links: [
      { to: '/features', label: 'Features' },
      { to: '/pricing', label: 'Pricing' },
      { to: '/security', label: 'Security' },
    ],
  },
  {
    title: 'Company',
    links: [
      { to: '/about', label: 'About' },
      { to: '/contact', label: 'Contact' },
    ],
  },
  {
    title: 'Console',
    links: [
      { to: '/login', label: 'Sign in' },
      { to: '/app/dashboard', label: 'Dashboard' },
    ],
  },
]

export function MarketingFooter() {
  return (
    <footer className="border-t border-slate-200 bg-white">
      <div className="mx-auto max-w-6xl px-4 py-12 sm:px-6">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <Link to="/" className="flex items-center gap-2.5" aria-label="SignFlow home">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-600 text-white">
                <PenTool className="h-4 w-4" aria-hidden />
              </span>
              <span className="text-sm font-semibold tracking-tight text-slate-900">SignFlow</span>
            </Link>
            <p className="mt-3 max-w-xs text-sm leading-relaxed text-slate-500">
              Contract signing, sealed with a trusted audit trail. Built for teams that take
              agreements seriously.
            </p>
          </div>

          {columns.map((col) => (
            <div key={col.title}>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                {col.title}
              </h3>
              <ul className="mt-3 space-y-2.5">
                {col.links.map((l) => (
                  <li key={l.label}>
                    <Link
                      to={l.to}
                      className="text-sm text-slate-600 transition-colors duration-150 hover:text-slate-900"
                    >
                      {l.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-10 flex flex-col items-start justify-between gap-3 border-t border-slate-100 pt-6 sm:flex-row sm:items-center">
          <p className="text-xs text-slate-400">
            © {new Date().getFullYear()} AeroXe SignFlow. All rights reserved.
          </p>
          <div className="flex items-center gap-5 text-xs text-slate-400">
            <Link to="/security" className="transition-colors hover:text-slate-700">
              Privacy
            </Link>
            <Link to="/security" className="transition-colors hover:text-slate-700">
              Terms
            </Link>
            <span className="tabular">v1.0</span>
          </div>
        </div>
      </div>
    </footer>
  )
}
