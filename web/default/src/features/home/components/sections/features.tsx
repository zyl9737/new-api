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
import { ListChecks, Route, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      title: t('Stable video routes'),
      desc: t(
        'Submit, poll, and preview through one consistent API surface.'
      ),
      icon: <Route className='size-5 text-blue-500' />,
      items: [t('Submit endpoint'), t('Status endpoint'), t('Preview endpoint')],
    },
    {
      title: t('Status and preview'),
      desc: t(
        'Keep task IDs, states, and outputs easy to follow.'
      ),
      icon: <ListChecks className='size-5 text-emerald-500' />,
      items: [t('Queued'), t('Rendering'), t('Preview ready'), t('Task log')],
    },
    {
      title: t('Spend control'),
      desc: t(
        'Keep routing, ratios, and usage readable as volume grows.'
      ),
      icon: <WalletCards className='size-5 text-amber-500' />,
      items: [
        t('Provider routing'),
        t('Budget guard'),
        t('Usage and pricing stay visible'),
      ],
    },
  ]

  return (
    <section className='relative z-10 px-6 py-20 md:py-24'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-12 max-w-2xl'>
          <p className='text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase'>
            {t('Core Features')}
          </p>
          <h2 className='text-2xl leading-tight font-bold tracking-tight md:text-3xl'>
            {t('Built for video generation')}
          </h2>
          <p className='text-muted-foreground mt-4 max-w-xl text-sm leading-relaxed md:text-base'>
            {t('Routes, task status, and spend control in one place.')}
          </p>
        </AnimateInView>

        <div className='grid gap-4 md:grid-cols-3'>
          {features.map((feature, index) => (
            <AnimateInView
              key={feature.title}
              delay={index * 80}
              animation='scale-in'
              className='border-border/50 bg-background hover:bg-muted/20 rounded-2xl border p-6 transition-colors duration-300 md:p-7'
            >
              <div className='mb-4 flex items-center gap-3'>
                <div className='bg-muted border-border/50 flex size-10 items-center justify-center rounded-xl border'>
                  {feature.icon}
                </div>
                <h3 className='text-base font-semibold'>{feature.title}</h3>
              </div>
              <p className='text-muted-foreground text-sm leading-relaxed'>
                {feature.desc}
              </p>
              <div className='mt-5 flex flex-wrap gap-2'>
                {feature.items.map((item) => (
                  <span
                    key={item}
                    className='border-border/50 bg-muted/40 text-muted-foreground rounded-full border px-3 py-1 text-xs'
                  >
                    {item}
                  </span>
                ))}
              </div>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
