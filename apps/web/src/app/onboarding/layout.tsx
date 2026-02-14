import Image from 'next/image'
import Link from 'next/link'

export default function OnboardingLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <header className="flex h-14 items-center justify-center border-b px-6">
        <Link href="/" className="flex items-center gap-2">
          <Image
            src="/logo.png"
            alt="MyFusion Helper"
            width={160}
            height={20}
            className="dark:brightness-0 dark:invert"
          />
        </Link>
      </header>
      <main className="mx-auto max-w-3xl px-4 py-10">{children}</main>
    </div>
  )
}
