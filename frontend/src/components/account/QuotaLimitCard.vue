<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'

const { t } = useI18n()

const props = defineProps<{
  totalLimit: number | null
  dailyLimit: number | null
  weeklyLimit: number | null
  dailyResetMode: 'rolling' | 'fixed' | null
  dailyResetHour: number | null
  weeklyResetMode: 'rolling' | 'fixed' | null
  weeklyResetDay: number | null
  weeklyResetHour: number | null
  resetTimezone: string | null
}>()

const emit = defineEmits<{
  'update:totalLimit': [value: number | null]
  'update:dailyLimit': [value: number | null]
  'update:weeklyLimit': [value: number | null]
  'update:dailyResetMode': [value: 'rolling' | 'fixed' | null]
  'update:dailyResetHour': [value: number | null]
  'update:weeklyResetMode': [value: 'rolling' | 'fixed' | null]
  'update:weeklyResetDay': [value: number | null]
  'update:weeklyResetHour': [value: number | null]
  'update:resetTimezone': [value: string | null]
}>()

const enabled = computed(() =>
  (props.totalLimit != null && props.totalLimit > 0) ||
  (props.dailyLimit != null && props.dailyLimit > 0) ||
  (props.weeklyLimit != null && props.weeklyLimit > 0)
)

const localEnabled = ref(enabled.value)

// Sync when props change externally
watch(enabled, (val) => {
  localEnabled.value = val
})

// When toggle is turned off, clear all values
watch(localEnabled, (val) => {
  if (!val) {
    emit('update:totalLimit', null)
    emit('update:dailyLimit', null)
    emit('update:weeklyLimit', null)
    emit('update:dailyResetMode', null)
    emit('update:dailyResetHour', null)
    emit('update:weeklyResetMode', null)
    emit('update:weeklyResetDay', null)
    emit('update:weeklyResetHour', null)
    emit('update:resetTimezone', null)
  }
})

// Whether any fixed mode is active (to show timezone selector)
const hasFixedMode = computed(() =>
  props.dailyResetMode === 'fixed' || props.weeklyResetMode === 'fixed'
)

// Common timezone options
const timezoneOptions = [
  'UTC',
  'Asia/Shanghai',
  'Asia/Tokyo',
  'Asia/Seoul',
  'Asia/Singapore',
  'Asia/Kolkata',
  'Asia/Dubai',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Moscow',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Sao_Paulo',
  'Australia/Sydney',
  'Pacific/Auckland',
]

// Hours for dropdown (0-23)
const hourOptions = Array.from({ length: 24 }, (_, i) => i)

// Day of week options
const dayOptions = [
  { value: 1, key: 'monday' },
  { value: 2, key: 'tuesday' },
  { value: 3, key: 'wednesday' },
  { value: 4, key: 'thursday' },
  { value: 5, key: 'friday' },
  { value: 6, key: 'saturday' },
  { value: 0, key: 'sunday' },
]

const resetModeOptions = computed<SelectOption[]>(() => [
  { value: 'rolling', label: t('admin.accounts.quotaResetModeRolling') },
  { value: 'fixed', label: t('admin.accounts.quotaResetModeFixed') },
])

const hourSelectOptions = computed<SelectOption[]>(() =>
  hourOptions.map((h) => ({ value: h, label: `${String(h).padStart(2, '0')}:00` }))
)

const daySelectOptions = computed<SelectOption[]>(() =>
  dayOptions.map((d) => ({ value: d.value, label: t('admin.accounts.dayOfWeek.' + d.key) }))
)

const timezoneSelectOptions = computed<SelectOption[]>(() =>
  timezoneOptions.map((tz) => ({ value: tz, label: tz }))
)

const onTotalInput = (e: Event) => {
  const raw = (e.target as HTMLInputElement).valueAsNumber
  emit('update:totalLimit', Number.isNaN(raw) ? null : raw)
}
const onDailyInput = (e: Event) => {
  const raw = (e.target as HTMLInputElement).valueAsNumber
  emit('update:dailyLimit', Number.isNaN(raw) ? null : raw)
}
const onWeeklyInput = (e: Event) => {
  const raw = (e.target as HTMLInputElement).valueAsNumber
  emit('update:weeklyLimit', Number.isNaN(raw) ? null : raw)
}

const onDailyModeChange = (value: string | number | boolean | null) => {
  const val = value === 'fixed' ? 'fixed' : 'rolling'
  emit('update:dailyResetMode', val)
  if (val === 'fixed') {
    if (props.dailyResetHour == null) emit('update:dailyResetHour', 0)
    if (!props.resetTimezone) emit('update:resetTimezone', 'UTC')
  }
}

const onWeeklyModeChange = (value: string | number | boolean | null) => {
  const val = value === 'fixed' ? 'fixed' : 'rolling'
  emit('update:weeklyResetMode', val)
  if (val === 'fixed') {
    if (props.weeklyResetDay == null) emit('update:weeklyResetDay', 1)
    if (props.weeklyResetHour == null) emit('update:weeklyResetHour', 0)
    if (!props.resetTimezone) emit('update:resetTimezone', 'UTC')
  }
}

const onDailyResetHourChange = (value: string | number | boolean | null) => {
  const hour = Number(value)
  emit('update:dailyResetHour', Number.isFinite(hour) ? hour : 0)
}

const onWeeklyResetDayChange = (value: string | number | boolean | null) => {
  const day = Number(value)
  emit('update:weeklyResetDay', Number.isFinite(day) ? day : 1)
}

