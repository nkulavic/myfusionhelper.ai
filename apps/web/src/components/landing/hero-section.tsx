'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { ArrowRight, Sparkles, Play } from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

export function HeroSection() {
  return (
    <section className="relative min-h-[90vh] overflow-hidden bg-gradient-to-br from-[hsl(212,100%,16%)] via-[hsl(212,100%,22%)] to-[hsl(220,69%,40%)]">
      {/* Decorative blobs */}
      <div className="absolute -left-40 -top-40 h-96 w-96 rounded-full bg-brand-green/10 blur-3xl" />
      <div className="absolute -right-20 top-1/3 h-80 w-80 rounded-full bg-brand-blue/15 blur-3xl" />
      <div className="absolute bottom-0 left-1/3 h-64 w-64 rounded-full bg-brand-green/8 blur-3xl" />

      <div className="container relative z-10 flex min-h-[90vh] flex-col items-center justify-center pb-16 pt-24 text-center">
        <motion.div
          variants={staggerContainer}
          initial="hidden"
          animate="visible"
          className="flex max-w-4xl flex-col items-center gap-6"
        >
          {/* Badge */}
          <motion.div variants={fadeUp}>
            <span className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-4 py-1.5 text-sm font-medium text-white backdrop-blur-sm">
              <Sparkles className="h-4 w-4 text-brand-green" />
              AI-Powered CRM Intelligence
            </span>
          </motion.div>

          {/* Headline */}
          <motion.h1
            variants={fadeUp}
            className="text-4xl font-bold leading-tight tracking-tight text-white sm:text-5xl md:text-6xl lg:text-7xl"
          >
            Your CRM data is a goldmine.
            <br />
            <span className="bg-gradient-to-r from-[hsl(77,85%,55%)] to-[hsl(77,85%,70%)] bg-clip-text text-transparent">
              Start mining it.
            </span>
          </motion.h1>

          {/* Subheadline */}
          <motion.p
            variants={fadeUp}
            className="max-w-2xl text-lg text-white/70 sm:text-xl"
          >
            Visualize your data. Analyze patterns. Strategize with AI recommendations.
            Automate with 62 helpers across Keap, GoHighLevel, ActiveCampaign, Ontraport,
            and HubSpot — no coding required.
          </motion.p>

          {/* CTAs */}
          <motion.div variants={fadeUp} className="flex flex-wrap items-center justify-center gap-4">
            <Link
              href="/register"
              className="inline-flex h-12 items-center gap-2 rounded-lg bg-brand-green px-8 text-sm font-semibold text-white shadow-lg shadow-brand-green/25 transition-all hover:bg-brand-green/90 hover:shadow-xl hover:shadow-brand-green/30"
            >
              Start Free Trial
              <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="#how-it-works"
              className="inline-flex h-12 items-center gap-2 rounded-lg border border-white/20 bg-white/5 px-8 text-sm font-semibold text-white backdrop-blur-sm transition-all hover:bg-white/10"
            >
              <Play className="h-4 w-4" />
              See How It Works
            </a>
          </motion.div>

          {/* Social proof */}
          <motion.p variants={fadeUp} className="mt-4 text-sm text-white/40">
            Trusted by 1,200+ CRM professionals worldwide
          </motion.p>
        </motion.div>

        {/* Product preview mockup */}
        <motion.div
          initial={{ opacity: 0, y: 60 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.4, ease: [0.25, 0.46, 0.45, 0.94] as [number, number, number, number] }}
          className="mt-12 w-full max-w-5xl"
        >
          <div className="relative mx-auto overflow-hidden rounded-xl border border-white/10 bg-white/5 p-1.5 shadow-2xl backdrop-blur-sm">
            <div className="rounded-lg bg-[hsl(215,40%,10%)] p-4">
              {/* Mock dashboard header */}
              <div className="mb-4 flex items-center gap-3">
                <div className="flex gap-1.5">
                  <div className="h-3 w-3 rounded-full bg-red-400/60" />
                  <div className="h-3 w-3 rounded-full bg-yellow-400/60" />
                  <div className="h-3 w-3 rounded-full bg-green-400/60" />
                </div>
                <div className="flex-1 rounded-md bg-white/5 px-3 py-1 text-xs text-white/30 font-mono">
                  app.myfusionhelper.ai/dashboard
                </div>
              </div>
              {/* Mock cards grid */}
              <div className="grid grid-cols-4 gap-3">
                {['Active Helpers', 'Executions Today', 'Success Rate', 'Connections'].map(
                  (label, i) => (
                    <div
                      key={label}
                      className="rounded-lg bg-white/5 p-3"
                    >
                      <div className="text-[10px] text-white/40">{label}</div>
                      <div className="mt-1 text-lg font-bold text-white/80">
                        {['14', '2,847', '99.2%', '5'][i]}
                      </div>
                    </div>
                  )
                )}
              </div>
              {/* Mock table rows */}
              <div className="mt-3 space-y-2">
                {[
                  { name: 'Tag It', status: 'Active', runs: '1,245' },
                  { name: 'Copy It', status: 'Active', runs: '892' },
                  { name: 'Score It', status: 'Active', runs: '710' },
                ].map((row) => (
                  <div key={row.name} className="flex items-center gap-3 rounded-md bg-white/3 px-3 py-2">
                    <div className="h-2 w-2 rounded-full bg-brand-green/60" />
                    <span className="text-xs text-white/60 flex-1">{row.name}</span>
                    <span className="text-[10px] text-white/30">{row.status}</span>
                    <span className="text-[10px] text-white/30">{row.runs} runs</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
          {/* Gradient fade at bottom */}
          <div className="absolute inset-x-0 bottom-0 h-20 bg-gradient-to-t from-[hsl(212,100%,16%)] to-transparent" />
        </motion.div>
      </div>
    </section>
  )
}
