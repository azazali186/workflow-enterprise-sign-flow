import { motion, useReducedMotion } from 'framer-motion'
import { Eyebrow } from './Section'
import { EASE } from '../lib/motion'

/** Shared hero for marketing sub-pages — consistent voice, one heading. */
export function PageHero({
  eyebrow,
  title,
  lead,
}: {
  eyebrow: string
  title: string
  lead?: string
}) {
  const reduce = useReducedMotion()
  return (
    <section className="border-b border-slate-100 bg-gradient-to-b from-slate-50/80 to-white">
      <motion.div
        initial={reduce ? false : { opacity: 0, y: 18 }}
        animate={reduce ? undefined : { opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: EASE }}
        className="mx-auto max-w-6xl px-4 py-16 sm:px-6 sm:py-20"
      >
        <Eyebrow>{eyebrow}</Eyebrow>
        <h1 className="mt-3 max-w-2xl text-4xl font-semibold tracking-tight text-slate-900 sm:text-5xl">
          {title}
        </h1>
        {lead && <p className="mt-5 max-w-2xl text-base leading-relaxed text-slate-500 sm:text-lg">{lead}</p>}
      </motion.div>
    </section>
  )
}