const onWeeklyResetHourChange = (value: string | number | boolean | null) => {
  const hour = Number(value)
  emit('update:weeklyResetHour', Number.isFinite(hour) ? hour : 0)
}

const onResetTimezoneChange = (value: string | number | boolean | null) => {
  emit('update:resetTimezone', typeof value === 'string' ? value : 'UTC')
}
</script>

<template>
  <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
      <div class="mb-3 flex items-center justify-between">
        <div>
          <label class="input-label mb-0">{{ t('admin.accounts.quotaLimitToggle') }}</label>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.accounts.quotaLimitToggleHint') }}
          </p>
        </div>
        <button
          type="button"
          @click="localEnabled = !localEnabled"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
            localEnabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
          ]"
        >
          <span
            :class="[
              'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
              localEnabled ? 'translate-x-5' : 'translate-x-0'
            ]"
          />
        </button>
      </div>

      <div v-if="localEnabled" class="space-y-3">
        <!-- 日配额 -->
        <div>
          <label class="input-label">{{ t('admin.accounts.quotaDailyLimit') }}</label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 dark:text-gray-400">$</span>
            <input
              :value="dailyLimit"
              @input="onDailyInput"
              type="number"
              min="0"
              step="0.01"
              class="input pl-7"
              :placeholder="t('admin.accounts.quotaLimitPlaceholder')"
            />
          </div>
          <!-- 日配额重置模式 -->
          <div class="mt-2 flex items-center gap-2">
            <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetMode') }}</label>
            <Select
              :model-value="dailyResetMode || 'rolling'"
              :options="resetModeOptions"
              class="w-36"
              @update:model-value="onDailyModeChange"
            />
          </div>
          <!-- 固定模式：小时选择 -->
          <div v-if="dailyResetMode === 'fixed'" class="mt-2 flex items-center gap-2">
            <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetHour') }}</label>
            <Select
              :model-value="dailyResetHour ?? 0"
              :options="hourSelectOptions"
              class="w-24"
              @update:model-value="onDailyResetHourChange"
            />
          </div>
          <p class="input-hint">
            <template v-if="dailyResetMode === 'fixed'">
              {{ t('admin.accounts.quotaDailyLimitHintFixed', { hour: String(dailyResetHour ?? 0).padStart(2, '0'), timezone: resetTimezone || 'UTC' }) }}
            </template>
            <template v-else>
              {{ t('admin.accounts.quotaDailyLimitHint') }}
            </template>
          </p>
        </div>

        <!-- 周配额 -->
        <div>
          <label class="input-label">{{ t('admin.accounts.quotaWeeklyLimit') }}</label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 dark:text-gray-400">$</span>
            <input
              :value="weeklyLimit"
              @input="onWeeklyInput"
              type="number"
              min="0"
              step="0.01"
              class="input pl-7"
              :placeholder="t('admin.accounts.quotaLimitPlaceholder')"
            />
          </div>
          <!-- 周配额重置模式 -->
          <div class="mt-2 flex items-center gap-2">
            <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetMode') }}</label>
            <Select
              :model-value="weeklyResetMode || 'rolling'"
              :options="resetModeOptions"
              class="w-36"
              @update:model-value="onWeeklyModeChange"
            />
          </div>
          <!-- 固定模式：星期几 + 小时 -->
          <div v-if="weeklyResetMode === 'fixed'" class="mt-2 flex items-center gap-2 flex-wrap">
            <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaWeeklyResetDay') }}</label>
            <Select
              :model-value="weeklyResetDay ?? 1"
              :options="daySelectOptions"
              class="w-32"
              @update:model-value="onWeeklyResetDayChange"
            />
            <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetHour') }}</label>
            <Select
              :model-value="weeklyResetHour ?? 0"
              :options="hourSelectOptions"
              class="w-24"
              @update:model-value="onWeeklyResetHourChange"
            />
          </div>
          <p class="input-hint">
            <template v-if="weeklyResetMode === 'fixed'">
              {{ t('admin.accounts.quotaWeeklyLimitHintFixed', { day: t('admin.accounts.dayOfWeek.' + (dayOptions.find(d => d.value === (weeklyResetDay ?? 1))?.key || 'monday')), hour: String(weeklyResetHour ?? 0).padStart(2, '0'), timezone: resetTimezone || 'UTC' }) }}
            </template>
            <template v-else>
              {{ t('admin.accounts.quotaWeeklyLimitHint') }}
            </template>
          </p>
        </div>

        <!-- 时区选择（当任一维度使用固定模式时显示） -->
        <div v-if="hasFixedMode">
          <label class="input-label">{{ t('admin.accounts.quotaResetTimezone') }}</label>
          <Select
            :model-value="resetTimezone || 'UTC'"
            :options="timezoneSelectOptions"
            @update:model-value="onResetTimezoneChange"
          />
        </div>

        <!-- 总配额 -->
        <div>
          <label class="input-label">{{ t('admin.accounts.quotaTotalLimit') }}</label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 dark:text-gray-400">$</span>
            <input
              :value="totalLimit"
              @input="onTotalInput"
              type="number"
              min="0"
              step="0.01"
              class="input pl-7"
              :placeholder="t('admin.accounts.quotaLimitPlaceholder')"
            />
          </div>
          <p class="input-hint">{{ t('admin.accounts.quotaTotalLimitHint') }}</p>
        </div>
      </div>
  </div>
</template>
