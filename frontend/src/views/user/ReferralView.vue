<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Stats Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
            <Icon name="users" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('referral.totalRewards') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">
            ${{ stats?.total_rewards?.toFixed(2) || '0.00' }}
          </p>
          <p class="mt-2 text-sm text-primary-100">
            {{ t('referral.inviteeCount') }}: {{ stats?.invitee_count || 0 }}
          </p>
        </div>
      </div>

      <!-- Invitation Link Card -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('referral.invitationLink') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('referral.invitationLinkHint') }}
          </p>
        </div>
        <div class="p-6">
          <div v-if="loading" class="text-center py-4">
            <div class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary-500 border-r-transparent"></div>
          </div>
          <div v-else class="space-y-4">
            <div>
              <label class="input-label">{{ t('referral.invitationCode') }}</label>
              <div class="mt-1 flex gap-2">
                <input
                  :value="invitationCode"
                  readonly
                  class="input flex-1"
                />
                <button @click="copyCode" class="btn btn-secondary">
                  <Icon name="copy" size="md" />
                </button>
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('referral.invitationUrl') }}</label>
              <div class="mt-1 flex gap-2">
                <input
                  :value="invitationUrl"
                  readonly
                  class="input flex-1"
                />
                <button @click="copyUrl" class="btn btn-primary">
                  <Icon name="copy" size="md" class="mr-2" />
                  {{ t('common.copy') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Reward History -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('referral.rewardHistory') }}
          </h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('referral.inviteeEmail') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('referral.rewardAmount') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('referral.triggerAmount') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('referral.createdAt') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="historyLoading">
                <td colspan="4" class="px-6 py-8 text-center text-gray-500">
                  <div class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary-500 border-r-transparent"></div>
                </td>
              </tr>
              <tr v-else-if="!history.length">
                <td colspan="4" class="px-6 py-8 text-center text-gray-500">
                  {{ t('referral.noHistory') }}
                </td>
              </tr>
              <tr v-else v-for="item in history" :key="item.id" class="hover:bg-gray-50 dark:hover:bg-dark-800">
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-white">
                  {{ item.invitee_email }}
                </td>
                <td class="px-6 py-4 text-sm font-medium text-green-600 dark:text-green-400">
                  +${{ item.reward_amount.toFixed(2) }}
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  ${{ item.trigger_code_value.toFixed(2) }}
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  {{ formatDate(item.created_at) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { referralAPI, type ReferralStats, type ReferralReward } from '@/api/referral'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const historyLoading = ref(true)
const invitationCode = ref('')
const stats = ref<ReferralStats | null>(null)
const history = ref<ReferralReward[]>([])

const invitationUrl = computed(() => {
  if (!invitationCode.value) return ''
  const origin = window.location.origin
  return `${origin}/register?code=${invitationCode.value}`
})

async function loadData() {
  loading.value = true
  try {
    const [codeRes, statsRes] = await Promise.all([
      referralAPI.getInvitationCode(),
      referralAPI.getStats()
    ])
    invitationCode.value = codeRes.code
    stats.value = statsRes
  } catch (error: any) {
    appStore.showError(error.message || t('common.unknownError'))
  } finally {
    loading.value = false
  }
}

async function loadHistory() {
  historyLoading.value = true
  try {
    const res = await referralAPI.getHistory()
    history.value = res.items
  } catch (error: any) {
    appStore.showError(error.message || t('common.unknownError'))
  } finally {
    historyLoading.value = false
  }
}

async function copyCode() {
  try {
    await navigator.clipboard.writeText(invitationCode.value)
    appStore.showSuccess(t('referral.codeCopied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

async function copyUrl() {
  try {
    await navigator.clipboard.writeText(invitationUrl.value)
    appStore.showSuccess(t('referral.urlCopied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  loadData()
  loadHistory()
})
</script>

