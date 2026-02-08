import Image from 'next/image'
import Link from 'next/link'

const footerLinks = {
  Product: [
    { label: 'Helpers', href: '#capabilities' },
    { label: 'Data Explorer', href: '#capabilities' },
    { label: 'AI Insights', href: '#capabilities' },
    { label: 'Pricing', href: '#pricing' },
  ],
  Platforms: [
    { label: 'Keap', href: '#platforms' },
    { label: 'GoHighLevel', href: '#platforms' },
    { label: 'ActiveCampaign', href: '#platforms' },
    { label: 'Ontraport', href: '#platforms' },
    { label: 'HubSpot', href: '#platforms' },
    { label: 'Stripe', href: '#platforms' },
  ],
  Company: [
    { label: 'About', href: '#' },
    { label: 'Blog', href: '#' },
    { label: 'Contact', href: '#' },
  ],
  Legal: [
    { label: 'Privacy Policy', href: '/privacy' },
    { label: 'Terms of Service', href: '/terms' },
    { label: 'EULA', href: '/eula' },
  ],
}

export function LandingFooter() {
  return (
    <footer className="border-t bg-[hsl(212,100%,22%)]">
      <div className="container py-12 md:py-16">
        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-6">
          {/* Brand */}
          <div className="lg:col-span-2">
            <Link href="/" className="inline-block">
              <Image
                src="/logo.png"
                alt="MyFusion Helper"
                width={160}
                height={20}
                className="brightness-0 invert"
              />
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-white/50">
              AI-powered CRM intelligence for Keap, GoHighLevel, ActiveCampaign,
              Ontraport, HubSpot, and Stripe. Built by MyFusion Solutions.
            </p>
          </div>

          {/* Link columns */}
          {Object.entries(footerLinks).map(([title, links]) => (
            <div key={title}>
              <h4 className="mb-4 text-sm font-semibold text-white">{title}</h4>
              <ul className="space-y-2.5">
                {links.map((link) => (
                  <li key={link.label}>
                    <Link
                      href={link.href}
                      className="text-sm text-white/50 transition-colors hover:text-white/80"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-4 border-t border-white/10 pt-8 md:flex-row">
          <p className="text-sm text-white/40">
            &copy; {new Date().getFullYear()} MyFusion Solutions. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  )
}
