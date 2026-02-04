import Link from 'next/link'
import { ArrowRight, Zap, Shield, BarChart3, Sparkles } from 'lucide-react'

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-14 items-center">
          <div className="mr-4 flex">
            <Link href="/" className="mr-6 flex items-center space-x-2">
              <Zap className="h-6 w-6" />
              <span className="font-bold">MyFusion Helper</span>
            </Link>
          </div>
          <div className="flex flex-1 items-center justify-end space-x-4">
            <nav className="flex items-center space-x-6">
              <Link href="/login" className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
                Sign in
              </Link>
              <Link
                href="/register"
                className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90"
              >
                Get Started
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero */}
      <main className="flex-1">
        <section className="container flex flex-col items-center justify-center gap-4 pb-8 pt-6 md:py-10 lg:py-24">
          <div className="flex max-w-[980px] flex-col items-center gap-4 text-center">
            <span className="inline-flex items-center rounded-lg bg-muted px-3 py-1 text-sm font-medium">
              <Sparkles className="mr-1.5 h-4 w-4" />
              AI-Powered Automation
            </span>
            <h1 className="text-3xl font-bold leading-tight tracking-tighter md:text-5xl lg:text-6xl lg:leading-[1.1]">
              CRM automation that works
              <br className="hidden sm:inline" />
              <span className="text-muted-foreground">across all your platforms</span>
            </h1>
            <p className="max-w-[750px] text-lg text-muted-foreground sm:text-xl">
              Connect Keap, GoHighLevel, ActiveCampaign, and more. Build powerful automations with AI assistance. No coding required.
            </p>
          </div>
          <div className="flex gap-4">
            <Link
              href="/register"
              className="inline-flex h-11 items-center justify-center rounded-md bg-primary px-8 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90"
            >
              Start Free Trial
              <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
            <Link
              href="/demo"
              className="inline-flex h-11 items-center justify-center rounded-md border border-input bg-background px-8 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              Watch Demo
            </Link>
          </div>
        </section>

        {/* Features */}
        <section className="container py-8 md:py-12 lg:py-24">
          <div className="mx-auto grid max-w-5xl gap-8 md:grid-cols-3">
            <div className="flex flex-col items-center gap-2 text-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <Zap className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-bold">90+ Helpers</h3>
              <p className="text-muted-foreground">
                Pre-built automation helpers for tags, fields, webhooks, and more.
              </p>
            </div>
            <div className="flex flex-col items-center gap-2 text-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <Shield className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-bold">Multi-Platform</h3>
              <p className="text-muted-foreground">
                Works with Keap, GoHighLevel, ActiveCampaign, Ontraport, and more.
              </p>
            </div>
            <div className="flex flex-col items-center gap-2 text-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <BarChart3 className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-bold">AI Insights</h3>
              <p className="text-muted-foreground">
                Get intelligent reports, email assistance, and automation suggestions.
              </p>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t py-6 md:py-0">
        <div className="container flex flex-col items-center justify-between gap-4 md:h-16 md:flex-row">
          <p className="text-sm text-muted-foreground">
            Built by MyFusion Solutions. Open source on GitHub.
          </p>
        </div>
      </footer>
    </div>
  )
}
