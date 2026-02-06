'use client'

import { motion } from 'framer-motion'
import { Link2, Settings, Zap } from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

const steps = [
  {
    number: '01',
    title: 'Connect Your CRM',
    description:
      'Link your Keap, GoHighLevel, ActiveCampaign, Ontraport, or HubSpot account with OAuth or API key. Encrypted and connected in under 60 seconds.',
    icon: Link2,
  },
  {
    number: '02',
    title: 'Configure Helpers',
    description:
      'Browse 62 pre-built helpers across 7 categories. Configure them through a visual UI and get a unique webhook URL for each one.',
    icon: Settings,
  },
  {
    number: '03',
    title: 'Trigger and Monitor',
    description:
      'Drop the webhook URL into your CRM automation. When it fires, your helper executes in real-time. Track success rates and get AI-powered optimization suggestions.',
    icon: Zap,
  },
]

export function HowItWorksSection() {
  return (
    <section
      id="how-it-works"
      className="relative overflow-hidden bg-[hsl(212,100%,22%)] py-20 md:py-28"
    >
      {/* Decorative blobs */}
      <div className="absolute -right-32 top-0 h-80 w-80 rounded-full bg-brand-green/5 blur-3xl" />
      <div className="absolute -left-20 bottom-0 h-64 w-64 rounded-full bg-brand-blue/5 blur-3xl" />

      <div className="container relative z-10">
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
            How It Works
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold text-white md:text-4xl"
          >
            Up and running in three steps
          </motion.h2>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="relative mt-16 grid gap-8 md:grid-cols-3"
        >
          {/* Connecting line */}
          <div className="absolute left-0 right-0 top-12 hidden h-px bg-gradient-to-r from-transparent via-brand-green/30 to-transparent md:block" />

          {steps.map((step) => (
            <motion.div
              key={step.number}
              variants={fadeUp}
              className="relative flex flex-col items-center text-center"
            >
              <div className="relative mb-6">
                <div className="flex h-24 w-24 items-center justify-center rounded-2xl border border-brand-green/20 bg-brand-green/10">
                  <step.icon className="h-10 w-10 text-brand-green" />
                </div>
                <span className="absolute -right-2 -top-2 flex h-8 w-8 items-center justify-center rounded-full bg-brand-green text-sm font-bold text-white">
                  {step.number}
                </span>
              </div>
              <h3 className="mb-2 text-xl font-semibold text-white">
                {step.title}
              </h3>
              <p className="max-w-xs text-sm leading-relaxed text-white/60">
                {step.description}
              </p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
