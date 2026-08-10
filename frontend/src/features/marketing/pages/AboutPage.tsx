import { usePageTitle } from '@/hooks/usePageTitle'
import { PageHero } from '../components/PageHero'
import { Reveal, Section } from '../components/Section'
import { SignatureStroke } from '../components/SignatureStroke'

const values = [
  {
    title: 'Trust is earned, then recorded',
    body: 'We believe every agreement should carry a complete, checkable story — not just a signature on the last page.',
  },
  {
    title: 'Precision over theatrics',
    body: 'Quiet, deliberate product design. Every screen, transition, and message exists to make the work clearer.',
  },
  {
    title: 'Security is a feature',
    body: 'Encryption, revocation, and audit are not add-ons. They are the product working as intended.',
  },
]

export function AboutPage() {
  usePageTitle('About')
  return (
    <>
      <PageHero
        eyebrow="About"
        title="SignFlow exists to make signing something you never have to worry about"
        lead="We build the plumbing of agreements — draft, sign, verify, archive — so teams can commit to paper without the paper."
      />
      <Section>
        <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2">
          <Reveal>
            <p className="text-lg leading-relaxed text-slate-600">
              SignFlow started with a simple observation: most agreements are still signed by hand,
              emailed, or lost in a folder. The contracts that matter most deserve a record that is
              just as serious as the terms themselves.
            </p>
            <p className="mt-4 text-base leading-relaxed text-slate-500">
              So we built a console where every change is logged, every signature is captured with
              context, and every executed contract ends in a certificate — all behind the same
              calm interface.
            </p>
            <div className="mt-8">
              <SignatureStroke className="h-14 w-28 text-primary-600" />
              <p className="mt-2 font-mono text-xs text-slate-400">— the SignFlow team</p>
            </div>
          </Reveal>

          <div className="space-y-4">
            {values.map((v, i) => (
              <Reveal key={v.title} delay={i * 0.08}>
                <div className="rounded-2xl border border-slate-200/80 bg-white p-6 transition-all duration-200 hover:border-primary-200 hover:shadow-pop">
                  <p className="font-mono text-xs text-slate-300 tabular">0{i + 1}</p>
                  <h3 className="mt-2 text-base font-semibold text-slate-900">{v.title}</h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-slate-500">{v.body}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </Section>
    </>
  )
}
