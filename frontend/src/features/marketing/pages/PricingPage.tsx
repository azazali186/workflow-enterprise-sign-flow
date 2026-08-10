import { Link } from 'react-router-dom'
import { Check } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { PageHero } from '../components/PageHero'
import { Reveal, Section } from '../components/Section'
import { cn } from '@/utils/cn'

const tiers = [
  {
    name: 'Starter',
    price: '$0',
    cadence: 'forever',
    blurb: 'For teams evaluating the signing flow.',
    features: [
      'Unlimited drafts',
      'Up to 3 signers per contract',
      'Basic audit trail',
      'Community support',
    ],
    cta: 'Start for free',
    featured: false,
  },
  {
    name: 'Business',
    price: '$49',
    cadence: 'per seat / month',
    blurb: 'For teams that sign every day.',
    features: [
      'Unlimited contracts & signers',
      'Identity verification (OTP)',
      'Certificates & compliance checks',
      'Full audit trail with export',
      'Role-based access control',
      'Priority support',
    ],
    cta: 'Get started',
    featured: true,
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    cadence: 'annual agreement',
    blurb: 'For organizations with governance needs.',
    features: [
      'Everything in Business',
      'SSO & custom roles',
      'Data residency options',
      'Dedicated success manager',
      'SLA & compliance review',
    ],
    cta: 'Talk to sales',
    featured: false,
  },
]

export function PricingPage() {
  usePageTitle('Pricing')
  return (
    <>
      <PageHero
        eyebrow="Pricing"
        title="Simple pricing that scales with your signing volume"
        lead="Start free. Upgrade when your team needs verification, certificates, and governance."
      />
      <Section>
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          {tiers.map((tier, i) => (
            <Reveal key={tier.name} delay={i * 0.08}>
              <div
                className={cn(
                  'relative flex h-full flex-col rounded-2xl border p-7 transition-all duration-200',
                  tier.featured
                    ? 'border-primary-300 bg-slate-900 text-white shadow-pop'
                    : 'border-slate-200 bg-white hover:-translate-y-1 hover:shadow-pop',
                )}
              >
                {tier.featured && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-500 px-3 py-1 font-mono text-[11px] font-medium uppercase tracking-wide text-white">
                    Most popular
                  </span>
                )}
                <h2 className={cn('text-sm font-semibold uppercase tracking-wider', tier.featured ? 'text-primary-300' : 'text-slate-400')}>
                  {tier.name}
                </h2>
                <div className="mt-4 flex items-baseline gap-1.5">
                  <span className="text-4xl font-semibold tracking-tight tabular">{tier.price}</span>
                  <span className={cn('text-xs', tier.featured ? 'text-slate-400' : 'text-slate-400')}>
                    {tier.cadence}
                  </span>
                </div>
                <p className={cn('mt-2 text-sm', tier.featured ? 'text-slate-400' : 'text-slate-500')}>
                  {tier.blurb}
                </p>
                <ul className="mt-6 flex-1 space-y-3">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5 text-sm">
                      <Check
                        className={cn('mt-0.5 h-4 w-4 shrink-0', tier.featured ? 'text-primary-400' : 'text-success-600')}
                        aria-hidden
                      />
                      <span className={tier.featured ? 'text-slate-200' : 'text-slate-600'}>{f}</span>
                    </li>
                  ))}
                </ul>
                <Link
                  to="/login"
                  className={cn(
                    'mt-8 inline-flex h-11 items-center justify-center rounded-xl text-sm font-medium transition-all duration-200',
                    tier.featured
                      ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/30 hover:bg-primary-400'
                      : 'border border-slate-200 text-slate-700 hover:border-slate-300 hover:bg-slate-50',
                  )}
                >
                  {tier.cta}
                </Link>
              </div>
            </Reveal>
          ))}
        </div>
        <Reveal className="mt-12 text-center">
          <p className="text-sm text-slate-500">
            All plans include encryption at rest, a revocable audit trail, and the same production
            security posture. Questions? <Link to="/contact" className="font-medium text-primary-600 hover:text-primary-700">Talk to us</Link>.
          </p>
        </Reveal>
      </Section>
    </>
  )
}
