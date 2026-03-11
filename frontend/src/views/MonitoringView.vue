<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- 状态概览头部 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 px-7 py-6"
        style="box-shadow: 0 2px 8px rgba(0,0,0,0.08), 0 1px 3px rgba(0,0,0,0.06)"
      >
        <div class="flex items-center flex-wrap gap-4">
          <!-- 统计数字 -->
          <div class="flex items-center gap-3">
            <span class="flex items-center gap-1.5 text-[13px] font-medium text-gray-700 dark:text-gray-300">
              <span class="w-2 h-2 rounded-full bg-green-500"></span>
              {{ onlineCount }} {{ t('monitoring.online') }}
            </span>
            <span class="flex items-center gap-1.5 text-[13px] font-medium text-gray-700 dark:text-gray-300">
              <span class="w-2 h-2 rounded-full bg-red-500"></span>
              {{ offlineCount }} {{ t('monitoring.offline') }}
            </span>
            <span class="w-px h-4 bg-gray-300 dark:bg-gray-600"></span>
            <span class="text-[13px] text-gray-500 dark:text-gray-400">
              {{ groups.length }} {{ t('monitoring.totalGroupsCount') }}
            </span>
          </div>

          <!-- 右侧：刷新按钮 -->
          <div class="flex items-center gap-3 ml-auto">
            <button
              @click="loadData"
              :disabled="loading"
              class="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
            >
              <svg
                class="w-3.5 h-3.5"
                :class="{ 'animate-spin': loading }"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {{ t('common.refresh') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading && !groups.length" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- 空状态 -->
      <div
        v-else-if="!loading && groups.length === 0"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-12 text-center"
        style="box-shadow: 0 2px 8px rgba(0,0,0,0.08), 0 1px 3px rgba(0,0,0,0.06)"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
          {{ t('monitoring.noGroups') }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('monitoring.noGroupsDescription') }}
        </p>
      </div>

      <!-- 分组监控卡片网格 -->
      <div
        v-else
        class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
      >
        <GroupMonitoringCard
          v-for="group in groups"
          :key="group.group_id"
          :group="group"
          :history="historyMap[group.group_id]"
          @load-history="loadGroupHistory(group.group_id)"
        />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { monitoringAPI } from '@/api/monitoring'
import type { GroupMonitoringStat, MonitoringHistoryPoint } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import GroupMonitoringCard from '@/components/monitoring/GroupMonitoringCard.vue'

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

const loading = ref(false)
const groups = ref<GroupMonitoringStat[]>([])
const historyMap = ref<Record<number, MonitoringHistoryPoint[]>>({})

const onlineCount = computed(() => {
  return groups.value.filter(g => g.probe_status === 'online').length
})

const offlineCount = computed(() => {
  return groups.value.filter(g => g.probe_status !== 'online').length
})

const loadGroupHistory = async (groupId: number) => {
  // 已加载过则跳过
  if (historyMap.value[groupId]) return

  try {
    const res = await monitoringAPI.getGroupMonitoringHistory(groupId, 60)
    historyMap.value = { ...historyMap.value, [groupId]: res.history || [] }
  } catch {
    // 忽略单个分组的历史加载失败
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const response = await monitoringAPI.getGroupMonitoring()
    groups.value = response.groups || []
    historyMap.value = {}

    // 自动加载所有分组的历史数据
    await Promise.allSettled(
      groups.value.map(g => loadGroupHistory(g.group_id))
    )
  } catch (error) {
    appStore.showError(t('monitoring.loadError'))
    console.error('Error loading group monitoring data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (appStore.cachedPublicSettings?.group_monitoring_enabled === false) {
    router.replace('/dashboard')
    return
  }
  loadData()
})
</script>
