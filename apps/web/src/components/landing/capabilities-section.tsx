'use client'

import { motion } from 'framer-motion'
import {
  Eye,
  Brain,
  Lightbulb,
  Zap,
  BarChart3,
  Search,
  Download,
  TrendingUp,
  AlertTriangle,
  FileText,
  Target,
  LineChart,
  Star,
  Blocks,
  Tag,
  Mail,
  Webhook,
  Calculator,
} from 'lucide-react'
import { fadeUp, staggerContainer, slideInLeft, slideInRight } from './animation-variants'

const capabilities = [
  {
    label: 'Visualize',
    headline: 'See your business clearly',
    subheadline:
      'Dashboards and data explorer tools that turn scattered CRM records into a complete picture of your business.',
    icon: Eye,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10',
    borderColor: 'border-blue-500/20',
    features: [
      { text: 'Interactive Data Explorer with filters, search, and export', icon: Search },
      { text: 'Real-time dashboard with execution stats and trends', icon: BarChart3 },
      { text: 'Contact detail views with full field visibility', icon: Eye },
      { text: 'Cross-platform data exploration without switching tabs', icon: Download },
    ],
  },
  {
    label: 'Analyze',
    headline: 'Spot what you are missing',
    subheadline:
      'AI-powered insights surface patterns, anomalies, and opportunities hiding in your CRM data -- before they become problems.',
    icon: Brain,
    color: 'text-purple-400',
    bgColor: 'bg-purple-500/10',
    borderColor: 'border-purple-500/20',
    features: [
      { text: 'Anomaly detection that flags unusual execution patterns', icon: AlertTriangle },
      { text: 'Trend analysis showing contact growth and engagement shifts', icon: TrendingUp },
      { text: 'Error analysis with root-cause identification', icon: Target },
      { text: 'Performance reports delivered on your schedule', icon: FileText },
    ],
  },
  {
    label: 'Strategize',
    headline: 'Know exactly what to do next',
    subheadline:
      'AI recommendations turn data into decisions. Get specific, actionable suggestions tailored to your CRM activity and business goals.',
    icon: Lightbulb,
    color: 'text-amber-400',
    bgColor: 'bg-amber-500/10',
    borderColor: 'border-amber-500/20',
    features: [
      { text: 'Helper recommendations based on your CRM usage patterns', icon: Star },
      { text: 'Optimization suggestions to improve success rates', icon: LineChart },
      { text: 'ROI tracking for every automation you run', icon: TrendingUp },
      { text: 'Weekly performance digests with actionable next steps', icon: FileText },
    ],
  },
  {
    label: 'Automate',
    headline: 'Put your CRM on autopilot',
    subheadline:
      '62 pre-built helpers handle the tedious work -- tagging, formatting, scoring, syncing, notifying -- triggered automatically from your CRM workflows.',
    icon: Zap,
    color: 'text-brand-green',
    bgColor: 'bg-brand-green/10',
    borderColor: 'border-brand-green/20',
    features: [
      { text: 'Tag management: apply, remove, toggle, clear by prefix', icon: Tag },
      { text: 'Data operations: copy fields, format values, math calculations', icon: Calculator },
      { text: 'Notifications: email alerts, Slack messages, webhook callbacks', icon: Mail },
      { text: 'Scoring, date math, conditional logic, and custom webhooks', icon: Webhook },
    ],
  },
]

