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
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='relative flex min-h-svh flex-col items-center justify-center overflow-x-clip px-4 py-10'>
      {/* Brand glow, echoes the landing page hero */}
      <div
        aria-hidden='true'
        className='pointer-events-none absolute inset-0 bg-[radial-gradient(60%_50%_at_50%_0%,rgba(99,102,241,0.16),transparent_70%)] dark:bg-[radial-gradient(60%_50%_at_50%_0%,rgba(99,102,241,0.22),transparent_70%)]'
      />
      <Link
        to='/'
        className='relative z-10 mb-8 flex flex-col items-center gap-3 transition-opacity hover:opacity-80'
      >
        <div className='relative h-14 w-14'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-2xl' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-14 w-14 rounded-2xl object-cover shadow-lg shadow-indigo-500/30'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-7 w-28' />
        ) : (
          <h1 className='bg-gradient-to-r from-indigo-500 to-violet-500 bg-clip-text text-2xl font-bold tracking-tight text-transparent'>
            {systemName}
          </h1>
        )}
      </Link>
      <div className='relative z-10 w-full sm:max-w-md'>
        <div className='bg-card rounded-2xl border p-6 shadow-xl shadow-indigo-500/5 sm:p-8'>
          {children}
        </div>
      </div>
    </div>
  )
}
