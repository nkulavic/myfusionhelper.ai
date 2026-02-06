'use client'

import { useState } from 'react'
import Link from 'next/link'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, Minus } from 'lucide-react'
import { fadeUp, staggerContainer, scaleUp } from './animation-variants'
import { cn } from '@/lib/utils'

const plans = [
  {
    name: 'Start',
    monthlyPrice: 39,
    annualPrice: 33,
    description: "When you're starting out",
    features: [
      '10 active helpers',
      '5,000 executions/mo',
      '1 CRM connection',
      'Webhooks & rest hooks',
      'Email support',
      'Execution logs (7 days)',
    ],
    cta: 'Start 14-Day Trial',
    highlighted: false,
  },
  {
    name: 'Grow',
    monthlyPrice: 59,
    annualPrice: 49,
    description: 'For growing teams and businesses',
    features: [
      '50 active helpers',
      '50,000 executions/mo',
      '3 CRM connections',
      'AI insights & email composer',
      'Scheduled reports',
      'Priority support',
      'Execution logs (30 days)',
    ],
    cta: 'Start 14-Day Trial',
    highlighted: true,
  },
  {
    name: 'Deliver',
    monthlyPrice: 79,
    annualPrice: 66,
    description: 'Built for expert marketers',
    features: [
      'Unlimited active helpers',
      'Unlimited executions',
      'Unlimited CRM connections',
      'AI insights & email composer',
      'Scheduled reports',
      'Phone & priority support',
      'Execution logs (90 days)',
    ],
    cta: 'Start 14-Day Trial',
    highlighted: false,
  },
]

const comparisonRows = [
  { feature: 'Active Helpers', start: '10', grow: '50', deliver: 'Unlimited' },
  { feature: 'Executions/month', start: '5,000', grow: '50,000', deliver: 'Unlimited' },
  { feature: 'CRM Connections', start: '1', grow: '3', deliver: 'Unlimited' },
  { feature: 'Webhook Triggers', start: true, grow: true, deliver: true },
  { feature: 'AI Insights', start: false, grow: true, deliver: true },
  { feature: 'AI Email Composer', start: false, grow: true, deliver: true },
  { feature: 'Data Explorer', start: 'Basic', grow: 'Full', deliver: 'Full' },
  { feature: 'Scheduled Reports', start: false, grow: true, deliver: true },
  { feature: 'Log Retention', start: '7 days', grow: '30 days', deliver: '90 days' },
  { feature: 'Support', start: 'Email', grow: 'Priority', deliver: 'Phone + Priority' },
  { feature: 'Team Members', start: '1', grow: '3', deliver: '10' },
  { feature: 'API Access', start: false, grow: true, deliver: true },
]

function ComparisonCell({ value }: { value: string | boolean }) {
  if (typeof value === 'boolean') {
    return value ? (
      <Check className="mx-auto h-4 w-4 text-brand-green" />
    ) : (
      <Minus className="mx-auto h-4 w-4 text-muted-foreground/40" />
    )
  }
  return <span className="text-sm">{value}</span>
}

