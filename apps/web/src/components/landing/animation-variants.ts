import type { Variants, Transition } from 'framer-motion'

export const transition: Transition = {
  duration: 0.5,
  ease: [0.25, 0.46, 0.45, 0.94] as [number, number, number, number],
}

export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 15 },
  visible: { opacity: 1, y: 0, transition },
}

export const staggerContainer: Variants = {
  hidden: {},
  visible: {
    transition: {
      staggerChildren: 0.08,
    },
  },
}

export const slideInLeft: Variants = {
  hidden: { opacity: 0, x: -30 },
  visible: { opacity: 1, x: 0, transition: { ...transition, duration: 0.6 } },
}

export const slideInRight: Variants = {
  hidden: { opacity: 0, x: 30 },
  visible: { opacity: 1, x: 0, transition: { ...transition, duration: 0.6 } },
}

export const scaleUp: Variants = {
  hidden: { opacity: 0, scale: 0.97 },
  visible: { opacity: 1, scale: 1, transition: { ...transition, duration: 0.4 } },
}
