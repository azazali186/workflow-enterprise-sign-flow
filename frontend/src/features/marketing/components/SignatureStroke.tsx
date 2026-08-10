import { motion, useReducedMotion } from 'framer-motion'
import { EASE } from '../lib/motion'

/**
 * The signature motif: a handwritten "S" stroke drawn with SVG path animation
 * when it scrolls into view. One memorable element per the design direction —
 * the act of signing is what this product is about.
 */
export function SignatureStroke({ className }: { className?: string }) {
  const reduce = useReducedMotion()

  return (
    <svg
      viewBox="0 0 120 60"
      fill="none"
      className={className}
      role="img"
      aria-label="Animated signature stroke"
    >
      <motion.path
        d="M8 38 C 14 8, 20 50, 30 26 S 44 12, 48 30 S 60 50, 66 30 S 76 14, 82 32 S 96 52, 112 28"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
        initial={{ pathLength: 0, opacity: 0 }}
        whileInView={{ pathLength: 1, opacity: 1 }}
        viewport={{ once: true, amount: 0.6 }}
        transition={reduce ? { duration: 0 } : { duration: 1.4, ease: EASE, delay: 0.2 }}
      />
      {/* The dot that completes the signature. */}
      <motion.circle
        cx="112"
        cy="28"
        r="3"
        fill="currentColor"
        initial={{ scale: 0, opacity: 0 }}
        whileInView={{ scale: 1, opacity: 1 }}
        viewport={{ once: true, amount: 0.6 }}
        transition={reduce ? { duration: 0 } : { duration: 0.2, delay: 1.5 }}
      />
    </svg>
  )
}
