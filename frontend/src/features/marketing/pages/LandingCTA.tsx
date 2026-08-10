import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { Reveal } from '../components/Section'

export function LandingCTA() {
  return (
    <section className="mx-auto max-w-6xl px-4 pb-24 sm:px-6">
      <Reveal>
        <div className="relative overflow-hidden rounded-3xl bg-slate-900 px-6 py-16 text-center sm:px-16">
          <div
            className="pointer-events-none absolute inset-0"
            style={{
              background:
                'radial-gradient(520px 320px at 50% -10%, rgba(99,102,241,0.35), transparent 65%)',
            }}
            aria-hidden
          />
          <div className="relative">
            <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-primary-300">
              Ready when you are
            </p>
            <h2 className="mx-auto mt-4 max-w-xl text-3xl font-semibold tracking-tight text-white sm:text-4xl">
              Sign your first contract in minutes, not weeks.
            </h2>
            <p className="mx-auto mt-4 max-w-md text-base leading-relaxed text-slate-400">
              Draft it, attach the parties, and send. Everything after that is automatic — and
              audited.
            </p>
            <Link
              to="/login"
              className="group mt-8 inline-flex h-12 items-center gap-2 rounded-xl bg-primary-500 px-8 text-sm font-medium text-white shadow-lg shadow-primary-500/30 transition-all duration-200 hover:bg-primary-400 hover:shadow-xl hover:shadow-primary-500/40"
            >
              Get started
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" aria-hidden />
            </Link>
          </div>
        </div>
      </Reveal>
    </section>
  )
}
