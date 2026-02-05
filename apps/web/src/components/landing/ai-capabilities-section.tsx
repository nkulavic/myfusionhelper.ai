'use client'

import { motion } from 'framer-motion'
import { Sparkles, Mail, Lightbulb } from 'lucide-react'
import { fadeUp, slideInLeft, slideInRight, staggerContainer } from './animation-variants'

const capabilities = [
  {
    title: 'Smart Insights',
    description:
      'AI analyzes your execution patterns, identifies failures, and surfaces actionable recommendations to optimize your automations.',
    icon: Sparkles,
    features: ['Anomaly detection', 'Performance recommendations', 'Failure analysis'],
  },
  {
    title: 'AI Email Composer',
    description:
      'Generate personalized emails with AI. Choose your tone, add merge fields, and create subject lines that convert.',
    icon: Mail,
    features: ['Tone control', 'Subject line generation', 'Personalization tokens'],
  },
  {
    title: 'Automation Suggestions',
    description:
      'Get intelligent suggestions for new helpers based on your CRM data patterns and business goals.',
    icon: Lightbulb,
    features: ['Pattern recognition', 'Helper recommendations', 'Workflow optimization'],
  },
]

export function AICapabilitiesSection() {
  return (
    <section className="overflow-hidden py-20 md:py-28">
      <div className="container px-6 lg:px-8">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, margin: '-100px' }}
          variants={staggerContainer}
          className="text-center"
        >
          <motion.p
            variants={fadeUp}
            className="mb-3 text-sm font-medium uppercase tracking-widest text-brand-green"
          >
            AI-First
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Powered by artificial intelligence
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            AI isn&apos;t a bolt-on feature — it&apos;s woven into every part of MyFusion Helper.
          </motion.p>
        </motion.div>

        <div className="mt-16 space-y-20">
          {capabilities.map((cap, i) => {
            const isReversed = i % 2 !== 0
            return (
              <motion.div
                key={cap.title}
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true, margin: '-50px' }}
                variants={staggerContainer}
                className={`flex flex-col items-center gap-12 md:flex-row ${
                  isReversed ? 'md:flex-row-reverse' : ''
                }`}
              >
                {/* Text */}
                <motion.div
                  variants={isReversed ? slideInRight : slideInLeft}
                  className="flex-1 space-y-4"
                >
                  <div className="inline-flex rounded-lg bg-primary/10 p-3">
                    <cap.icon className="h-6 w-6 text-primary" />
                  </div>
                  <h3 className="text-2xl font-bold">{cap.title}</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {cap.description}
                  </p>
                  <ul className="space-y-2">
                    {cap.features.map((feat) => (
                      <li
                        key={feat}
                        className="flex items-center gap-2 text-sm text-muted-foreground"
                      >
                        <div className="h-1.5 w-1.5 rounded-full bg-brand-green" />
                        {feat}
                      </li>
                    ))}
                  </ul>
                </motion.div>

                {/* Visual */}
                <motion.div
                  variants={isReversed ? slideInLeft : slideInRight}
                  className="flex-1"
                >
                  <div className="rounded-xl border bg-card p-6 shadow-lg">
                    <div className="mb-4 flex items-center gap-2">
                      <cap.icon className="h-5 w-5 text-primary" />
                      <span className="text-sm font-semibold">{cap.title}</span>
                    </div>
                    <div className="space-y-3">
                      {cap.features.map((feat, j) => (
                        <div
                          key={feat}
                          className="flex items-center gap-3 rounded-lg bg-muted/50 p-3"
                        >
                          <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center">
                            <span className="text-xs font-bold text-primary">
                              {j + 1}
                            </span>
                          </div>
                          <div className="flex-1">
                            <div className="text-sm font-medium">{feat}</div>
                            <div className="mt-0.5 h-1.5 w-2/3 rounded-full bg-muted">
                              <div
                                className="h-full rounded-full bg-brand-green/50"
                                style={{ width: `${80 - j * 15}%` }}
                              />
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </motion.div>
              </motion.div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