export function CapabilitiesSection() {
  return (
    <section id="capabilities" className="py-20 md:py-28">
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
            Capabilities
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Four ways to get more from your CRM
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            MyFusion Helper goes beyond simple automation. It transforms how you see,
            understand, plan, and act on your CRM data.
          </motion.p>
        </motion.div>

        <div className="mt-16 space-y-20">
          {capabilities.map((cap, index) => {
            const isEven = index % 2 === 0
            return (
              <motion.div
                key={cap.label}
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true, amount: 0.1 }}
                variants={staggerContainer}
                className="grid items-center gap-12 lg:grid-cols-2"
              >
                <motion.div
                  variants={isEven ? slideInLeft : slideInRight}
                  className={isEven ? 'lg:order-1' : 'lg:order-2'}
                >
                  <div
                    className={`mb-4 inline-flex items-center gap-2 rounded-full ${cap.bgColor} px-3 py-1`}
                  >
                    <cap.icon className={`h-4 w-4 ${cap.color}`} />
                    <span className={`text-sm font-semibold ${cap.color}`}>{cap.label}</span>
                  </div>
                  <h3 className="text-2xl font-bold md:text-3xl">{cap.headline}</h3>
                  <p className="mt-3 text-muted-foreground">{cap.subheadline}</p>
                  <ul className="mt-6 space-y-3">
                    {cap.features.map((feature) => (
                      <li key={feature.text} className="flex items-start gap-3">
                        <div
                          className={`mt-0.5 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md ${cap.bgColor}`}
                        >
                          <feature.icon className={`h-3.5 w-3.5 ${cap.color}`} />
                        </div>
                        <span className="text-sm leading-relaxed">{feature.text}</span>
                      </li>
                    ))}
                  </ul>
                </motion.div>

                <motion.div
                  variants={isEven ? slideInRight : slideInLeft}
                  className={isEven ? 'lg:order-2' : 'lg:order-1'}
                >
                  <div
                    className={`rounded-xl border ${cap.borderColor} bg-gradient-to-br from-muted/50 to-muted/20 p-6`}
                  >
                    <CapabilityMockup capability={cap.label} color={cap.color} />
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

function CapabilityMockup({
  capability,
  color,
}: {
  capability: string
  color: string
}) {
  switch (capability) {
    case 'Visualize':
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Search className={`h-4 w-4 ${color}`} />
            Data Explorer
          </div>
          <div className="grid grid-cols-3 gap-2">
            {['Contacts', 'Tags', 'Custom Fields'].map((label) => (
              <div key={label} className="rounded-lg bg-background/60 p-3 text-center">
                <div className="text-lg font-bold">
                  {label === 'Contacts' ? '12,458' : label === 'Tags' ? '342' : '89'}
                </div>
                <div className="text-[11px] text-muted-foreground">{label}</div>
              </div>
            ))}
          </div>
          <div className="space-y-2">
            {[
              'Sarah Chen — VIP, Hot Lead',
              'Marcus Johnson — New, Webinar',
              'Lisa Park — Customer, Active',
            ].map((row) => (
              <div
                key={row}
                className="flex items-center gap-2 rounded-md bg-background/40 px-3 py-2"
              >
                <div className="h-2 w-2 rounded-full bg-brand-green" />
                <span className="text-xs">{row}</span>
              </div>
            ))}
          </div>
        </div>
      )
    case 'Analyze':
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Brain className={`h-4 w-4 ${color}`} />
            AI Insights
          </div>
          <div className="space-y-2">
            {[
              {
                type: 'Anomaly',
                text: 'Tag It helper failure rate increased 340% this week',
              },
              {
                type: 'Trend',
                text: 'Contact growth rate up 12% month-over-month',
              },
              {
                type: 'Suggestion',
                text: '3 helpers have 0% execution rate — consider removing',
              },
            ].map((insight) => (
              <div key={insight.type} className="rounded-lg bg-background/40 p-3">
                <div className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                  {insight.type}
                </div>
                <div className="text-xs leading-relaxed">{insight.text}</div>
              </div>
            ))}
          </div>
        </div>
      )
    case 'Strategize':
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Lightbulb className={`h-4 w-4 ${color}`} />
            Recommendations
          </div>
          <div className="space-y-2">
            {[
              {
                label: 'Add a Score It helper',
                desc: 'Based on your tag patterns, lead scoring would boost engagement tracking',
              },
              {
                label: 'Optimize Format It',
                desc: 'Phone number formatting has 8% error rate — switch to E.164',
              },
              {
                label: 'Weekly digest',
                desc: 'Enable weekly reports to track ROI across all helpers',
              },
            ].map((rec) => (
              <div key={rec.label} className="rounded-lg bg-background/40 p-3">
                <div className="mb-0.5 text-xs font-semibold">{rec.label}</div>
                <div className="text-[11px] leading-relaxed text-muted-foreground">
                  {rec.desc}
                </div>
              </div>
            ))}
          </div>
        </div>
      )
    case 'Automate':
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Blocks className={`h-4 w-4 ${color}`} />
            Helper Library
          </div>
          <div className="grid grid-cols-2 gap-2">
            {[
              { name: 'Tag It', runs: '1,245' },
              { name: 'Copy It', runs: '892' },
              { name: 'Score It', runs: '710' },
              { name: 'Format It', runs: '568' },
              { name: 'Math It', runs: '423' },
              { name: 'Hook It', runs: '312' },
            ].map((helper) => (
              <div
                key={helper.name}
                className="flex items-center gap-2 rounded-lg bg-background/40 px-3 py-2"
              >
                <div className="h-2 w-2 rounded-full bg-brand-green" />
                <div className="flex-1">
                  <div className="text-xs font-medium">{helper.name}</div>
                  <div className="text-[10px] text-muted-foreground">{helper.runs} runs</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )
    default:
      return null
  }
}