export function PricingTeaserSection() {
  const [isAnnual, setIsAnnual] = useState(false)

  return (
    <section id="pricing" className="bg-muted/30 py-20 md:py-28">
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
            Pricing
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Simple, transparent pricing
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            All plans include a 14-day free trial. Upgrade anytime as you grow.
          </motion.p>

          {/* Billing toggle */}
          <motion.div variants={fadeUp} className="mt-8 flex items-center justify-center gap-3">
            <span
              className={cn(
                'text-sm font-medium transition-colors',
                !isAnnual ? 'text-foreground' : 'text-muted-foreground'
              )}
            >
              Monthly
            </span>
            <button
              type="button"
              role="switch"
              aria-checked={isAnnual}
              onClick={() => setIsAnnual(!isAnnual)}
              className={cn(
                'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors',
                isAnnual ? 'bg-brand-green' : 'bg-muted-foreground/30'
              )}
            >
              <span
                className={cn(
                  'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
                  isAnnual ? 'translate-x-5' : 'translate-x-0'
                )}
              />
            </button>
            <span
              className={cn(
                'text-sm font-medium transition-colors',
                isAnnual ? 'text-foreground' : 'text-muted-foreground'
              )}
            >
              Annual
            </span>
            {isAnnual && (
              <span className="rounded-full bg-brand-green/10 px-2.5 py-0.5 text-xs font-semibold text-brand-green">
                Save up to 17%
              </span>
            )}
          </motion.div>
        </motion.div>

        {/* Pricing cards */}
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-12 grid gap-6 md:grid-cols-3"
        >
          {plans.map((plan) => {
            const price = isAnnual ? plan.annualPrice : plan.monthlyPrice
            const savings = plan.monthlyPrice - plan.annualPrice

            return (
              <motion.div
                key={plan.name}
                variants={plan.highlighted ? scaleUp : fadeUp}
                className={cn(
                  'relative rounded-xl border bg-card p-8 transition-shadow hover:shadow-lg',
                  plan.highlighted && 'border-primary shadow-md scale-[1.02]'
                )}
              >
                {plan.highlighted && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-brand-green px-4 py-1 text-xs font-semibold text-white">
                    Most Popular
                  </span>
                )}
                <h3 className="text-lg font-semibold">{plan.name}</h3>
                <div className="mt-4 flex items-baseline gap-1">
                  <AnimatePresence mode="wait">
                    <motion.span
                      key={price}
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: 10 }}
                      transition={{ duration: 0.2 }}
                      className="text-4xl font-bold"
                    >
                      ${price}
                    </motion.span>
                  </AnimatePresence>
                  <span className="text-muted-foreground">/month</span>
                </div>
                {isAnnual && (
                  <p className="mt-1 text-xs text-brand-green">
                    Save ${savings * 12}/year
                  </p>
                )}
                <p className="mt-2 text-sm text-muted-foreground">
                  {plan.description}
                </p>
                <ul className="mt-6 space-y-3">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-center gap-2 text-sm">
                      <Check className="h-4 w-4 flex-shrink-0 text-brand-green" />
                      {feature}
                    </li>
                  ))}
                </ul>
                <Link
                  href="/register"
                  className={cn(
                    'mt-8 flex h-10 w-full items-center justify-center rounded-md text-sm font-semibold transition-colors',
                    plan.highlighted
                      ? 'bg-brand-green text-white hover:bg-brand-green/90'
                      : 'border border-input bg-background hover:bg-accent'
                  )}
                >
                  {plan.cta}
                </Link>
              </motion.div>
            )
          })}
        </motion.div>

        {/* Comparison table */}
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={fadeUp}
          className="mt-16"
        >
          <h3 className="mb-6 text-center text-lg font-semibold">Compare plans</h3>
          <div className="overflow-x-auto rounded-xl border bg-card">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="px-6 py-4 text-left font-semibold">Feature</th>
                  <th className="px-6 py-4 text-center font-semibold">Start</th>
                  <th className="px-6 py-4 text-center font-semibold text-brand-green">Grow</th>
                  <th className="px-6 py-4 text-center font-semibold">Deliver</th>
                </tr>
              </thead>
              <tbody>
                {comparisonRows.map((row) => (
                  <tr key={row.feature} className="border-b last:border-0">
                    <td className="px-6 py-3 text-muted-foreground">{row.feature}</td>
                    <td className="px-6 py-3 text-center">
                      <ComparisonCell value={row.start} />
                    </td>
                    <td className="px-6 py-3 text-center bg-brand-green/5">
                      <ComparisonCell value={row.grow} />
                    </td>
                    <td className="px-6 py-3 text-center">
                      <ComparisonCell value={row.deliver} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
