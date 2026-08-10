import type { ReactNode } from 'react'
import { motion, useReducedMotion } from 'framer-motion'
import { cn } from '@/utils/cn'
import { EASE } from '../lib/motion'

/** Scroll-reveal wrapper — one shared entrance so every section feels consistent. */
export function Reveal({
  children,
  className,
  delay = 0,
}: {
  children: ReactNode
  className?: string
  delay?: number
}) {
  const reduce = useReducedMotion()
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 24 }}
      whileInView={reduce ? undefined : { opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.25 }}
      transition={{ duration: 0.5, ease: EASE, delay }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

/** Mono eyebrow label — the data-forward voice of the brand. */
export function Eyebrow({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <p
      className={cn(
        'font-mono text-xs font-medium uppercase tracking-[0.18em] text-primary-600',
        className,
      )}
    >
      {children}
    </p>
  )
}

/** Section header block: eyebrow + heading + lead. */
export function SectionHead({
  eyebrow,
  title,
  lead,
  align = 'center',
  className,
}: {
  eyebrow: string
  title: string
  lead?: string
  align?: 'center' | 'left'
  className?: string
}) {
  return (
    <Reveal
      className={cn(
        'max-w-2xl',
        align === 'center' ? 'mx-auto text-center' : 'text-left',
        className,
      )}
    >
      <Eyebrow className={align === 'center' ? 'justify-center' : ''}>{eyebrow}</Eyebrow>
      <h2 className="mt-3 text-3xl font-semibold tracking-tight text-slate-900 sm:text-4xl">
        {title}
      </h2>
      {lead && <p className="mt-4 text-base leading-relaxed text-slate-500">{lead}</p>}
    </Reveal>
  )
}

/** Full-width section with the site grid and vertical rhythm. */
export function Section({
  children,
  className,
  id,
}: {
  children: ReactNode
  className?: string
  id?: string
}) {
  return (
    <section id={id} className={cn('mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-24', className)}>
      {children}
    </section>
  )
}
