import { Reveal, Section, SectionHead } from '../components/Section'
import { FileCheck2, FileSignature, SearchCheck, ShieldCheck } from 'lucide-react'

/**
 * The signing lifecycle is a genuine sequence — Draft → Sign → Verify →
 * Archive — so numbered markers encode real order, not decoration.
 */
const steps = [
  {
    icon: FileSignature,
    title: 'Draft',
    body: 'Build agreements from templates with signers attached, expiration dates, and full version history from the very first edit.',
  },
  {
    icon: FileCheck2,
    title: 'Sign',
    body: 'Send a signature request in one click. Every signer gets their own flow — draw, type, or upload — tracked to the second.',
  },
  {
    icon: SearchCheck,
    title: 'Verify',
    body: 'Identity checks confirm the right person signed. Certificates and verification records are issued automatically.',
  },
  {
    icon: ShieldCheck,
    title: 'Archive',
    body: 'Executed contracts are sealed with an immutable audit trail — who changed what, when, and from where.',
  },
]

export function LandingLifecycle() {
  return (
    <Section id="how-it-works">
      <SectionHead
        eyebrow="How it works"
        title="From draft to sealed in four steps"
        lead="Every stage is recorded against the same agreement, so the final signature carries the whole story with it."
      />
      <div className="mt-14 grid grid-cols-1 gap-10 sm:grid-cols-2 lg:grid-cols-4">
        {steps.map((step, i) => (
          <Reveal key={step.title} delay={i * 0.08}>
            <div className="group relative">
              <div className="flex items-center gap-3">
                <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-900 text-white transition-transform duration-300 group-hover:-translate-y-0.5">
                  <step.icon className="h-5 w-5" aria-hidden />
                </span>
                <span className="font-mono text-sm text-slate-300 tabular">0{i + 1}</span>
              </div>
              <h3 className="mt-5 text-base font-semibold text-slate-900">{step.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-slate-500">{step.body}</p>
              {i < steps.length - 1 && (
                <div className="absolute -right-5 top-5 hidden h-px w-10 bg-slate-200 lg:block" aria-hidden />
              )}
            </div>
          </Reveal>
        ))}
      </div>
    </Section>
  )
}
