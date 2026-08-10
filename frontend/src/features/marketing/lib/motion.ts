import type { Easing, Variants } from 'framer-motion'

/** The site's easing curve — ease-out-quad, shared across every animation. */
export const EASE: Easing = [0.22, 1, 0.36, 1]

/** Entrance variants used by the hero and reveals. */
export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 24 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EASE } },
}

export const heroContainer: Variants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.12 } },
}
