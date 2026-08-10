import { usePageTitle } from '@/hooks/usePageTitle'
import { PageHero } from '../components/PageHero'
import { Reveal, Section } from '../components/Section'
import { Fingerprint, KeyRound, Lock, ShieldCheck, UserX, Workflow } from 'lucide-react'

const items = [
  {
    icon: Lock,
    title: 'Encryption at rest',
    body: 'Sensitive fields — phones, signature data, storage keys — are sealed with AES-256-GCM using a server-side key that is validated at boot.',
  },
  {
    icon: KeyRound,
    title: 'Strong authentication',
    body: 'Passwords are bcrypt-hashed, sessions are single-login with revocable short-lived JWTs, and failed attempts trigger account lockout.',
  },
  {
    icon: ShieldCheck,
    title: 'Brute-force protection',
    body: 'Per-account failure counters and rate limits make credential stuffing impractical, and unknown accounts are indistinguishable from wrong passwords.',
  },
  {
    icon: Workflow,
    title: 'Role-based access',
    body: 'Every API route is a permission. Roles grant access precisely, and changes invalidate cached grants immediately.',
  },
  {
    icon: Fingerprint,
    title: 'Immutable audit trail',
    body: 'Login logs and audit records capture actor, timestamp, request id, and before/after state — with sensitive values scrubbed before they are stored.',
  },
  {
    icon: UserX,
    title: 'Instant revocation',
    body: 'Suspending or deleting a user kills their session and cached permissions immediately — no waiting for token expiry.',
  },
]

export function SecurityPage() {
  usePageTitle('Security')
  return (
    <>
      <PageHero
        eyebrow="Security"
        title="Security that matches the documents it protects"
        lead="SignFlow treats trust as a feature: encryption, lockout, revocation, and an audit trail that can stand up in a review."
      />
      <Section>
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
          {items.map((item, i) => (
            <Reveal key={item.title} delay={(i % 2) * 0.07}>
              <div className="flex h-full gap-4 rounded-2xl border border-slate-200/80 bg-white p-6 transition-all duration-200 hover:border-primary-200 hover:shadow-pop">
                <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-900 text-white">
                  <item.icon className="h-5 w-5" aria-hidden />
                </span>
                <div>
                  <h3 className="text-base font-semibold text-slate-900">{item.title}</h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-slate-500">{item.body}</p>
                </div>
              </div>
            </Reveal>
          ))}
        </div>

        <Reveal className="mt-16">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-8 text-center">
            <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-primary-600">
              Production posture
            </p>
            <p className="mx-auto mt-3 max-w-2xl text-base leading-relaxed text-slate-600">
              In production, SignFlow refuses to boot with weak secrets, enforces server timeouts
              and request limits, contains background-worker panics, and bounds every log and queue
              it writes to. Fail-closed is the default.
            </p>
          </div>
        </Reveal>
      </Section>
    </>
  )
}
