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
import { describe, expect, it } from 'vitest'

import { formatCurrencyFromUSD } from '@/lib/currency'

import { formatChannelBalance, isKimiCodingPlanChannel } from '../channel-utils'

const kimiPlanChannel = {
  type: 58,
  base_url: 'https://api.kimi.com/coding',
}

describe('isKimiCodingPlanChannel', () => {
  it('matches advanced custom channels pointing at the kimi coding host', () => {
    expect(isKimiCodingPlanChannel(kimiPlanChannel)).toBe(true)
  })

  it('matches the host case-insensitively and with a /v1 suffix', () => {
    expect(
      isKimiCodingPlanChannel({
        type: 58,
        base_url: 'https://API.KIMI.COM/coding/v1',
      })
    ).toBe(true)
  })

  it('rejects advanced custom channels on other hosts', () => {
    expect(
      isKimiCodingPlanChannel({ type: 58, base_url: 'https://api.deepseek.com' })
    ).toBe(false)
  })

  it('rejects non-advanced-custom channels on the kimi coding host', () => {
    expect(
      isKimiCodingPlanChannel({ type: 43, base_url: 'https://api.kimi.com/coding' })
    ).toBe(false)
  })

  it('rejects lookalike hosts that only share the suffix', () => {
    expect(
      isKimiCodingPlanChannel({
        type: 58,
        base_url: 'https://api.kimi.com.evil.net/coding',
      })
    ).toBe(false)
  })

  it('rejects missing base_url', () => {
    expect(isKimiCodingPlanChannel({ type: 58, base_url: null })).toBe(false)
    expect(isKimiCodingPlanChannel({ type: 58 })).toBe(false)
  })
})

describe('formatChannelBalance', () => {
  it('renders kimi coding plan balance as a weekly quota percentage', () => {
    expect(formatChannelBalance(kimiPlanChannel, 63)).toBe('63%')
  })

  it('renders an exhausted plan as 0%', () => {
    expect(formatChannelBalance(kimiPlanChannel, 0)).toBe('0%')
  })

  it('renders missing balance as a placeholder', () => {
    expect(formatChannelBalance(kimiPlanChannel, null)).toBe('-')
  })

  it('falls back to currency formatting for other channels', () => {
    const deepseekChannel = { type: 43, base_url: 'https://api.deepseek.com' }
    expect(formatChannelBalance(deepseekChannel, 63)).toBe(
      formatCurrencyFromUSD(63)
    )
  })
})
