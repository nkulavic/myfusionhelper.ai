import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@myfusionhelper/ui', '@myfusionhelper/types'],
}

export default nextConfig
