import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { Reveal } from '../components/Section'
import { SignatureStroke } from '../components/SignatureStroke'

const points = [
  ['AES-256-GCM', 'sensitive data encrypted at rest'],
  ['Audit logs', 'every change, actor, and request id'],
  ['Role-based access', 'permissioned down to the route'],
  ['Short-lived tokens', 'sessions revocable instantly'],
]

/** Dark band: the security story told with restraint and one signature. */
export function LandingSecurity() {
  return (
    <section className="overflow-hidden bg-slate-950 text-slate-300">
      <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-24">
        <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2">
          <Reveal>
            <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-primary-400">
              Security
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-tight text-white sm:text-4xl">
              Built like the documents it holds: sealed, dated, and tamper-evident.
            </h2>
            <p className="mt-4 max-w-lg text-base leading-relaxed text-slate-400">
              SignFlow encrypts what matters, logs everything that happens, and lets you revoke
              access the moment you need to. Production secrets are validated at boot — weak
              configuration simply won't run.
            </p>
            <Link
              to="/security"
              className="group mt-7 inline-flex items-center gap-2 text-sm font-medium text-primary-300 transition-colors hover:text-primary-200"
            >
              Read the security model
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" aria-hidden />
            </Link>
          </Reveal>

          <Reveal delay={0.1}>
            <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-8">
              <ul className="space-y-5">
                {points.map(([k, v]) => (
                  <li key={k} className="flex items-baseline justify-between gap-4 border-b border-slate-800 pb-4 last:border-0 last:pb-0">
                    <span className="text-sm font-semibold text-white">{k}</span>
                    <span className="text-right text-xs text-slate-500">{v}</span>
                  </li>
                ))}
              </ul>
              <div className="mt-8 flex justify-end">
                <SignatureStroke className="h-12 w-24 text-primary-500/80" />
              </div>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  )
}
