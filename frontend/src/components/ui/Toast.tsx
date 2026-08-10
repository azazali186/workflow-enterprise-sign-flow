import { CheckCircle2, Info, X, XCircle } from 'lucide-react'
import { useAppDispatch, useAppSelector } from '@/store'
import { dismissToast } from '@/store/slices/uiSlice'
import { cn } from '@/utils/cn'

const icons = {
  success: <CheckCircle2 className="h-4.5 w-4.5 text-success-600" aria-hidden />,
  error: <XCircle className="h-4.5 w-4.5 text-danger-600" aria-hidden />,
  info: <Info className="h-4.5 w-4.5 text-primary-600" aria-hidden />,
}

/** Toast stack — fixed, aria-live, auto-dismissed by middleware. */
export function ToastViewport() {
  const toasts = useAppSelector((s) => s.ui.toasts)
  const dispatch = useAppDispatch()

  return (
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-[60] flex w-full max-w-sm flex-col gap-2"
      aria-live="polite"
    >
      {toasts.map((t) => (
        <div
          key={t.id}
          className={cn(
            'pointer-events-auto flex items-start gap-3 rounded-xl border border-slate-200 bg-white/95 p-3.5 shadow-pop backdrop-blur animate-toast-in',
            t.kind === 'error' ? 'border-danger-200' : '',
          )}
          role="status"
        >
          {icons[t.kind]}
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-slate-900">{t.title}</p>
            {t.description && <p className="mt-0.5 text-xs text-slate-500">{t.description}</p>}
          </div>
          <button
            onClick={() => dispatch(dismissToast(t.id))}
            className="rounded-md p-1 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            aria-label="Dismiss notification"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      ))}
    </div>
  )
}
