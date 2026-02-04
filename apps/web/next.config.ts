import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@myfusionhelper/ui', '@myfusionhelper/types'],
  experimental: {
    typedRoutes: true,
  },
}

export default nextConfig
