'use client'

import { motion } from 'framer-motion'
import { Shield, Lock, Clock, Globe, Quote } from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

const testimonials = [
  {
    quote:
      'MyFusion Helper replaced three separate tools we were using. The AI insights alone saved us 5 hours per week.',
    author: 'Marketing Director',
    company: 'SaaS Company',
  },
  {
    quote:
      'Setting up our Keap automations used to take days. With MyFusion Helper, we had everything running in an afternoon.',
    author: 'Operations Manager',
    company: 'E-commerce',
  },
  {
    quote:
      'The Data Explorer is incredible. We can finally see all our contacts across GoHighLevel and ActiveCampaign in one place.',
    author: 'CRM Administrator',
    company: 'Agency',
  },
]

const trustSignals = [
  { icon: Shield, label: 'SOC 2 compliant infrastructure' },
  { icon: Lock, label: 'Encrypted at rest and in transit' },
  { icon: Clock, label: '99.9% uptime SLA' },
  { icon: Globe, label: 'GDPR-ready data handling' },
]

export function SocialProofSection() {
  return (
    <section
      className="relative overflow-hidden bg-[hsl(212,100%,22%)] py-20 md:py-28"
    >
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
            Trusted
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold text-white md:text-4xl"
          >
            Trusted by CRM professionals worldwide
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-white/60"
          >
            Join 1,200+ businesses that rely on MyFusion Helper to run their CRM
            automations.
          </motion.p>
        </motion.div>

        {/* Testimonials */}
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 grid gap-6 md:grid-cols-3"
        >
          {testimonials.map((t) => (
            <motion.div
              key={t.author}
              variants={fadeUp}
              className="rounded-xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
            >
              <Quote className="mb-4 h-6 w-6 text-brand-green/60" />
              <p className="mb-6 text-sm leading-relaxed text-white/70">
                &ldquo;{t.quote}&rdquo;
              </p>
              <div>
                <p className="text-sm font-semibold text-white">{t.author}</p>
                <p className="text-xs text-white/40">{t.company}</p>
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Trust signals */}
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 flex flex-wrap items-center justify-center gap-8"
        >
          {trustSignals.map((signal) => (
            <motion.div
              key={signal.label}
              variants={fadeUp}
              className="flex items-center gap-2.5 text-sm text-white/50"
            >
              <signal.icon className="h-4 w-4 text-brand-green/60" />
              {signal.label}
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
