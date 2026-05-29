/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const workflows = [
    t('Text to Video'),
    t('Image to Video'),
    t('Reference Video'),
    t('Video Remix'),
  ]

  return (
    <section className='relative z-10 overflow-hidden px-6 pt-24 pb-16 md:pt-32 md:pb-24'>
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-30 dark:opacity-[0.14]'
        style={{
          background: [
            'radial-gradient(ellipse 58% 48% at 18% 16%, oklch(0.72 0.16 255 / 72%) 0%, transparent 72%)',
            'radial-gradient(ellipse 44% 38% at 84% 20%, oklch(0.74 0.14 205 / 56%) 0%, transparent 72%)',
            'radial-gradient(ellipse 42% 36% at 62% 88%, oklch(0.72 0.15 290 / 28%) 0%, transparent 72%)',
          ].join(', '),
        }}
      />
      <div
        aria-hidden
        className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_72%_56%_at_50%_30%,black_24%,transparent_100%)] bg-[size:4.5rem_4.5rem] opacity-[0.08]'
      />

      <div className='mx-auto grid max-w-6xl gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,480px)] lg:items-center'>
        <div className='max-w-2xl'>
          <div
            className='landing-animate-fade-up border-border/50 bg-background/70 text-muted-foreground inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium opacity-0 backdrop-blur-sm'
            style={{ animationDelay: '0ms' }}
          >
            {t('Video Generation Workflow')}
          </div>
          <h1
            className='landing-animate-fade-up mt-5 text-[clamp(2.35rem,6vw,4.5rem)] leading-[1.05] font-bold tracking-tight opacity-0'
            style={{ animationDelay: '80ms' }}
          >
            {t('Video workflows')}
            <br />
            <span className='bg-gradient-to-r from-blue-500 via-violet-500 to-fuchsia-500 bg-clip-text text-transparent'>
              {t('without the clutter')}
            </span>
          </h1>
          <p
            className='landing-animate-fade-up text-muted-foreground/85 mt-6 max-w-xl text-base leading-relaxed opacity-0 md:text-lg'
            style={{ animationDelay: '160ms' }}
          >
            {t(
              'Route requests, track tasks, and keep pricing visible from one video ops console.'
            )}
          </p>
          <div
            className='landing-animate-fade-up mt-6 flex flex-wrap gap-2 opacity-0'
            style={{ animationDelay: '220ms' }}
          >
            {workflows.map((workflow) => (
              <span
                key={workflow}
                className='border-border/50 bg-background/80 text-foreground/80 rounded-full border px-3 py-1 text-xs font-medium backdrop-blur-sm'
              >
                {workflow}
              </span>
            ))}
          </div>
          <div
            className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3 opacity-0'
            style={{ animationDelay: '300ms' }}
          >
            {props.isAuthenticated ? (
              <Button
                className='group rounded-lg'
                render={<Link to='/dashboard' />}
              >
                {t('Go to Dashboard')}
                <ArrowRight className='ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5' />
              </Button>
            ) : (
              <>
                <Button
                  className='group rounded-lg'
                  render={<Link to='/sign-up' />}
                >
                  {t('Get Started')}
                  <ArrowRight className='ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
                <Button
                  variant='outline'
                  className='border-border/50 hover:border-border hover:bg-muted/50 rounded-lg'
                  render={<Link to='/pricing' />}
                >
                  {t('View Pricing')}
                </Button>
              </>
            )}
          </div>
          <p
            className='landing-animate-fade-up text-muted-foreground mt-5 max-w-lg text-sm leading-relaxed opacity-0'
            style={{ animationDelay: '380ms' }}
          >
            {t('Built for video tools and generation pipelines.')}
          </p>
        </div>
        <div
          className='landing-animate-fade-left w-full opacity-0'
          style={{ animationDelay: '280ms' }}
        >
          <HeroTerminalDemo />
        </div>
      </div>
    </section>
  )
}
