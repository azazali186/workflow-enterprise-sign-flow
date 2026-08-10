import { motion, useReducedMotion } from 'framer-motion'
import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'
import { SignatureStroke } from '../components/SignatureStroke'
import { fadeUp, heroContainer } from '../lib/motion'

const container = heroContainer
const item = fadeUp

const stats = [
  { value: '100%', label: 'tamper-proof audit trail' },
  { value: '24/7', label: 'signing, any timezone' },
  { value: '1 file', label: 'per agreement, always in sync' },
]

export function LandingHero() {
  const reduce = useReducedMotion()

  return (
    <section className="relative overflow-hidden">
      {/* Quiet ambient background — indigo wash, not decoration overload. */}
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'radial-gradient(720px 420px at 12% 0%, rgba(99,102,241,0.08), transparent 60%), radial-gradient(600px 380px at 88% 12%, rgba(79,70,229,0.06), transparent 60%)',
        }}
        aria-hidden
      />

      <div className="relative mx-auto max-w-6xl px-4 pb-20 pt-16 sm:px-6 sm:pb-28 sm:pt-24">
        <motion.div
          variants={container}
          initial={reduce ? false : 'hidden'}
          animate={reduce ? undefined : 'visible'}
          className="mx-auto max-w-3xl text-center"
        >
          <motion.p
            variants={item}
            className="inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-3.5 py-1.5 font-mono text-xs font-medium text-primary-700"
          >
            Trusted signing for modern teams
          </motion.p>

          <motion.h1
            variants={item}
            className="mt-6 text-4xl font-semibold tracking-tight text-slate-900 sm:text-6xl sm:leading-[1.05]"
          >
            Contracts that sign themselves,<br className="hidden sm:block" /> audited from the first draft.
          </motion.h1>

          <motion.p
            variants={item}
            className="mx-auto mt-6 max-w-xl text-base leading-relaxed text-slate-500 sm:text-lg"
          >
            Create agreements, route them to signers, and keep a complete, encrypted record of
            every change — so nothing is ever signed on faith.
          </motion.p>

          <motion.div
            variants={item}
            className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row"
          >
            <Link
              to="/login"
              className="group inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-primary-600 px-7 text-sm font-medium text-white shadow-lg shadow-primary-600/25 transition-all duration-200 hover:bg-primary-700 hover:shadow-xl hover:shadow-primary-600/30 sm:w-auto"
            >
              Start signing
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" aria-hidden />
            </Link>
            <Link
              to="/features"
              className="inline-flex h-12 w-full items-center justify-center rounded-xl border border-slate-200 bg-white px-7 text-sm font-medium text-slate-700 transition-all duration-200 hover:border-slate-300 hover:bg-slate-50 sm:w-auto"
            >
              See how it works
            </Link>
          </motion.div>

          {/* The signature moment — draws itself into view. */}
          <motion.div variants={item} className="mt-12 flex justify-center">
            <SignatureStroke className="h-16 w-32 text-primary-600 sm:h-20 sm:w-40" />
          </motion.div>

          <motion.div
            variants={item}
            className="mt-12 grid grid-cols-1 gap-6 border-t border-slate-100 pt-8 sm:grid-cols-3"
          >
            {stats.map((s) => (
              <div key={s.label} className="text-center">
                <p className="text-2xl font-semibold tracking-tight text-slate-900 tabular">{s.value}</p>
                <p className="mt-1 text-xs text-slate-400">{s.label}</p>
              </div>
            ))}
          </motion.div>
        </motion.div>
      </div>
    </section>
  )
}
