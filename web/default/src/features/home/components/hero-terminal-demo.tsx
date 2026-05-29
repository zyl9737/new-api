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
import {
  CheckCircle2,
  ListChecks,
  Play,
  Route,
  Sparkles,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

export function HeroTerminalDemo() {
  const { t } = useTranslation()

  const workflowModes = [
    t('Text to Video'),
    t('Image to Video'),
    t('Video Remix'),
  ]

  const pipeline = [
    {
      icon: <Sparkles className='size-3.5 text-violet-500' />,
      label: t('Submit'),
      route: '/v1/video/generations',
    },
    {
      icon: <ListChecks className='size-3.5 text-blue-500' />,
      label: t('Poll'),
      route: '/v1/videos/{task_id}',
    },
    {
      icon: <Play className='size-3.5 text-emerald-500' />,
      label: t('Preview output'),
      route: '/v1/videos/{task_id}/content',
    },
  ]

  const signals = [
    {
      icon: <Route className='size-3.5 text-blue-500' />,
      title: t('Model route'),
      value: 'Kling / Veo / Pika',
    },
    {
      icon: <ListChecks className='size-3.5 text-violet-500' />,
      title: t('Task state'),
      value: t('Queued -> Rendering -> Ready'),
    },
  ]

  return (
    <div className='mx-auto w-full max-w-[480px]'>
      <div className='border-border/60 bg-background/92 overflow-hidden rounded-3xl border shadow-[0_24px_80px_-36px_rgba(15,23,42,0.35)] backdrop-blur-xl'>
        <div className='border-border/50 flex items-center justify-between border-b px-5 py-4'>
          <div>
            <p className='text-muted-foreground text-[11px] font-semibold tracking-[0.18em] uppercase'>
              {t('Live request')}
            </p>
            <h3 className='mt-1 text-sm font-semibold'>
              {t('Video generation')}
            </h3>
          </div>
          <div className='inline-flex items-center gap-2 rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-600 dark:text-emerald-400'>
            <CheckCircle2 className='size-3.5' />
            {t('Task ready')}
          </div>
        </div>

        <div className='grid gap-4 p-5'>
          <div className='border-border/50 bg-muted/20 rounded-2xl border p-4'>
            <div className='flex items-start justify-between gap-3'>
              <div>
                <span className='inline-flex rounded-md bg-violet-500/10 px-1.5 py-0.5 font-mono text-[10px] font-semibold tracking-wider text-violet-600 dark:text-violet-400'>
                  POST
                </span>
                <code className='text-foreground/85 mt-2 block font-mono text-sm'>
                  /v1/video/generations
                </code>
              </div>
              <span className='border-border/50 bg-background text-muted-foreground rounded-full border px-2.5 py-1 text-[11px] font-medium'>
                kling-v2
              </span>
            </div>

            <div className='bg-background/80 mt-4 space-y-2 rounded-xl p-3 font-mono text-xs'>
              <div className='flex items-center justify-between gap-3'>
                <span className='text-muted-foreground'>"prompt"</span>
                <span className='text-foreground/80 truncate text-right'>
                  "night drive, rain"
                </span>
              </div>
              <div className='flex items-center justify-between gap-3'>
                <span className='text-muted-foreground'>"mode"</span>
                <span className='text-foreground/80'>"text2video"</span>
              </div>
              <div className='flex items-center justify-between gap-3'>
                <span className='text-muted-foreground'>"duration"</span>
                <span className='text-foreground/80'>"8s"</span>
              </div>
            </div>

            <div className='mt-3 flex flex-wrap gap-2'>
              {workflowModes.map((mode) => (
                <span
                  key={mode}
                  className='border-border/50 bg-background text-muted-foreground rounded-full border px-2.5 py-1 text-[11px] font-medium'
                >
                  {mode}
                </span>
              ))}
            </div>
          </div>

          <div className='grid gap-4 lg:grid-cols-[minmax(0,1fr)_180px]'>
            <div className='border-border/50 bg-muted/20 rounded-2xl border p-4'>
              <p className='text-muted-foreground text-[11px] font-semibold tracking-[0.18em] uppercase'>
                {t('Request flow')}
              </p>
              <div className='mt-4 space-y-3'>
                {pipeline.map((step, index) => (
                  <div
                    key={step.route}
                    className='bg-background/80 flex items-center gap-3 rounded-xl px-3 py-2.5'
                  >
                    <div className='bg-muted border-border/50 flex size-8 shrink-0 items-center justify-center rounded-full border'>
                      {step.icon}
                    </div>
                    <div className='min-w-0'>
                      <p className='text-sm font-medium'>{step.label}</p>
                      <code className='text-muted-foreground block truncate font-mono text-[11px]'>
                        {step.route}
                      </code>
                    </div>
                    <span className='text-muted-foreground ml-auto text-xs font-medium'>
                      0{index + 1}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className='grid gap-3'>
              {signals.map((signal) => (
                <div
                  key={signal.title}
                  className='border-border/50 bg-muted/20 rounded-2xl border p-4'
                >
                  <div className='border-border/50 bg-background mb-3 flex size-8 items-center justify-center rounded-full border'>
                    {signal.icon}
                  </div>
                  <p className='text-sm font-medium'>{signal.title}</p>
                  <p className='text-muted-foreground mt-1 text-xs leading-relaxed'>
                    {signal.value}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
