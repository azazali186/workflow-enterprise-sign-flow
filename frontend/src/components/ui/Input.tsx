import { forwardRef, type InputHTMLAttributes, type SelectHTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

const fieldBase =
  'w-full rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 ' +
  'placeholder:text-slate-400 shadow-sm transition-colors duration-150 ' +
  'hover:border-slate-300 focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10 ' +
  'disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400'

export interface FieldShellProps {
  label?: string
  hint?: string
  error?: string
  required?: boolean
  className?: string
  htmlFor?: string
}

/** Shared label/hint/error wrapper used by every form control. */
export function FieldShell({
  label,
  hint,
  error,
  required,
  className,
  htmlFor,
  children,
}: FieldShellProps & { children: React.ReactNode }) {
  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      {label && (
        <label htmlFor={htmlFor} className="text-xs font-medium text-slate-600">
          {label}
          {required && <span className="ml-0.5 text-danger-500">*</span>}
        </label>
      )}
      {children}
      {error ? (
        <p role="alert" className="text-xs text-danger-600">
          {error}
        </p>
      ) : hint ? (
        <p className="text-xs text-slate-400">{hint}</p>
      ) : null}
    </div>
  )
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  hint?: string
  error?: string
  /** Left-side icon. */
  icon?: React.ReactNode
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, hint, error, required, icon, ...props }, ref) => {
    return (
      <FieldShell label={label} hint={hint} error={error} required={required} htmlFor={props.id}>
        <div className="relative">
          {icon && (
            <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-slate-400">
              {icon}
            </span>
          )}
          <input
            ref={ref}
            className={cn(fieldBase, 'h-9', icon && 'pl-9', error && 'border-danger-500', className)}
            {...props}
          />
        </div>
      </FieldShell>
    )
  },
)
Input.displayName = 'Input'

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  hint?: string
  error?: string
  options: { value: string; label: string }[]
  placeholder?: string
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, label, hint, error, required, options, placeholder, ...props }, ref) => {
    return (
      <FieldShell label={label} hint={hint} error={error} required={required} htmlFor={props.id}>
        <select
          ref={ref}
          className={cn(fieldBase, 'h-9 pr-8 appearance-none bg-no-repeat', className)}
          style={{
            backgroundImage:
              "url(\"data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E\")",
            backgroundPosition: 'right 0.6rem center',
          }}
          {...props}
        >
          {placeholder && <option value="">{placeholder}</option>}
          {options.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      </FieldShell>
    )
  },
)
Select.displayName = 'Select'

export interface BareSelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  options: { value: string; label: string }[]
}

/** Label-less select for toolbars and compact filters. */
export const BareSelect = forwardRef<HTMLSelectElement, BareSelectProps>(
  ({ options, ...props }, ref) => {
    return (
      <select ref={ref} className={cn(fieldBase, 'h-8 w-auto pr-8 text-xs')} {...props}>
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    )
  },
)
BareSelect.displayName = 'BareSelect'

export { fieldBase }
