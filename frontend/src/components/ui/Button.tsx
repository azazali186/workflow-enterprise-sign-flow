import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/utils/cn'

type Variant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
type Size = 'sm' | 'md' | 'lg'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  loading?: boolean
  /** Icon-only buttons get a square ratio. */
  iconOnly?: boolean
}

const base =
  'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-150 ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 disabled:pointer-events-none disabled:opacity-50 ' +
  'active:scale-[0.98] select-none whitespace-nowrap'

const variants: Record<Variant, string> = {
  primary:
    'bg-primary-600 text-white shadow-sm shadow-primary-600/20 hover:bg-primary-700 hover:shadow-md hover:shadow-primary-600/25',
  secondary:
    'bg-slate-100 text-slate-700 hover:bg-slate-200 hover:text-slate-900',
  outline:
    'border border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900',
  ghost: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
  danger:
    'bg-danger-600 text-white shadow-sm shadow-danger-600/20 hover:bg-danger-700',
}

const sizes: Record<Size, string> = {
  sm: 'h-8 px-3 text-xs',
  md: 'h-9 px-4 text-sm',
  lg: 'h-11 px-5 text-sm',
}

const iconSizes: Record<Size, string> = {
  sm: 'h-8 w-8',
  md: 'h-9 w-9',
  lg: 'h-11 w-11',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, iconOnly, children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(base, variants[variant], iconOnly ? iconSizes[size] : sizes[size], className)}
        {...props}
      >
        {loading && <Loader2 className="h-4 w-4 animate-spin" aria-hidden />}
        {children}
      </button>
    )
  },
)

Button.displayName = 'Button'
