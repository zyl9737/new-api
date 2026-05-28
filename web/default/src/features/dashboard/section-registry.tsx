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
import type { TFunction } from 'i18next'
import { createSectionRegistry } from '@/features/system-settings/utils/section-registry'

/**
 * Dashboard page section definitions
 */
const DASHBOARD_SECTIONS = [
  {
    id: 'overview',
    titleKey: 'Overview',
    descriptionKey: 'View dashboard overview and statistics',
    build: () => null,
  },
  {
    id: 'models',
    titleKey: 'Model Call Analytics',
    descriptionKey: 'View model call count analytics and charts',
    build: () => null,
  },
  {
    id: 'users',
    titleKey: 'User Analytics',
    descriptionKey: 'View user consumption statistics and charts',
    adminOnly: true,
    build: () => null,
  },
] as const

export type DashboardSectionId = (typeof DASHBOARD_SECTIONS)[number]['id']

const ADMIN_ONLY_SECTIONS = new Set<string>(['users'])

const dashboardRegistry = createSectionRegistry<
  DashboardSectionId,
  Record<string, never>,
  []
>({
  sections: DASHBOARD_SECTIONS,
  defaultSection: 'overview',
  basePath: '/dashboard',
  urlStyle: 'path',
})

export const DASHBOARD_SECTION_IDS = dashboardRegistry.sectionIds
export const DASHBOARD_DEFAULT_SECTION = dashboardRegistry.defaultSection

type DashboardSectionOptions = {
  isAdmin?: boolean
  overviewEnabled?: boolean
}

export function getVisibleDashboardSections(
  options: DashboardSectionOptions = {}
) {
  const { isAdmin = false, overviewEnabled = true } = options

  return DASHBOARD_SECTION_IDS.filter((section) => {
    if (!overviewEnabled && section === 'overview') return false
    if (!isAdmin && ADMIN_ONLY_SECTIONS.has(section)) return false
    return true
  })
}

export function getDashboardDefaultSection(
  options: DashboardSectionOptions = {}
) {
  return getVisibleDashboardSections(options)[0] ?? DASHBOARD_DEFAULT_SECTION
}

export function getDashboardSectionNavItems(
  t: TFunction,
  options?: DashboardSectionOptions
) {
  const all = dashboardRegistry.getSectionNavItems(t)
  const visible = new Set(getVisibleDashboardSections(options))
  return all.filter((_, idx) => visible.has(DASHBOARD_SECTIONS[idx].id))
}
