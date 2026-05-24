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
import { useEffect, useState } from 'react'
import { Loader2, ExternalLink, Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'

interface VideoPreviewDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  videoUrl: string
  taskId?: string
}

export function VideoPreviewDialog({
  open,
  onOpenChange,
  videoUrl,
  taskId,
}: VideoPreviewDialogProps) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard({ notify: true })
  const [videoError, setVideoError] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (open) {
      setVideoError(false)
      setIsLoading(true)
    }
  }, [open, videoUrl])

  const openInNewTab = () => {
    window.open(videoUrl, '_blank', 'noopener,noreferrer')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[85vh] sm:max-w-5xl'>
        <DialogHeader>
          <DialogTitle>{t('Click to preview video')}</DialogTitle>
          <DialogDescription>
            {taskId ? `${t('Task ID:')} ${taskId}` : t('Details')}
          </DialogDescription>
        </DialogHeader>

        <div className='bg-muted/40 relative h-[65vh] overflow-hidden rounded-lg border'>
          {videoError ? (
            <div className='flex h-full flex-col items-center justify-center gap-3 p-6 text-center'>
              <p className='text-muted-foreground text-sm'>
                {t('Failed to load video')}
              </p>
              <div className='flex flex-wrap items-center justify-center gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  className='h-8 gap-1 text-xs'
                  onClick={openInNewTab}
                >
                  <ExternalLink className='size-3.5' />
                  {t('Open in new tab')}
                </Button>
                <Button
                  variant='outline'
                  size='sm'
                  className='h-8 gap-1 text-xs'
                  onClick={() => copyToClipboard(videoUrl)}
                >
                  <Copy className='size-3.5' />
                  {t('Copy Link')}
                </Button>
              </div>
              <p className='text-muted-foreground max-w-full text-[11px] break-all'>
                {videoUrl}
              </p>
            </div>
          ) : (
            <>
              {isLoading && (
                <div className='bg-background/50 absolute inset-0 z-10 flex items-center justify-center'>
                  <Loader2 className='text-muted-foreground size-5 animate-spin' />
                </div>
              )}
              <video
                src={videoUrl}
                controls
                className='h-full w-full object-contain'
                onError={() => {
                  setVideoError(true)
                  setIsLoading(false)
                }}
                onLoadedData={() => setIsLoading(false)}
                onLoadStart={() => setIsLoading(true)}
              />
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
