'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { ChevronDown } from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'
import { cn } from '@/lib/utils'

const faqs = [
  {
    question: 'How does MyFusion Helper connect to my CRM?',
    answer:
      'We use official OAuth2 or API key connections depending on your platform. For Keap, GoHighLevel, and HubSpot, you authorize through a secure OAuth flow. For ActiveCampaign and Ontraport, you provide your API key. We never store your CRM login credentials.',
  },
  {
    question: 'Is my data secure?',
    answer:
      'Yes. All data is encrypted in transit (TLS 1.3) and at rest (AES-256). Our infrastructure runs on AWS with SOC 2 compliant services. We follow the principle of least privilege — helpers only access the specific CRM data they need.',
  },
  {
    question: 'Can I use MyFusion Helper with multiple CRM platforms?',
    answer:
      'Absolutely. You can connect multiple CRM platforms simultaneously. The Grow plan supports up to 3 connections and the Deliver plan supports unlimited connections. This is ideal for agencies managing clients across different platforms.',
  },
  {
    question: 'What happens when a helper fails?',
    answer:
      'Failed executions are logged with full error details. You can configure email or Slack notifications for failures. The AI insights system analyzes error patterns and suggests fixes. Transient failures like API timeouts are automatically retried.',
  },
  {
    question: 'Do I need coding skills?',
    answer:
      'No. Every helper is configured through a visual interface with dropdowns, toggles, and form fields. For advanced users, we support custom webhook payloads and conditional logic, but these are all configured visually.',
  },
  {
    question: 'How do webhooks work?',
    answer:
      'Each helper gets a unique webhook URL. You add this URL as an HTTP POST action in your CRM automation or campaign. When the automation fires, it sends contact data to MyFusion Helper, which executes the configured helper. Full request and response logs are available.',
  },
  {
    question: 'Can I cancel anytime?',
    answer:
      'Yes. All plans are month-to-month with no long-term commitments. You can cancel from your Settings page at any time. Annual plans are billed upfront but can be canceled with a prorated refund.',
  },
  {
    question: 'What CRM platforms do you support?',
    answer:
      'We currently support Keap (Infusionsoft), GoHighLevel, ActiveCampaign, Ontraport, and HubSpot. Each platform connects through OAuth or API key authentication. The same 62 helpers work across all supported CRMs through a unified connector layer.',
  },
  {
    question: 'How is MyFusion Helper different from Zapier or Make?',
    answer:
      'Zapier and Make are general-purpose integration platforms. MyFusion Helper is purpose-built for CRM automation. Our 62 helpers are designed specifically for CRM operations -- tagging, scoring, formatting, date calculations, contact merging, and more. They run faster, cost less, and require no multi-step flow building.',
  },
  {
    question: 'What kind of support do you offer?',
    answer:
      'All plans include email support. The Grow plan adds priority support with faster response times. The Deliver plan includes phone support for direct, immediate help. We also provide documentation, in-app AI chat assistance, and a knowledge base with setup guides.',
  },
]

function FAQItem({
  question,
  answer,
}: {
  question: string
  answer: string
}) {
  const [open, setOpen] = useState(false)

  return (
    <div className="border-b last:border-0">
      <button
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between py-5 text-left"
      >
        <span className="pr-8 font-medium">{question}</span>
        <ChevronDown
          className={cn(
            'h-5 w-5 flex-shrink-0 text-muted-foreground transition-transform duration-200',
            open && 'rotate-180'
          )}
        />
      </button>
      <div
        className={cn(
          'grid transition-all duration-200',
          open ? 'grid-rows-[1fr] pb-5' : 'grid-rows-[0fr]'
        )}
      >
        <div className="overflow-hidden">
          <p className="text-sm leading-relaxed text-muted-foreground">{answer}</p>
        </div>
      </div>
    </div>
  )
}

export function FAQSection() {
  return (
    <section id="faq" className="py-20 md:py-28">
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
            FAQ
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            Common questions, clear answers
          </motion.h2>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={fadeUp}
          className="mx-auto mt-12 max-w-3xl rounded-xl border bg-card"
        >
          <div className="divide-y px-6">
            {faqs.map((faq) => (
              <FAQItem key={faq.question} question={faq.question} answer={faq.answer} />
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  )
}
