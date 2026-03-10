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
        :class="statusBadgeClass"
      >
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="statusDotClass"
        ></span>
        {{ statusLabel }}
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

    <div class="mb-3 text-[11px] text-gray-500 dark:text-gray-400">
      <span class="font-medium">{{ t('monitoring.lastProbe') }}</span>
      <span class="ml-1">{{ formatLastProbe() }}</span>
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

    <!-- 历史趋势图表 -->
    <div
      class="mt-3.5 pt-3 border-t border-gray-200 dark:border-gray-700"
      :class="{ 'cursor-pointer': !history }"
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

        <!-- 交互式图表容器 -->
        <div class="relative" ref="chartContainerRef">
          <svg
            ref="svgRef"
            :viewBox="`0 0 ${svgWidth} ${svgTotalHeight}`"
            class="w-full select-none"
            preserveAspectRatio="xMidYMid meet"
            @mousemove="onSvgMouseMove"
            @mouseleave="hoveredIndex = -1"
          >
            <!-- 水平网格线 -->
            <line
              v-for="i in 3"
              :key="'grid-' + i"
              :x1="0"
              :y1="svgPaddingTop + (svgPlotHeight / 4) * i"
              :x2="svgWidth"
              :y2="svgPaddingTop + (svgPlotHeight / 4) * i"
              stroke="currentColor"
              class="text-gray-100 dark:text-gray-700/50"
              stroke-width="0.5"
            />

            <!-- 可用率区域填充 -->
            <path
              v-if="availabilityAreaPath"
              :d="availabilityAreaPath"
              fill="#3b82f6"
              fill-opacity="0.06"
            />

            <!-- 缓存命中率区域填充 -->
            <path
              v-if="cacheHitAreaPath"
              :d="cacheHitAreaPath"
              fill="#22c55e"
              fill-opacity="0.06"
            />

            <!-- 可用率平滑曲线 -->
            <path
              v-if="availabilityLinePath"
              :d="availabilityLinePath"
              fill="none"
              stroke="#3b82f6"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />

            <!-- 缓存命中率平滑曲线 -->
            <path
              v-if="cacheHitLinePath"
              :d="cacheHitLinePath"
              fill="none"
              stroke="#22c55e"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />

            <!-- 悬浮竖直虚线 -->
            <line
              v-if="hoveredIndex >= 0"
              :x1="getPointX(hoveredIndex)"
              :y1="svgPaddingTop"
              :x2="getPointX(hoveredIndex)"
              :y2="svgPaddingTop + svgPlotHeight"
              stroke="currentColor"
              class="text-gray-300 dark:text-gray-500"
              stroke-width="0.5"
              stroke-dasharray="2,2"
            />

            <!-- 悬浮圆点指示器 -->
            <template v-if="hoveredIndex >= 0 && chartData">
              <circle
                v-if="chartData[hoveredIndex]?.availability_rate >= 0"
                :cx="getPointX(hoveredIndex)"
                :cy="computeY(chartData[hoveredIndex].availability_rate)"
                r="2.5"
                fill="white"
                stroke="#3b82f6"
                stroke-width="1.5"
              />
              <circle
                v-if="chartData[hoveredIndex]?.cache_hit_rate >= 0"
                :cx="getPointX(hoveredIndex)"
                :cy="computeY(chartData[hoveredIndex].cache_hit_rate)"
                r="2.5"
                fill="white"
                stroke="#22c55e"
                stroke-width="1.5"
              />
            </template>

            <!-- X 轴时间标签 -->
            <text
              v-for="label in xAxisLabels"
              :key="'t-' + label.index"
              :x="label.x"
              :y="svgTotalHeight - 1"
              text-anchor="middle"
              class="fill-gray-400 dark:fill-gray-500"
              style="font-size: 7px; font-family: system-ui, sans-serif"
            >{{ label.text }}</text>
          </svg>

          <!-- 悬浮 Tooltip -->
          <div
            v-if="hoveredIndex >= 0 && tooltipData"
            class="absolute z-10 pointer-events-none"
            :style="tooltipStyle"
          >
            <div class="bg-gray-800/95 dark:bg-black/90 backdrop-blur-sm text-white rounded-lg shadow-xl px-3 py-2 text-[11px] whitespace-nowrap border border-gray-700/50">
              <div class="font-medium text-gray-300 mb-1.5 text-[10px]">{{ tooltipData.time }}</div>
              <div class="flex items-center gap-1.5 mb-0.5">
                <span class="w-1.5 h-1.5 rounded-full bg-blue-400 flex-shrink-0"></span>
                <span class="text-gray-300">{{ t('monitoring.availabilityRate') }}</span>
                <span class="font-semibold ml-2">{{ tooltipData.availability }}</span>
              </div>
              <div class="flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-green-400 flex-shrink-0"></span>
                <span class="text-gray-300">{{ t('monitoring.cacheHitRate') }}</span>
                <span class="font-semibold ml-2">{{ tooltipData.cacheHit }}</span>
              </div>
            </div>
          </div>
        </div>
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
import { ref, computed, watch } from 'vue'
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
const hoveredIndex = ref(-1)
const chartContainerRef = ref<HTMLElement | null>(null)
const svgRef = ref<SVGSVGElement | null>(null)

