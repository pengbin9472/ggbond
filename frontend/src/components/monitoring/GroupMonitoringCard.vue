<template>
  <div
    class="group-card"
    :class="{ 'group-card-hover': true }"
  >
    <!-- 分组名称和状态标签 -->
    <div class="flex items-center mb-1" style="gap: 8px">
      <div
        class="flex-1 min-w-0 text-[15px] font-semibold text-gray-900 dark:text-white truncate"
        :title="group.group_name"
      >
        {{ group.group_name }}
      </div>
      <span
        class="inline-flex items-center gap-1.5 text-[11px] font-semibold px-2.5 py-0.5 rounded-full flex-shrink-0"
        :class="isOnline
          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
          : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'"
      >
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="isOnline ? 'bg-green-500' : 'bg-red-500'"
        ></span>
        {{ isOnline ? t('monitoring.online') : t('monitoring.offline') }}
      </span>
    </div>

    <!-- 平台 | 倍率 -->
    <div class="flex items-center mb-3.5 text-xs text-gray-500 dark:text-gray-400" style="gap: 8px">
      <span class="truncate min-w-0 flex-1">{{ getPlatformLabel(group.platform) }}</span>
      <template v-if="group.rate_multiplier !== 1">
        <span class="w-px h-3 bg-gray-300 dark:bg-gray-600 flex-shrink-0"></span>
        <span class="font-semibold flex-shrink-0">{{ group.rate_multiplier }}x</span>
      </template>
    </div>

    <!-- 可用率进度条 -->
    <div class="mb-3">
      <div class="flex items-center justify-between mb-1.5">
        <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
          {{ t('monitoring.availabilityRate') }}
        </span>
        <span
          class="text-[13px] font-semibold tabular-nums"
          :style="{ color: getRateColor(group.availability_rate) }"
        >
          {{ formatRate(group.availability_rate) }}
        </span>
      </div>
      <div class="h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all duration-500 ease-out"
          :style="{
            width: group.availability_rate >= 0 ? `${group.availability_rate}%` : '0%',
            backgroundColor: getRateColor(group.availability_rate)
          }"
        ></div>
      </div>
    </div>

    <!-- 缓存命中率进度条 -->
    <div class="mb-0">
      <div class="flex items-center justify-between mb-1.5">
        <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
          {{ t('monitoring.cacheHitRate') }}
        </span>
        <span
          class="text-[13px] font-semibold tabular-nums"
          :style="{ color: getRateColor(group.cache_hit_rate) }"
        >
          {{ formatCacheRate(group.cache_hit_rate) }}
        </span>
      </div>
      <div class="h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all duration-500 ease-out"
          :style="{
            width: group.cache_hit_rate >= 0 ? `${group.cache_hit_rate}%` : '0%',
            backgroundColor: getRateColor(group.cache_hit_rate)
          }"
        ></div>
      </div>
    </div>

    <!-- 迷你历史图表 -->
    <div
      class="mt-3.5 pt-3 border-t border-gray-200 dark:border-gray-700 cursor-pointer"
      @click="onToggleHistory"
    >
      <template v-if="history && history.length > 1">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[11px] font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
            HISTORY
          </span>
          <div class="flex items-center gap-2.5">
            <span class="flex items-center gap-1 text-[10px] text-gray-500 dark:text-gray-400">
              <span class="inline-block w-2 h-0.5 rounded bg-blue-500"></span>
              {{ t('monitoring.availabilityRate') }}
            </span>
            <span class="flex items-center gap-1 text-[10px] text-gray-500 dark:text-gray-400">
              <span class="inline-block w-2 h-0.5 rounded bg-green-500"></span>
              {{ t('monitoring.cacheHitRate') }}
            </span>
          </div>
        </div>
        <svg :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="w-full" style="height: 40px">
          <!-- 可用率折线 -->
          <polyline
            :points="availabilityPoints"
            fill="none"
            stroke="#3b82f6"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <!-- 缓存命中率折线 -->
          <polyline
            :points="cacheHitPoints"
            fill="none"
            stroke="#22c55e"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </template>
      <template v-else-if="historyLoading">
        <div class="flex items-center justify-center py-2">
          <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ t('common.loading') }}...</span>
        </div>
      </template>
      <template v-else>
        <div class="flex items-center justify-center py-1">
          <span class="text-[11px] text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
            {{ t('monitoring.showHistory') }}
          </span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GroupMonitoringStat, MonitoringHistoryPoint } from '@/types'

