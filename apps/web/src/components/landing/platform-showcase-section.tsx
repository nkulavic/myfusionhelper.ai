'use client'

import { motion } from 'framer-motion'
import { fadeUp, staggerContainer } from './animation-variants'
import { activeCRMPlatforms } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'
import { Check } from 'lucide-react'

const platformDetails: Record<string, { tagline: string }> = {
  keap: {
    tagline:
      'Go beyond Keap\'s native automation limits. Format data, calculate dates, split-test campaigns, and chain automation sequences.',
  },
  gohighlevel: {
    tagline:
      'Fill the gaps in GHL workflows with lead scoring, data formatting, tag-based routing, and A/B testing.',
  },
  activecampaign: {
    tagline:
      'Extend ActiveCampaign with powerful data operations. Format fields, score contacts, and chain complex multi-step workflows.',
  },
  ontraport: {
    tagline:
      'Unlock capabilities Ontraport does not offer natively. Tag-based scoring, data normalization, and conditional routing.',
  },
  hubspot: {
    tagline:
      'Complement HubSpot\'s native tools with deal stage management, property mapping, list sync, and data formatting.',
  },
  stripe: {
    tagline:
      'Automate payment workflows, sync customer data with your CRM, manage subscriptions, and trigger actions on invoice events.',
  },
}

export function PlatformShowcaseSection() {
  return (
    <section id="platforms" className="py-20 md:py-28">
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
            Integrations
          </motion.p>
          <motion.h2
            variants={fadeUp}
            className="mx-auto max-w-2xl text-3xl font-bold md:text-4xl"
          >
            One platform for every CRM
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mx-auto mt-4 max-w-xl text-lg text-muted-foreground"
          >
            Connect in 60 seconds. No coding required. Full API access to your contacts,
            tags, fields, and automations.
          </motion.p>
        </motion.div>

        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mt-16 grid gap-6 sm:grid-cols-2 lg:grid-cols-3"
        >
          {activeCRMPlatforms.map((platform) => {
            const details = platformDetails[platform.id]
            return (
              <motion.div
                key={platform.id}
                variants={fadeUp}
                className="group rounded-xl border bg-card p-6 transition-all hover:border-primary/50 hover:shadow-lg"
              >
                <div className="mb-4 flex items-center gap-3">
                  <PlatformLogo platform={platform} size={48} />
                  <div>
                    <h3 className="text-lg font-semibold">{platform.name}</h3>
                    <span className="text-xs text-muted-foreground">
                      {platform.authType === 'oauth2' ? 'OAuth 2.0 connection' : 'API Key connection'}
                    </span>
                  </div>
                </div>
                <p className="mb-4 text-sm leading-relaxed text-muted-foreground">
                  {details?.tagline}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {platform.capabilities.map((cap) => (
                    <span
                      key={cap}
                      className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-0.5 text-xs text-muted-foreground"
                    >
                      <Check className="h-3 w-3 text-brand-green" />
                      {cap}
                    </span>
                  ))}
                </div>
              </motion.div>
            )
          })}
        </motion.div>
      </div>
    </section>
  )
}