const onToggleHistory = async () => {
  if (props.history) return
  historyLoading.value = true
  emit('load-history')
  setTimeout(() => { historyLoading.value = false }, 5000)
}

watch(() => props.history, (val) => {
  if (val) historyLoading.value = false
})

// ── SVG 图表尺寸 ──
const svgWidth = 300
const svgPaddingTop = 6
const svgPaddingBottom = 16 // 留给时间标签
const svgTotalHeight = 80
const svgPlotHeight = svgTotalHeight - svgPaddingTop - svgPaddingBottom

// ── 过滤有效数据（至少 availability_rate 或 cache_hit_rate >= 0）──
const chartData = computed(() => {
  if (!props.history || props.history.length < 2) return null
  return props.history
})

// ── Y 轴范围（共享的 min/max）──
const yRange = computed(() => {
  if (!chartData.value) return { min: 0, max: 100 }

  const allValues: number[] = []
  for (const d of chartData.value) {
    if (d.availability_rate >= 0) allValues.push(d.availability_rate)
    if (d.cache_hit_rate >= 0) allValues.push(d.cache_hit_rate)
  }

  if (allValues.length === 0) return { min: 0, max: 100 }

  let min = Math.min(...allValues)
  let max = Math.max(...allValues)

  // 数据无波动时上下扩展，确保曲线在中间
  if (max - min < 1) {
    min = Math.max(0, min - 5)
    max = Math.min(100, max + 5)
  }

  // 留一点 padding
  const padding = (max - min) * 0.1
  min = Math.max(0, min - padding)
  max = Math.min(100, max + padding)

  return { min, max }
})

// ── 坐标计算 ──
const getPointX = (index: number): number => {
  if (!chartData.value || chartData.value.length < 2) return 0
  return (index / (chartData.value.length - 1)) * svgWidth
}

const computeY = (value: number): number => {
  const { min, max } = yRange.value
  const range = max - min || 1
  return svgPaddingTop + svgPlotHeight - ((value - min) / range) * svgPlotHeight
}

// ── Catmull-Rom 样条曲线 ──
const smoothLine = (points: { x: number; y: number }[]): string => {
  if (points.length < 2) return ''
  if (points.length === 2) {
    return `M${points[0].x.toFixed(1)},${points[0].y.toFixed(1)}L${points[1].x.toFixed(1)},${points[1].y.toFixed(1)}`
  }

  let d = `M${points[0].x.toFixed(1)},${points[0].y.toFixed(1)}`
  const tension = 0.3

  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[Math.max(0, i - 1)]
    const p1 = points[i]
    const p2 = points[i + 1]
    const p3 = points[Math.min(points.length - 1, i + 2)]

    const cp1x = p1.x + (p2.x - p0.x) * tension / 3
    const cp1y = p1.y + (p2.y - p0.y) * tension / 3
    const cp2x = p2.x - (p3.x - p1.x) * tension / 3
    const cp2y = p2.y - (p3.y - p1.y) * tension / 3

    d += ` C${cp1x.toFixed(1)},${cp1y.toFixed(1)} ${cp2x.toFixed(1)},${cp2y.toFixed(1)} ${p2.x.toFixed(1)},${p2.y.toFixed(1)}`
  }

  return d
}

// ── 生成折线路径和区域填充路径 ──
const makeLinePath = (field: 'availability_rate' | 'cache_hit_rate'): string => {
  if (!chartData.value) return ''
  const points: { x: number; y: number }[] = []
  for (let i = 0; i < chartData.value.length; i++) {
    const val = chartData.value[i][field]
    if (val >= 0) {
      points.push({ x: getPointX(i), y: computeY(val) })
    }
  }
  return smoothLine(points)
}

