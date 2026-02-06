'use client'

import { motion } from 'framer-motion'
import { fadeUp, staggerContainer } from './animation-variants'
import { crmPlatforms } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'
import { Check, Zap } from 'lucide-react'

const platformDetails: Record<string, { tagline: string; helperCount: number }> = {
  keap: {
    tagline:
      'Intelligent tag management, lead scoring automation, and deep contact field manipulation.',
    helperCount: 18,
  },
  gohighlevel: {
    tagline:
      'AI-driven contact segmentation, pipeline optimization, and automated workflow triggers.',
    helperCount: 15,
  },
  activecampaign: {
    tagline:
      'Advanced tag logic, deal stage tracking, and predictive engagement scoring.',
    helperCount: 14,
  },
  ontraport: {
    tagline:
      'Powerful data formatting, sequence management, and cross-platform contact synchronization.',
    helperCount: 10,
  },
  hubspot: {
    tagline:
      'Intelligent pipeline management, property automation, and list-based workflow triggers.',
    helperCount: 12,
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
          {crmPlatforms.map((platform) => {
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
                      {platform.authType === 'oauth2' ? 'OAuth2' : 'API Key'} connection
                    </span>
                  </div>
                </div>
                <p className="mb-4 text-sm leading-relaxed text-muted-foreground">
                  {details?.tagline}
                </p>
                <ul className="mb-4 space-y-1.5">
                  {platform.capabilities.slice(0, 4).map((cap) => (
                    <li key={cap} className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Check className="h-3.5 w-3.5 flex-shrink-0 text-brand-green" />
                      {cap}
                    </li>
                  ))}
                </ul>
                <div className="flex items-center gap-1.5 text-sm font-medium text-brand-green">
                  <Zap className="h-4 w-4" />
                  {details?.helperCount}+ helpers available
                </div>
              </motion.div>
            )
          })}
        </motion.div>
      </div>
    </section>
  )
}
