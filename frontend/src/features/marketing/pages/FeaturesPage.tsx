import { usePageTitle } from '@/hooks/usePageTitle'
import { PageHero } from '../components/PageHero'
import { Reveal, Section } from '../components/Section'
import {
  BellRing,
  FileClock,
  Fingerprint,
  History,
  Lock,
  Scale,
  Send,
  Workflow,
} from 'lucide-react'

const groups = [
  {
    title: 'Create & manage',
    items: [
      {
        icon: Send,
        name: 'Signature requests',
        body: 'Send a request in one click. Signers receive a dedicated flow and a record of every attempt.',
      },
      {
        icon: Workflow,
        name: 'Multi-party signing',
        body: 'Ordered signers, roles, and per-party status — track exactly where an agreement stands.',
      },
      {
        icon: FileClock,
        name: 'Version history',
        body: 'Every edit is captured with before/after data, so the final signature reflects the real agreement.',
      },
    ],
  },
  {
    title: 'Verify & trust',
    items: [
      {
        icon: Fingerprint,
        name: 'Captured signatures',
        body: 'Draw, type, or upload. Each capture stores type, IP, user agent, and timestamp.',
      },
      {
        icon: BellRing,
        name: 'Identity verification',
        body: 'OTP-backed verification ties the signature to the person who signed it.',
      },
      {
        icon: Lock,
        name: 'Certificates',
        body: 'Signed agreements get issued certificates with serial numbers and validity windows.',
      },
    ],
  },
  {
    title: 'Govern',
    items: [
      {
        icon: History,
        name: 'Immutable audit log',
        body: 'Append-only records of every action — actor, timestamp, request id, before and after.',
      },
      {
        icon: Scale,
        name: 'Role-based access',
        body: 'Permissions are mapped to API routes. System roles are protected from mutation.',
      },
      {
        icon: FileClock,
        name: 'Compliance checks',
        body: 'Attach review evidence to contracts and keep an approval trail for your obligations.',
      },
    ],
  },
]

export function FeaturesPage() {
  usePageTitle('Features')
  return (
    <>
      <PageHero
        eyebrow="Features"
        title="The full signing workflow, in one place"
        lead="From the first draft to the archived certificate — every capability is built for teams that need to prove what happened."
      />
      <Section>
        <div className="space-y-16">
          {groups.map((group, gi) => (
            <div key={group.title}>
              <Reveal>
                <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-400">
                  {String(gi + 1).padStart(2, '0')} · {group.title}
                </h2>
              </Reveal>
              <div className="mt-6 grid grid-cols-1 gap-5 md:grid-cols-3">
                {group.items.map((item, i) => (
                  <Reveal key={item.name} delay={i * 0.06}>
                    <div className="h-full rounded-2xl border border-slate-200/80 bg-white p-6 transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-200 hover:shadow-pop">
                      <item.icon className="h-5 w-5 text-primary-600" aria-hidden />
                      <h3 className="mt-4 text-base font-semibold text-slate-900">{item.name}</h3>
                      <p className="mt-2 text-sm leading-relaxed text-slate-500">{item.body}</p>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>
          ))}
        </div>
      </Section>
    </>
  )
}
