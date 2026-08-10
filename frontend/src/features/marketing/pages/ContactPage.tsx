import { useState, type FormEvent } from 'react'
import { Mail, MapPin, Send } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { PageHero } from '../components/PageHero'
import { Reveal, Section } from '../components/Section'
import { FieldShell, Input } from '@/components/ui/Input'
import { useToast } from '@/hooks/useToast'

const CONTACT_EMAIL = 'hello@signflow.local'

export function ContactPage() {
  usePageTitle('Contact')
  const toast = useToast()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    const next: Record<string, string> = {}
    if (!name.trim()) next.name = 'Your name is required.'
    if (!email.trim()) next.email = 'An email is required.'
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) next.email = 'Enter a valid email address.'
    if (!message.trim()) next.message = 'A short message helps us help you.'
    setErrors(next)
    if (Object.keys(next).length > 0) return

    // Compose a real mailto — no fake API, works immediately.
    const subject = encodeURIComponent(`[SignFlow] Message from ${name}`)
    const body = encodeURIComponent(`From: ${name} <${email}>\n\n${message}`)
    window.location.href = `mailto:${CONTACT_EMAIL}?subject=${subject}&body=${body}`
    toast.success('Opening your mail client', `A draft to ${CONTACT_EMAIL} is ready.`)
  }

  return (
    <>
      <PageHero
        eyebrow="Contact"
        title="Talk to a human about your signing workflow"
        lead="Sales, onboarding, or a question about the product — we read everything and reply fast."
      />
      <Section>
        <div className="grid grid-cols-1 gap-12 lg:grid-cols-5">
          <Reveal className="lg:col-span-2">
            <div className="space-y-6">
              <div className="flex items-start gap-3">
                <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary-600">
                  <Mail className="h-5 w-5" aria-hidden />
                </span>
                <div>
                  <h3 className="text-sm font-semibold text-slate-900">Email</h3>
                  <a href={`mailto:${CONTACT_EMAIL}`} className="text-sm text-slate-500 hover:text-slate-700">
                    {CONTACT_EMAIL}
                  </a>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary-600">
                  <MapPin className="h-5 w-5" aria-hidden />
                </span>
                <div>
                  <h3 className="text-sm font-semibold text-slate-900">Headquarters</h3>
                  <p className="text-sm text-slate-500">AeroXe SignFlow, remote-first</p>
                </div>
              </div>
              <p className="text-sm leading-relaxed text-slate-500">
                Prefer to explore first? Sign in to the console and see the workflow in action —
                no sales call required.
              </p>
            </div>
          </Reveal>

          <Reveal delay={0.08} className="lg:col-span-3">
            <form onSubmit={onSubmit} noValidate className="space-y-4 rounded-2xl border border-slate-200/80 bg-white p-6 sm:p-8">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Input
                  id="contact-name"
                  label="Your name"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  error={errors.name}
                  placeholder="Ada Lovelace"
                />
                <Input
                  id="contact-email"
                  type="email"
                  label="Email address"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  error={errors.email}
                  placeholder="ada@company.com"
                />
              </div>
              <FieldShell label="Message" required htmlFor="contact-message" error={errors.message}>
                <textarea
                  id="contact-message"
                  rows={5}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="Tell us about your signing workflow…"
                  className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm shadow-sm transition-colors placeholder:text-slate-400 focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10"
                />
              </FieldShell>
              <button
                type="submit"
                className="inline-flex h-11 items-center gap-2 rounded-xl bg-primary-600 px-6 text-sm font-medium text-white shadow-sm shadow-primary-600/25 transition-all duration-200 hover:bg-primary-700 hover:shadow-md"
              >
                <Send className="h-4 w-4" aria-hidden /> Send message
              </button>
            </form>
          </Reveal>
        </div>
      </Section>
    </>
  )
}
