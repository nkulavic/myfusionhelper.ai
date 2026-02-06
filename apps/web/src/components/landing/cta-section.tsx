'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { ArrowRight } from 'lucide-react'
import { fadeUp, staggerContainer } from './animation-variants'

export function CTASection() {
  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-[hsl(212,100%,22%)] to-[hsl(220,69%,40%)] py-20 md:py-28">
      {/* Decorative blobs */}
      <div className="absolute -left-32 top-0 h-64 w-64 rounded-full bg-brand-green/10 blur-3xl" />
      <div className="absolute -right-20 bottom-0 h-48 w-48 rounded-full bg-brand-blue/10 blur-3xl" />

      <div className="container relative z-10">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.05 }}
          variants={staggerContainer}
          className="mx-auto max-w-2xl text-center"
        >
          <motion.h2
            variants={fadeUp}
            className="text-3xl font-bold text-white md:text-4xl lg:text-5xl"
          >
            Stop wrestling with your CRM.
            <br />
            Start getting results.
          </motion.h2>
          <motion.p
            variants={fadeUp}
            className="mt-4 text-lg text-white/70"
          >
            Join 1,200+ professionals who use MyFusion Helper to save hours every week.
            Connect your CRM, configure your first helper, and see results in minutes.
          </motion.p>
          <motion.div variants={fadeUp} className="mt-8">
            <Link
              href="/register"
              className="inline-flex h-12 items-center gap-2 rounded-lg bg-brand-green px-10 text-sm font-semibold text-white shadow-lg shadow-brand-green/25 transition-all hover:bg-brand-green/90 hover:shadow-xl"
            >
              Start Free Trial
              <ArrowRight className="h-4 w-4" />
            </Link>
          </motion.div>
          <motion.p variants={fadeUp} className="mt-4 text-sm text-white/40">
            14 days free. No credit card required. Cancel anytime.
          </motion.p>
        </motion.div>
      </div>
    </section>
  )
}
