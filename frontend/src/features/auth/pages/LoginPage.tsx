import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Eye, EyeOff, Lock, Mail, PenTool, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { FieldShell, Input } from '@/components/ui/Input'
import { useSession } from '@/hooks/useSession'
import { useToast } from '@/hooks/useToast'

export function LoginPage() {
  const { login } = useSession()
  const navigate = useNavigate()
  const toast = useToast()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const [fieldError, setFieldError] = useState<string | null>(null)

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFieldError(null)

    if (!email.trim() || !password) {
      setFieldError('Enter your email and password to continue.')
      return
    }

    setLoading(true)
    try {
      await login(email.trim(), password)
      toast.success('Welcome back', 'You’re signed in.')
      navigate('/app/dashboard', { replace: true })
    } catch (err) {
      const message =
        err && typeof err === 'object' && 'message' in err
          ? String((err as { message: unknown }).message)
          : 'Sign-in failed. Please try again.'
      toast.error('Sign-in failed', message)
      setFieldError(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-dvh bg-surface">
      {/* Brand panel */}
      <div className="relative hidden w-1/2 overflow-hidden bg-slate-900 lg:block">
        <div
          className="absolute inset-0 opacity-40"
          style={{
            background:
              'radial-gradient(600px 400px at 20% 15%, rgba(79,70,229,0.35), transparent 60%), radial-gradient(500px 380px at 80% 85%, rgba(99,102,241,0.22), transparent 60%)',
          }}
          aria-hidden
        />
        <div className="relative flex h-full flex-col justify-between p-10">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary-600 text-white shadow-lg shadow-primary-600/40">
              <PenTool className="h-5 w-5" aria-hidden />
            </div>
            <div>
              <p className="text-sm font-semibold text-white">SignFlow</p>
              <p className="text-[11px] text-slate-400">Operations console</p>
            </div>
          </div>

          <div className="max-w-md">
            <h1 className="text-3xl font-bold leading-tight text-white">
              Every contract, signed and sealed.
            </h1>
            <p className="mt-3 text-sm leading-relaxed text-slate-400">
              Create contracts, route them to signers, and keep a tamper-proof audit trail — all
              from one console.
            </p>
            <ul className="mt-8 space-y-3">
              {[
                'Cursor-paginated lists with live summaries',
                'Role-based access for every action',
                'Audit trail on every state change',
              ].map((f) => (
                <li key={f} className="flex items-center gap-2.5 text-sm text-slate-300">
                  <ShieldCheck className="h-4 w-4 shrink-0 text-primary-400" aria-hidden />
                  {f}
                </li>
              ))}
            </ul>
          </div>

          <p className="text-[11px] text-slate-500">AeroXe SignFlow · v1.0</p>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex w-full items-center justify-center px-4 py-12 lg:w-1/2">
        <div className="w-full max-w-sm animate-slide-up">
          <div className="mb-8 lg:hidden">
            <div className="flex items-center gap-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary-600 text-white">
                <PenTool className="h-5 w-5" aria-hidden />
              </div>
              <p className="text-sm font-semibold text-slate-900">SignFlow</p>
            </div>
          </div>

          <h2 className="text-xl font-semibold text-slate-900">Sign in to your console</h2>
          <p className="mt-1 text-sm text-slate-500">
            Use your administrator credentials to continue.
          </p>

          <form onSubmit={onSubmit} className="mt-8 space-y-4" noValidate>
            <Input
              id="email"
              type="email"
              label="Email address"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@company.com"
              autoComplete="email"
              icon={<Mail className="h-4 w-4" />}
            />
            <FieldShell label="Password" error={fieldError ?? undefined} htmlFor="password">
              <div className="relative">
                <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-slate-400">
                  <Lock className="h-4 w-4" />
                </span>
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  autoComplete="current-password"
                  className="h-9 w-full rounded-lg border border-slate-200 bg-white pl-9 pr-10 text-sm shadow-sm transition-colors placeholder:text-slate-400 hover:border-slate-300 focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute inset-y-0 right-0 flex items-center pr-3 text-slate-400 hover:text-slate-600"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </FieldShell>

            <Button type="submit" loading={loading} className="mt-2 w-full" size="lg">
              Sign in
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
