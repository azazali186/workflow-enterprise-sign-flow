import { usePageTitle } from '@/hooks/usePageTitle'
import { LandingHero } from './LandingHero'
import { LandingLifecycle } from './LandingLifecycle'
import { LandingFeatures } from './LandingFeatures'
import { LandingSecurity } from './LandingSecurity'
import { LandingCTA } from './LandingCTA'

export function LandingPage() {
  usePageTitle('Secure e-Signing Platform')
  return (
    <>
      <LandingHero />
      <LandingLifecycle />
      <LandingFeatures />
      <LandingSecurity />
      <LandingCTA />
    </>
  )
}
