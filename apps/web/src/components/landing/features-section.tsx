'use client'

import { motion } from 'framer-motion'
import {
  Blocks,
  Shield,
  Sparkles,
  Mail,
  BarChart3,
  Webhook,
} from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

const features = [
  {
    title: '60+ Automation Helpers',
    description:
      'Pre-built helpers for tags, fields, scoring, formatting, dates, notifications, and more. Configure in minutes.',
    icon: Blocks,
    gradient: 'from-brand-green/20 to-brand-green/5',
  },
  {
    title: 'Multi-Platform Support',
    description:
      'One unified platform for Keap, GoHighLevel, ActiveCampaign, and Ontraport. Switch CRMs without rebuilding.',
    icon: Shield,
    gradient: 'from-brand-blue/20 to-brand-blue/5',
  },
  {
    title: 'AI-Powered Insights',
    description:
      'Get intelligent recommendations, anomaly detection, and optimization suggestions for your automations.',
    icon: Sparkles,
    gradient: 'from-brand-green/20 to-brand-blue/5',
  },
  {
    title: 'Email Composer',
    description:
      'AI-assisted email writing with personalization tokens, tone control, and subject line generation.',
    icon: Mail,
    gradient: 'from-brand-blue/20 to-brand-green/5',
  },
  {
    title: 'Advanced Reports',
    description:
      'Scheduled reports on helper performance, execution trends, contact growth, and error analysis.',
    icon: BarChart3,
    gradient: 'from-brand-green/15 to-brand-green/5',
  },
  {
    title: 'Webhook Triggers',
    description:
      'Trigger helpers from your CRM automations via API. Real-time execution with detailed logging.',
    icon: Webhook,
    gradient: 'from-brand-blue/15 to-brand-blue/5',
  },
]

export function FeaturesSection() {
  return (
    <section id="features" className="py-20 md:py-28">
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
            Features
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Everything you need to automate your CRM
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            From simple tag operations to complex AI-driven workflows, MyFusion Helper
            has you covered.
          </motion.p>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 grid gap-6 sm:grid-cols-2 lg:grid-cols-3"
        >
          {features.map((feature) => (
            <motion.div
              key={feature.title}
              variants={fadeUp}
              className="group rounded-xl border bg-card p-6 transition-all hover:border-primary/50 hover:shadow-lg hover:-translate-y-1"
            >
              <div
                className={`mb-4 inline-flex rounded-lg bg-gradient-to-br ${feature.gradient} p-3`}
              >
                <feature.icon className="h-6 w-6 text-primary" />
              </div>
              <h3 className="mb-2 text-lg font-semibold">{feature.title}</h3>
              <p className="text-sm leading-relaxed text-muted-foreground">
                {feature.description}
              </p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
