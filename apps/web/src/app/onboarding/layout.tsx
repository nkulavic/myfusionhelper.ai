import Image from 'next/image'
import Link from 'next/link'

export default function OnboardingLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-br from-[hsl(212,100%,12%)] via-[hsl(212,80%,8%)] to-[hsl(220,60%,14%)] dark:from-[hsl(212,100%,8%)] dark:via-[hsl(212,80%,5%)] dark:to-[hsl(220,60%,10%)]">
      {/* Subtle grid pattern */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.025]"
        style={{
          backgroundImage:
            'linear-gradient(to right, rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.5) 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      {/* Green glow — top right */}
      <div className="pointer-events-none absolute -top-24 -right-16 h-[450px] w-[450px] rounded-full bg-[radial-gradient(circle,hsla(77,85%,45%,0.12),transparent_70%)] blur-[80px] dark:bg-[radial-gradient(circle,hsla(77,85%,45%,0.12),transparent_70%)]" />

      {/* Blue glow — bottom left */}
      <div className="pointer-events-none absolute -bottom-20 -left-10 h-[400px] w-[400px] rounded-full bg-[radial-gradient(circle,hsla(220,69%,56%,0.1),transparent_70%)] blur-[80px]" />

      {/* Small green accent — center left */}
      <div className="pointer-events-none absolute left-[8%] top-[55%] h-[200px] w-[200px] rounded-full bg-[radial-gradient(circle,hsla(77,85%,45%,0.07),transparent_70%)] blur-[50px]" />

      {/* Content */}
      <div className="relative z-10">
        <header className="flex h-14 items-center justify-center px-6">
          <Link href="/" className="flex items-center gap-2">
            <Image
              src="/logo.png"
              alt="MyFusion Helper"
              width={160}
              height={20}
              className="dark:hidden"
            />
            <Image
              src="/logo-full.png"
              alt="MyFusion Helper"
              width={160}
              height={20}
              className="hidden dark:block"
            />
          </Link>
        </header>
        <main className="mx-auto max-w-3xl px-4 py-10">{children}</main>
      </div>
    </div>
  )
}
