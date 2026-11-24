import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import type { DashboardStats } from '@/types'
import DashboardView from '../DashboardView.vue'

const { getSnapshotV2, getModelStats, getUserUsageTrend, getUserSpendingRanking } = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getModelStats: vi.fn(),
  getUserUsageTrend: vi.fn(),
  getUserSpendingRanking: vi.fn()
}))

vi.mock('vue-chartjs', () => ({
  Line: { name: 'Line', props: ['data', 'options'], template: '<div class="line-chart" />' }
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getModelStats,
      getUserUsageTrend,
      getUserSpendingRanking
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const createDashboardStats = (): DashboardStats => ({
  total_users: 0,
  today_new_users: 0,
  active_users: 0,
  hourly_active_users: 0,
  stats_updated_at: '',
  stats_stale: false,
  total_api_keys: 0,
  active_api_keys: 0,
  total_accounts: 0,
  normal_accounts: 0,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  average_duration_ms: 0,
  uptime: 0,
  rpm: 0,
  tpm: 0
})

describe('admin DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())

    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getUserUsageTrend.mockReset()
    getUserSpendingRanking.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [],
      models: []
    })
    getModelStats.mockResolvedValue({ models: [] })
    getUserUsageTrend.mockResolvedValue({
      trend: [],
      start_date: '',
      end_date: '',
      granularity: 'hour'
    })
    getUserSpendingRanking.mockResolvedValue({
      ranking: [],
      total_actual_cost: 0,
      total_requests: 0,
      total_tokens: 0,
      start_date: '',
      end_date: ''
    })
  })

  it('switches metrics and ignores a stale tokens response', async () => {
    let resolveTokens!: (value: any) => void
    getUserUsageTrend.mockImplementationOnce(() => new Promise(resolve => { resolveTokens = resolve }))
    getUserUsageTrend.mockResolvedValueOnce({ trend: [{ date: '2026-01-01', user_id: 2, username: 'spender', email: '', requests: 1, tokens: 10, cost: 5, actual_cost: 5 }] })
    const wrapper = mount(DashboardView, {
      global: { stubs: {
        AppLayout: { template: '<div><slot /></div>' }, LoadingSpinner: true,
        Icon: true, DateRangePicker: true, Select: true, ModelDistributionChart: true,
        TokenUsageTrend: true
      } }
    })
    await flushPromises()
    expect(getUserUsageTrend).toHaveBeenCalledWith(expect.objectContaining({ metric: 'tokens', limit: 12 }), expect.any(AbortSignal))
    await wrapper.findAll('button').find(button => button.text() === 'admin.dashboard.actualSpending')!.trigger('click')
    await flushPromises()
    expect(getUserUsageTrend).toHaveBeenCalledWith(expect.objectContaining({ metric: 'actual_cost', limit: 12 }), expect.any(AbortSignal))
    const chart = wrapper.findComponent({ name: 'Line' })
    expect(chart.props('data').datasets[0].data).toEqual([5])
    expect(chart.props('options').scales.y.ticks.callback(5)).toBe('$5.00')
    resolveTokens({ trend: [{ date: '2026-01-01', user_id: 1, username: 'tokens', email: '', requests: 1, tokens: 100, cost: 1, actual_cost: 1 }] })
    await flushPromises()
    expect(chart.props('data').datasets[0].data).toEqual([5])
  })

  it('uses today as the default dashboard range', async () => {
    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          Line: true
        }
      }
    })

    await flushPromises()

    const now = new Date()

    expect(getSnapshotV2).toHaveBeenCalledTimes(2)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      granularity: 'hour',
      include_stats: false,
      include_model_stats: false
    }), expect.any(AbortSignal))
  })
  it('shows statistics and the date filter while the model query is pending', async () => {
    getModelStats.mockImplementationOnce(() => new Promise(() => {}))
    const wrapper = mount(DashboardView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, LoadingSpinner: true,
      Icon: true, DateRangePicker: { template: '<div class="date-picker" />' }, Select: true,
      ModelDistributionChart: true, TokenUsageTrend: true, Line: true
    } } })
    await flushPromises()
    expect(wrapper.find('.date-picker').exists()).toBe(true)
    expect(wrapper.text()).toContain('admin.dashboard.apiKeys')
    expect(getUserUsageTrend).not.toHaveBeenCalled()
    expect(getUserSpendingRanking).not.toHaveBeenCalled()
    wrapper.unmount()
    expect(getModelStats.mock.calls[0][1].aborted).toBe(true)
  })

  it('keeps the date filter available when statistics fail', async () => {
    getSnapshotV2.mockRejectedValueOnce(new Error('timeout'))
    const wrapper = mount(DashboardView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, LoadingSpinner: true,
      Icon: true, DateRangePicker: { template: '<div class="date-picker" />' }, Select: true,
      ModelDistributionChart: true, TokenUsageTrend: true, Line: true
    } } })
    await flushPromises()
    expect(wrapper.find('.date-picker').exists()).toBe(true)
    expect(wrapper.text()).toContain('admin.dashboard.failedToLoad')
    expect(getUserSpendingRanking).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('reports failed model and user trend queries and continues to ranking', async () => {
    getModelStats.mockRejectedValueOnce(new Error('model timeout'))
    getUserUsageTrend.mockRejectedValueOnce(new Error('trend timeout'))
    const wrapper = mount(DashboardView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, LoadingSpinner: true,
      Icon: true, DateRangePicker: true, Select: true,
      ModelDistributionChart: { props: ['error'], template: '<div class="model-chart">{{ error }}</div>' },
      TokenUsageTrend: true, Line: true
    } } })
    await flushPromises()
    expect(wrapper.find('.model-chart').text()).toBe('true')
    expect(wrapper.findAll('[role="alert"]').length).toBeGreaterThan(0)
    expect(getUserSpendingRanking).toHaveBeenCalled()
    wrapper.unmount()
  })

})
