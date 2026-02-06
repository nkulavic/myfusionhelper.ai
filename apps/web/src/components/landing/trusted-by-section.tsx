'use client'

import { motion } from 'framer-motion'
import { fadeUp } from './animation-variants'
import { crmPlatforms } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'

export function TrustedBySection() {
  return (
    <section className="border-b bg-muted/50 py-12">
      <div className="container">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={fadeUp}
          className="text-center"
        >
          <p className="mb-8 text-sm font-medium uppercase tracking-widest text-muted-foreground">
            Connects with the CRMs you already use
          </p>
          <div className="flex flex-wrap items-center justify-center gap-8 sm:gap-12 md:gap-16">
            {crmPlatforms.map((platform) => (
              <div
                key={platform.id}
                className="group flex items-center gap-2.5 transition-all duration-300 grayscale hover:grayscale-0"
              >
                <PlatformLogo platform={platform} size={40} className="transition-opacity group-hover:opacity-100 opacity-70" />
                <span className="text-lg font-semibold text-muted-foreground transition-colors group-hover:text-foreground">
                  {platform.name}
                </span>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  )
}
