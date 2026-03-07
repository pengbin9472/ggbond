<template>
  <AppLayout>
    <div class="card p-4 sm:p-6">
      <div class="mb-6 flex items-center justify-between">
        <h3 class="text-xl font-bold text-gray-900 dark:text-gray-100 sm:text-2xl">
          {{ t('nav.groupMonitoring') }}
        </h3>
        <button
          @click="refresh"
          :disabled="loading"
          class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 disabled:opacity-50"
        >
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading && !groups.length" class="py-12 text-center text-gray-500">
        {{ t('common.loading') }}...
      </div>

      <!-- Empty -->
      <div v-else-if="!groups.length" class="py-12 text-center text-gray-500">
        {{ t('common.noData') }}
      </div>

      <!-- Group Cards -->
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <GroupMonitoringCard
          v-for="group in groups"
          :key="group.group_id"
          :group="group"
          :history="historyMap[group.group_id]"
          @load-history="loadHistory(group.group_id)"
        />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import GroupMonitoringCard from '@/components/monitoring/GroupMonitoringCard.vue'
import { getGroupMonitoring, getGroupMonitoringHistory } from '@/api/monitoring'
import type { GroupMonitoringStat, MonitoringHistoryPoint } from '@/types'

const { t } = useI18n()

const loading = ref(false)
const groups = ref<GroupMonitoringStat[]>([])
const historyMap = reactive<Record<number, MonitoringHistoryPoint[]>>({})

async function refresh() {
  loading.value = true
  try {
    const res = await getGroupMonitoring()
    groups.value = res.groups || []
  } finally {
    loading.value = false
  }
}

async function loadHistory(groupId: number) {
  try {
    const res = await getGroupMonitoringHistory(groupId)
    historyMap[groupId] = res.history || []
  } catch {
    historyMap[groupId] = []
  }
}

onMounted(refresh)
</script>
