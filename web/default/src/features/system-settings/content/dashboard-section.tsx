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
import * as z from 'zod'
import type { Resolver } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { SettingsSection } from '../components/settings-section'
import { useSettingsForm } from '../hooks/use-settings-form'
import { useUpdateOption } from '../hooks/use-update-option'

const dataDashboardSchema = z.object({
  console_setting: z.object({
    dashboard_overview_enabled: z.boolean(),
  }),
  DataExportEnabled: z.boolean(),
  DataExportInterval: z.coerce.number().int().min(1).max(1440),
  DataExportDefaultTime: z.enum(['hour', 'day', 'week']),
})

type DataDashboardFormValues = z.infer<typeof dataDashboardSchema>

type DashboardSectionProps = {
  defaultValues: DataDashboardFormValues
}

const granularityOptions = [
  { label: 'Hour', value: 'hour' },
  { label: 'Day', value: 'day' },
  { label: 'Week', value: 'week' },
]

export function DashboardSection({ defaultValues }: DashboardSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const { form, handleSubmit, isSubmitting } =
    useSettingsForm<DataDashboardFormValues>({
      resolver: zodResolver(dataDashboardSchema) as Resolver<
        DataDashboardFormValues,
        unknown,
        DataDashboardFormValues
      >,
      defaultValues,
      onSubmit: async (_data, changedFields) => {
        for (const [key, value] of Object.entries(changedFields)) {
          if (value === undefined || value === null) continue

          let serialized: string | boolean = value as string | boolean
          if (typeof value === 'boolean') {
            serialized = String(value)
          } else if (typeof value === 'number') {
            serialized = Number.isFinite(value) ? String(value) : '0'
          }

          await updateOption.mutateAsync({ key, value: serialized })
        }
      },
    })

  const isEnabled = form.watch('DataExportEnabled')

  return (
    <SettingsSection
      title={t('Data Dashboard')}
      description={t('Configure experimental data export for the dashboard')}
    >
      <Form {...form}>
        <form onSubmit={handleSubmit} className='space-y-6'>
          <FormField
            control={form.control}
            name='console_setting.dashboard_overview_enabled'
            render={({ field }) => (
              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                <div className='space-y-1'>
                  <FormLabel className='text-base'>
                    {t('Show console overview page')}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      'Hide the standalone overview entry and redirect users to the next available dashboard tab.'
                    )}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
 
          <FormField
            control={form.control}
            name='DataExportEnabled'
            render={({ field }) => (
              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <FormLabel className='text-base'>
                    {t('Enable Data Dashboard')}
                  </FormLabel>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
 
          <div className='grid gap-6 sm:grid-cols-2'>
            <FormField
              control={form.control}
              name='DataExportInterval'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Refresh interval (minutes)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={1440}
                      step={1}
                      disabled={!isEnabled}
                      value={field.value}
                      onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Keep this above 1 minute to avoid heavy database load')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
 
            <FormField
              control={form.control}
              name='DataExportDefaultTime'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Default time granularity')}</FormLabel>
                  <Select
                    items={[
                      ...granularityOptions.map((option) => ({
                        value: option.value,
                        label: option.label,
                      })),
                    ]}
                    onValueChange={field.onChange}
                    value={field.value}
                    disabled={!isEnabled}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder={t('Select granularity')} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {granularityOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {t(
                      'UI granularity only &mdash; data is still aggregated hourly'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
 
          <Button
            type='submit'
            disabled={updateOption.isPending || isSubmitting}
          >
            {updateOption.isPending || isSubmitting
              ? t('Saving...')
              : t('Save Changes')}
          </Button>
        </form>
      </Form>
    </SettingsSection>
  )
}
