'use client'

import { motion } from 'framer-motion'
import {
  Blocks,
  Shield,
  Sparkles,
  Mail,
  BarChart3,
  Webhook,
  Search,
  History,
} from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

const features = [
  {
    title: '62 Pre-Built Helpers',
    description:
      'Tag contacts, format data, calculate dates, score leads, sync to Google Sheets, send notifications, and more. Configure in minutes, not hours.',
    icon: Blocks,
    gradient: 'from-brand-green/20 to-brand-green/5',
  },
  {
    title: '6 Platforms',
    description:
      'One account works with Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot, and Stripe. Switch platforms without rebuilding your automations.',
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
    title: 'Data Explorer',
    description:
      'Query your CRM data with filters or natural language. Fast analytics across contacts, tags, deals, and custom fields.',
    icon: Search,
    gradient: 'from-brand-blue/15 to-brand-green/5',
  },
  {
    title: 'Webhook Triggers',
    description:
      'Each helper gets a unique webhook URL. Drop it into your CRM automation and helpers fire in real-time.',
    icon: Webhook,
    gradient: 'from-brand-blue/15 to-brand-blue/5',
  },
  {
    title: 'Execution History',
    description:
      'Every helper run is logged with status, duration, contact details, and error information. Filter and drill into any execution.',
    icon: History,
    gradient: 'from-brand-green/15 to-brand-blue/5',
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
            Everything you need to extend your CRM
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            From simple tag operations to AI-driven analytics, MyFusion Helper covers the
            full range of CRM automation needs.
          </motion.p>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 grid gap-6 sm:grid-cols-2 lg:grid-cols-4"
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