const makeAreaPath = (field: 'availability_rate' | 'cache_hit_rate'): string => {
  if (!chartData.value) return ''
  const points: { x: number; y: number }[] = []
  for (let i = 0; i < chartData.value.length; i++) {
    const val = chartData.value[i][field]
    if (val >= 0) {
      points.push({ x: getPointX(i), y: computeY(val) })
    }
  }
  if (points.length < 2) return ''

  const linePath = smoothLine(points)
  const bottomY = svgPaddingTop + svgPlotHeight
  return `${linePath} L${points[points.length - 1].x.toFixed(1)},${bottomY} L${points[0].x.toFixed(1)},${bottomY} Z`
}

const availabilityLinePath = computed(() => makeLinePath('availability_rate'))
const cacheHitLinePath = computed(() => makeLinePath('cache_hit_rate'))
const availabilityAreaPath = computed(() => makeAreaPath('availability_rate'))
const cacheHitAreaPath = computed(() => makeAreaPath('cache_hit_rate'))

// ── X 轴时间标签 ──
const formatTimestamp = (ts: number): string => {
  // 兼容秒级和毫秒级时间戳
  const d = new Date(ts > 1e12 ? ts : ts * 1000)
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
}

const xAxisLabels = computed(() => {
  if (!chartData.value || chartData.value.length < 2) return []

  const total = chartData.value.length
  const labelCount = Math.min(5, total)
  const labels: { x: number; text: string; index: number }[] = []

  for (let i = 0; i < labelCount; i++) {
    const dataIndex = Math.round((i / (labelCount - 1)) * (total - 1))
    const point = chartData.value[dataIndex]
    if (point) {
      labels.push({
        x: getPointX(dataIndex),
        text: formatTimestamp(point.recorded_at),
        index: dataIndex
      })
    }
  }

  return labels
})

// ── 鼠标悬浮交互 ──
const onSvgMouseMove = (e: MouseEvent) => {
  const svg = svgRef.value
  if (!svg || !chartData.value || chartData.value.length < 2) return

  const rect = svg.getBoundingClientRect()
  const xRatio = (e.clientX - rect.left) / rect.width
  const svgX = xRatio * svgWidth

  const step = svgWidth / (chartData.value.length - 1)
  const idx = Math.round(svgX / step)
  hoveredIndex.value = Math.max(0, Math.min(chartData.value.length - 1, idx))
}

// ── Tooltip 数据 ──
const tooltipData = computed(() => {
  if (hoveredIndex.value < 0 || !chartData.value) return null
  const point = chartData.value[hoveredIndex.value]
  if (!point) return null

  return {
    time: formatTimestamp(point.recorded_at),
    availability: point.availability_rate >= 0 ? point.availability_rate.toFixed(1) + '%' : '--',
    cacheHit: point.cache_hit_rate >= 0 ? point.cache_hit_rate.toFixed(1) + '%' : '--'
  }
})

// ── Tooltip 定位样式 ──
const tooltipStyle = computed(() => {
  if (hoveredIndex.value < 0 || !chartData.value) return {}

  const total = chartData.value.length
  const xPercent = (hoveredIndex.value / (total - 1)) * 100

  // 超过 65% 时 tooltip 翻转到左侧显示
  const isRight = xPercent > 65
  return {
    left: `${xPercent}%`,
    top: '0px',
    transform: isRight ? 'translateX(-100%) translateX(-8px)' : 'translateX(8px)'
  }
})

// ── 通用工具函数 ──
const probeStatus = computed(() => props.group.probe_status || 'unknown')

const statusLabel = computed(() => {
  switch (probeStatus.value) {
    case 'online':
      return t('monitoring.online')
    case 'degraded':
      return t('monitoring.degraded')
    case 'offline':
      return t('monitoring.offline')
    default:
      return t('monitoring.unknown')
  }
})

const statusBadgeClass = computed(() => {
  switch (probeStatus.value) {
    case 'online':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'degraded':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'offline':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
  }
})

const statusDotClass = computed(() => {
  switch (probeStatus.value) {
    case 'online':
      return 'bg-green-500'
    case 'degraded':
      return 'bg-amber-500'
    case 'offline':
      return 'bg-red-500'
    default:
      return 'bg-gray-400'
  }
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

const formatLastProbe = (): string => {
  if (!props.group.last_probe_at) return t('monitoring.notCollected')
  const d = new Date(props.group.last_probe_at * 1000)
  const hh = d.getHours().toString().padStart(2, '0')
  const mm = d.getMinutes().toString().padStart(2, '0')
  if (props.group.last_probe_latency_ms > 0) {
    return `${hh}:${mm} · ${props.group.last_probe_latency_ms}ms`
  }
  return `${hh}:${mm}`
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
