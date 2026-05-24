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
import { useMemo } from 'react'
import { Copy, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { TaskLog } from '../../types'

interface TaskLogDetailDialogProps {
  log: TaskLog
  open: boolean
  onOpenChange: (open: boolean) => void
}

function buildTaskDetailText(log: TaskLog): string {
  const detail: Record<string, unknown> = { ...log }
  if (typeof detail.data === 'string') {
    try {
      detail.data = JSON.parse(detail.data)
    } catch {
      // Keep original text when parsing fails.
    }
  }
  return JSON.stringify(detail, null, 2)
}

export function TaskLogDetailDialog({
  log,
  open,
  onOpenChange,
}: TaskLogDetailDialogProps) {
  const { t } = useTranslation()
  const { copiedText, copyToClipboard } = useCopyToClipboard({ notify: false })
  const detailText = useMemo(() => buildTaskDetailText(log), [log])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Log Details')}</DialogTitle>
          <DialogDescription>
            {t('View the complete details for this log entry')}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className='max-h-[65vh] pr-4'>
          <div className='space-y-4 py-4'>
            <div className='space-y-2'>
              <Label className='text-sm font-semibold'>{t('Details')}</Label>
              <div className='bg-muted/50 relative rounded-md border p-3'>
                <Button
                  variant='ghost'
                  size='sm'
                  className='absolute top-2 right-2 h-8 w-8 p-0'
                  onClick={() => copyToClipboard(detailText)}
                  title={t('Copy to clipboard')}
                >
                  {copiedText === detailText ? (
                    <Check className='size-4 text-green-600' />
                  ) : (
                    <Copy className='size-4' />
                  )}
                </Button>
                <pre className='overflow-x-auto pr-10 text-xs leading-relaxed whitespace-pre-wrap break-all'>
                  {detailText}
                </pre>
              </div>
            </div>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
