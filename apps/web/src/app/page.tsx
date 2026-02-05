'use client'

import { LandingHeader } from '@/components/landing/landing-header'
import { HeroSection } from '@/components/landing/hero-section'
import { TrustedBySection } from '@/components/landing/trusted-by-section'
import { FeaturesSection } from '@/components/landing/features-section'
import { HowItWorksSection } from '@/components/landing/how-it-works-section'
import { AICapabilitiesSection } from '@/components/landing/ai-capabilities-section'
import { PricingTeaserSection } from '@/components/landing/pricing-teaser-section'
import { CTASection } from '@/components/landing/cta-section'
import { LandingFooter } from '@/components/landing/landing-footer'

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      <LandingHeader />
      <main>
        <HeroSection />
        <TrustedBySection />
        <FeaturesSection />
        <HowItWorksSection />
        <AICapabilitiesSection />
        <PricingTeaserSection />
        <CTASection />
      </main>
      <LandingFooter />
    </div>
  )
}
