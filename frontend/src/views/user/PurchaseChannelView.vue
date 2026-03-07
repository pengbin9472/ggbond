<template>
  <AppLayout>
    <div class="mx-auto max-w-lg py-6 px-4">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="!channelEnabled" class="card flex items-center justify-center p-10 text-center">
        <div class="max-w-md">
          <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
            <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 01.75-.75h3a.75.75 0 01.75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349m-16.5 11.65V9.35m0 0a3.001 3.001 0 003.75-.615A2.993 2.993 0 009.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 002.25 1.016c.896 0 1.7-.393 2.25-1.016a3.001 3.001 0 003.75.614m-16.5 0a3.004 3.004 0 01-.621-4.72L4.318 3.44A1.5 1.5 0 015.378 3h13.243a1.5 1.5 0 011.06.44l1.19 1.189a3 3 0 01-.621 4.72m-13.5 8.65h3.75a.75.75 0 00.75-.75v-2.25a.75.75 0 00-.75-.75h-3.75a.75.75 0 00-.75.75v2.25c0 .414.336.75.75.75z" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('purchaseChannel.notEnabledTitle') }}
          </h3>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchaseChannel.notEnabledDesc') }}
          </p>
        </div>
      </div>

      <div v-else-if="!isValidUrl" class="card flex items-center justify-center p-10 text-center">
        <div class="max-w-md">
          <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
            <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m9.868-3.293a4.5 4.5 0 00-6.364-6.364L4.318 7.318a4.5 4.5 0 006.364 6.364l2.462-2.462" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('purchaseChannel.notConfiguredTitle') }}
          </h3>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchaseChannel.notConfiguredDesc') }}
          </p>
        </div>
      </div>

      <!-- Main content card -->
      <div v-else class="card overflow-hidden">
        <!-- Shop poster image -->
        <div v-if="channelImage" class="w-full">
          <img
            :src="channelImage"
            alt=""
            class="w-full object-contain"
          />
        </div>

        <!-- Link & actions -->
        <div class="space-y-4 p-5">
          <!-- URL row -->
          <div class="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-800">
            <svg class="h-4 w-4 flex-shrink-0 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m9.868-3.293a4.5 4.5 0 00-6.364-6.364L4.318 7.318a4.5 4.5 0 006.364 6.364l2.462-2.462" />
            </svg>
            <span class="flex-1 truncate text-sm text-gray-600 dark:text-gray-300">
              {{ channelUrl }}
            </span>
            <button
              type="button"
              class="flex-shrink-0 rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
              :class="copied
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
              @click="copyLink"
            >
              {{ copied ? t('purchaseChannel.copied') : t('purchaseChannel.copyLink') }}
            </button>
          </div>

          <!-- Open shop button -->
          <a
            :href="channelUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex w-full items-center justify-center gap-2 rounded-lg bg-red-500 px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-red-600 active:bg-red-700"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 01.75-.75h3a.75.75 0 01.75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349m-16.5 11.65V9.35m0 0a3.001 3.001 0 003.75-.615A2.993 2.993 0 009.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 002.25 1.016c.896 0 1.7-.393 2.25-1.016a3.001 3.001 0 003.75.614m-16.5 0a3.004 3.004 0 01-.621-4.72L4.318 3.44A1.5 1.5 0 015.378 3h13.243a1.5 1.5 0 011.06.44l1.19 1.189a3 3 0 01-.621 4.72m-13.5 8.65h3.75a.75.75 0 00.75-.75v-2.25a.75.75 0 00-.75-.75h-3.75a.75.75 0 00-.75.75v2.25c0 .414.336.75.75.75z" />
            </svg>
            {{ t('purchaseChannel.openShop') }}
          </a>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const copied = ref(false)

const channelEnabled = computed(() => {
  return appStore.cachedPublicSettings?.purchase_channel_enabled ?? false
})

const channelUrl = computed(() => {
  return (appStore.cachedPublicSettings?.purchase_channel_url || '').trim()
})

const channelImage = computed(() => {
  return (appStore.cachedPublicSettings?.purchase_channel_image || '').trim()
})

const isValidUrl = computed(() => {
  const url = channelUrl.value
  return url.startsWith('http://') || url.startsWith('https://')
})

async function copyLink() {
  try {
    await navigator.clipboard.writeText(channelUrl.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback
    const textarea = document.createElement('textarea')
    textarea.value = channelUrl.value
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}

onMounted(async () => {
  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    loading.value = false
  }
})
</script>
