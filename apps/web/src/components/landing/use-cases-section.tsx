'use client'

import { motion } from 'framer-motion'
import {
  Tag,
  Bell,
  FileText,
  RefreshCw,
  Star,
  Eraser,
  Webhook,
  ShieldAlert,
} from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

const useCases = [
  {
    title: 'Auto-tag leads by source',
    description:
      'Automatically apply tags to new contacts based on their lead source, UTM parameters, or referral data.',
    icon: Tag,
  },
  {
    title: 'VIP engagement alerts',
    description:
      'Get notified instantly when a high-value contact goes cold. AI monitors engagement and alerts you before opportunities slip.',
    icon: Bell,
  },
  {
    title: 'Weekly automation digest',
    description:
      'Receive a scheduled report every Monday with helper performance, execution trends, and recommended optimizations.',
    icon: FileText,
  },
  {
    title: 'Cross-platform contact sync',
    description:
      'Keep contact data consistent across all your CRMs. When a field updates in one, it propagates everywhere.',
    icon: RefreshCw,
  },
  {
    title: 'Smart lead scoring',
    description:
      'Automatically calculate and update lead scores based on engagement, activity, and custom criteria.',
    icon: Star,
  },
  {
    title: 'Data cleanup on autopilot',
    description:
      'Format phone numbers, standardize names, trim whitespace, and fix date formats across thousands of contacts.',
    icon: Eraser,
  },
  {
    title: 'Conditional webhook routing',
    description:
      'Route webhook payloads to different endpoints based on contact tags, field values, or custom logic.',
    icon: Webhook,
  },
  {
    title: 'Error monitoring and recovery',
    description:
      'Get alerted when helpers fail. AI analyzes error patterns and suggests fixes. Automatic retry handles transient failures.',
    icon: ShieldAlert,
  },
]

export function UseCasesSection() {
  return (
    <section className="bg-muted/30 py-20 md:py-28">
      <div className="container">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.1 }}
          variants={staggerContainer}
          className="text-center"
        >
          <motion.p
            variants={fadeUp}
            className="mb-3 text-sm font-medium uppercase tracking-widest text-brand-green"
          >
            Use Cases
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Real automation for real businesses
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            See how CRM professionals use MyFusion Helper to save hours every week.
          </motion.p>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 grid gap-4 sm:grid-cols-2 lg:grid-cols-4"
        >
          {useCases.map((useCase) => (
            <motion.div
              key={useCase.title}
              variants={fadeUp}
              className="group rounded-xl border bg-card p-5 transition-all hover:border-primary/50 hover:shadow-lg"
            >
              <div className="mb-3 inline-flex rounded-lg bg-brand-green/10 p-2.5">
                <useCase.icon className="h-5 w-5 text-brand-green" />
              </div>
              <h3 className="mb-2 font-semibold">{useCase.title}</h3>
              <p className="text-sm leading-relaxed text-muted-foreground">
                {useCase.description}
              </p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
