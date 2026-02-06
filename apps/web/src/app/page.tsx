'use client'

import { LandingHeader } from '@/components/landing/landing-header'
import { HeroSection } from '@/components/landing/hero-section'
import { TrustedBySection } from '@/components/landing/trusted-by-section'
import { CapabilitiesSection } from '@/components/landing/capabilities-section'
import { PlatformShowcaseSection } from '@/components/landing/platform-showcase-section'
import { HowItWorksSection } from '@/components/landing/how-it-works-section'
import { UseCasesSection } from '@/components/landing/use-cases-section'
import { SocialProofSection } from '@/components/landing/social-proof-section'
import { PricingTeaserSection } from '@/components/landing/pricing-teaser-section'
import { FAQSection } from '@/components/landing/faq-section'
import { CTASection } from '@/components/landing/cta-section'
import { LandingFooter } from '@/components/landing/landing-footer'

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      <LandingHeader />
      <main>
        <HeroSection />
        <TrustedBySection />
        <CapabilitiesSection />
        <PlatformShowcaseSection />
        <HowItWorksSection />
        <UseCasesSection />
        <SocialProofSection />
        <PricingTeaserSection />
        <FAQSection />
        <CTASection />
      </main>
      <LandingFooter />
    </div>
  )
}
