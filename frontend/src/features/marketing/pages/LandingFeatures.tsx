import { motion, useReducedMotion } from 'framer-motion'
import { Link } from 'react-router-dom'
import { Section, SectionHead } from '../components/Section'
import { ArrowUpRight, Fingerprint, History, Lock, Scale, Workflow } from 'lucide-react'
import { cn } from '@/utils/cn'
import { EASE } from '../lib/motion'

const features = [
  {
    icon: Fingerprint,
    title: 'Legally-grounded signatures',
    body: 'Each signature captures who signed, when, and the exact state of the agreement at that moment.',
  },
  {
    icon: History,
    title: 'Immutable audit trail',
    body: 'Every view, edit, and status change lands in an append-only log with actor, timestamp, and request id.',
  },
  {
    icon: Lock,
    title: 'Encrypted at rest',
    body: 'Sensitive fields are sealed with AES-256-GCM. Passwords are hashed, tokens are short-lived, sessions are revocable.',
  },
  {
    icon: Scale,
    title: 'Role-based control',
    body: 'Fine-grained permissions per route. Drafters draft, reviewers review, and administrators govern — nothing more.',
  },
  {
    icon: Workflow,
    title: 'Multi-party workflows',
    body: 'Attach as many signers as you need, each with their own order, role, and completion status.',
  },
  {
    icon: ArrowUpRight,
    title: 'One console for it all',
    body: 'Contracts, signers, certificates, verifications, and audit logs — a single source of truth with live summaries.',
  },
]

export function LandingFeatures() {
  const reduce = useReducedMotion()

  return (
    <Section id="features" className="border-t border-slate-100">
      <SectionHead
        eyebrow="Capabilities"
        title="Everything a serious agreement needs"
        lead="The mechanics of signing handled for you — so your team can focus on the terms, not the process."
      />
      <div className="mt-14 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((f, i) => (
          <motion.div
            key={f.title}
            initial={reduce ? false : { opacity: 0, y: 20 }}
            whileInView={reduce ? undefined : { opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.3 }}
            transition={{ duration: 0.45, delay: (i % 3) * 0.07, ease: EASE }}
            className="group rounded-2xl border border-slate-200/80 bg-white p-6 transition-all duration-200 hover:-translate-y-1 hover:border-primary-200 hover:shadow-pop"
          >
            <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary-600 transition-colors duration-200 group-hover:bg-primary-600 group-hover:text-white">
              <f.icon className="h-5 w-5" aria-hidden />
            </span>
            <h3 className="mt-5 text-base font-semibold text-slate-900">{f.title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-slate-500">{f.body}</p>
          </motion.div>
        ))}
      </div>
      <p className="mt-10 text-center text-sm text-slate-500">
        <Link
          to="/features"
          className={cn(
            'inline-flex items-center gap-1 font-medium text-primary-600 transition-colors hover:text-primary-700',
          )}
        >
          Explore every capability <ArrowUpRight className="h-4 w-4" aria-hidden />
        </Link>
      </p>
    </Section>
  )
}