const { t } = useI18n()

interface Props {
  group: GroupMonitoringStat
  history?: MonitoringHistoryPoint[]
}

const props = defineProps<Props>()
const emit = defineEmits<{ 'load-history': [] }>()

const historyLoading = ref(false)

const onToggleHistory = async () => {
  if (props.history) return
  historyLoading.value = true
  emit('load-history')
  // loading 状态会在 history prop 更新后通过 watch 自动消失
  // 但为安全起见设置一个超时
  setTimeout(() => { historyLoading.value = false }, 5000)
}

// 当 history 数据到达时关闭 loading
import { watch } from 'vue'
watch(() => props.history, (val) => {
  if (val) historyLoading.value = false
})

const chartWidth = 200
const chartHeight = 40

// 是否在线（有正常账户即为在线）
const isOnline = computed(() => {
  return props.group.normal_accounts > 0
})

// 生成 SVG 折线的 points 字符串（自适应 Y 轴）
const makePoints = (data: MonitoringHistoryPoint[], field: 'availability_rate' | 'cache_hit_rate'): string => {
  const validData = data.filter(d => d[field] >= 0)
  if (validData.length < 2) return ''

  const values = validData.map(d => d[field])
  let min = Math.min(...values)
  let max = Math.max(...values)

  // 数据无波动时，上下扩展 5% 的范围，确保线画在中间
  if (max - min < 1) {
    min = Math.max(0, min - 5)
    max = Math.min(100, max + 5)
  }

  // 上下留 padding（图表高度的 10%）
  const padding = chartHeight * 0.1
  const plotHeight = chartHeight - padding * 2
  const range = max - min || 1

  const xStep = chartWidth / (validData.length - 1)
  return validData
    .map((d, i) => {
      const x = i * xStep
      const y = padding + plotHeight - ((d[field] - min) / range) * plotHeight
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
}

const availabilityPoints = computed(() => {
  if (!props.history) return ''
  return makePoints(props.history, 'availability_rate')
})

const cacheHitPoints = computed(() => {
  if (!props.history) return ''
  return makePoints(props.history, 'cache_hit_rate')
})

const getPlatformLabel = (platform: string): string => {
  const platformMap: Record<string, string> = {
    anthropic: 'Claude',
    openai: 'OpenAI',
    gemini: 'Gemini',
    antigravity: 'Antigravity'
  }
  return platformMap[platform] || platform
}

const formatRate = (rate: number): string => {
  if (rate == null || isNaN(rate) || rate < 0) return '--'
  return rate.toFixed(1) + '%'
}

const formatCacheRate = (rate: number): string => {
  if (rate == null || isNaN(rate) || rate < 0) return t('monitoring.notCollected')
  return rate.toFixed(1) + '%'
}

const getRateColor = (rate: number): string => {
  if (rate < 0) return '#9ca3af' // gray
  if (rate >= 95) return '#22c55e' // green
  if (rate >= 90) return '#65a30d' // lime
  if (rate >= 85) return '#f59e0b' // amber
  if (rate >= 75) return '#ea580c' // orange
  return '#ef4444' // red
}
</script>

<style scoped>
.group-card {
  @apply bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 0 1px 3px rgba(0, 0, 0, 0.06);
  transition: all 0.25s ease;
}

.group-card-hover:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12), 0 2px 6px rgba(0, 0, 0, 0.06);
  transform: translateY(-1px);
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
