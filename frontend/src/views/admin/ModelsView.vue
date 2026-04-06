<template>
  <AppLayout>
    <div class="min-h-full bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.10),_transparent_32%),radial-gradient(circle_at_top_right,_rgba(249,115,22,0.10),_transparent_28%),linear-gradient(180deg,_#f8fafc_0%,_#eef2f7_100%)] px-4 py-6 dark:bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.12),_transparent_30%),radial-gradient(circle_at_top_right,_rgba(249,115,22,0.10),_transparent_26%),linear-gradient(180deg,_#0b1220_0%,_#111827_100%)] sm:px-6 lg:px-8">
      <div class="mx-auto max-w-7xl space-y-6">
        <section class="overflow-hidden rounded-[28px] border border-white/70 bg-white/80 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur dark:border-white/10 dark:bg-slate-900/70">
          <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
            <div class="space-y-3">
              <span class="inline-flex w-fit items-center rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-sky-700 dark:border-sky-500/30 dark:bg-sky-500/10 dark:text-sky-200">
                {{ t('admin.models.badge') }}
              </span>
              <div class="space-y-2">
                <h1 class="text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">
                  {{ t('admin.models.title') }}
                </h1>
                <p class="max-w-3xl text-sm leading-6 text-slate-600 dark:text-slate-300">
                  {{ t('admin.models.description') }}
                </p>
              </div>
            </div>
            <div class="grid gap-3 sm:grid-cols-3">
              <div class="rounded-2xl border border-slate-200/80 bg-slate-50/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-800/70">
                <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">{{ t('admin.models.stats.totalModels') }}</div>
                <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">{{ filteredModels.length }}</div>
              </div>
              <div class="rounded-2xl border border-slate-200/80 bg-slate-50/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-800/70">
                <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">{{ t('admin.models.stats.platforms') }}</div>
                <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">{{ platformOptions.length }}</div>
              </div>
              <div class="rounded-2xl border border-slate-200/80 bg-slate-50/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-800/70">
                <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">{{ t('admin.models.stats.pricedModels') }}</div>
                <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">{{ pricedModelsCount }}</div>
              </div>
            </div>
          </div>

          <div class="mt-6 flex flex-col gap-3 lg:flex-row lg:items-center">
            <div class="relative flex-1">
              <input
                v-model="searchQuery"
                type="text"
                class="w-full rounded-2xl border border-slate-200 bg-white/90 px-4 py-3 pl-11 text-sm text-slate-900 outline-none ring-0 transition placeholder:text-slate-400 focus:border-sky-400 dark:border-slate-700 dark:bg-slate-950/60 dark:text-white"
                :placeholder="t('admin.models.searchPlaceholder')"
              />
              <Icon name="search" size="sm" class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" />
            </div>
            <Select
              v-model="selectedPlatform"
              class="w-full lg:w-56"
              :options="platformSelectOptions"
            />
          </div>
        </section>

        <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="idx in 6"
            :key="idx"
            class="h-56 animate-pulse rounded-[28px] border border-white/70 bg-white/70 dark:border-white/10 dark:bg-slate-900/50"
          />
        </div>

        <section v-else-if="filteredModels.length > 0" class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="model in filteredModels"
            :key="model.id"
            class="group relative overflow-hidden rounded-[30px] border border-slate-200/80 bg-white/92 p-5 shadow-[0_18px_65px_rgba(15,23,42,0.08)] transition duration-200 hover:-translate-y-1 hover:shadow-[0_24px_80px_rgba(15,23,42,0.14)] dark:border-slate-700 dark:bg-slate-900/84"
          >
            <div class="absolute inset-x-5 top-0 h-px bg-gradient-to-r from-transparent via-sky-300/80 to-transparent dark:via-sky-500/60" />
            <div class="flex items-start gap-4">
              <div class="flex h-20 w-20 shrink-0 items-center justify-center rounded-[24px] border border-slate-200 bg-white shadow-[0_14px_30px_rgba(15,23,42,0.10)] dark:border-slate-700 dark:bg-slate-950">
                <ModelIcon :model="model.id" size="42px" />
              </div>
              <div class="min-w-0 flex-1 space-y-3">
                <div class="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3">
                  <div class="min-w-0">
                    <h2 class="text-[1.45rem] font-semibold leading-tight tracking-tight text-slate-800 [overflow-wrap:anywhere] dark:text-white">
                      {{ model.display_name }}
                    </h2>
                    <p
                      v-if="model.display_name !== model.id"
                      class="mt-1 text-xs font-mono text-slate-500 [overflow-wrap:anywhere] dark:text-slate-400"
                    >
                      {{ model.id }}
                    </p>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <span class="inline-flex items-center rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                        {{ platformLabel(model.platform) }}
                      </span>
                      <span class="inline-flex items-center rounded-full bg-violet-100 px-2.5 py-1 text-xs font-medium text-violet-700 dark:bg-violet-500/15 dark:text-violet-200">
                        {{ t('admin.models.usageBadge') }}
                      </span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-slate-200 bg-white text-slate-500 transition hover:border-sky-300 hover:text-sky-600 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300 dark:hover:border-sky-500 dark:hover:text-sky-300"
                    :title="t('admin.models.copyModelId')"
                    @click="copyModelId(model.id)"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                </div>

                <div class="grid gap-2 text-[1.05rem] text-slate-600 dark:text-slate-300 sm:grid-cols-2">
                  <span>{{ t('admin.models.inputPrice') }} {{ formatPrice(model.input_price) }}</span>
                  <span>{{ t('admin.models.outputPrice') }} {{ formatPrice(model.output_price) }}</span>
                  <span v-if="model.cache_write_price != null">{{ t('admin.models.cacheWritePrice') }} {{ formatPrice(model.cache_write_price) }}</span>
                  <span v-if="model.cache_read_price != null">{{ t('admin.models.cacheReadPrice') }} {{ formatPrice(model.cache_read_price) }}</span>
                  <span v-if="model.image_output_price != null">{{ t('admin.models.imageOutputPrice') }} {{ formatPrice(model.image_output_price) }}</span>
                </div>
              </div>
            </div>

            <div class="mt-8 flex flex-wrap items-center gap-3">
              <span class="inline-flex items-center rounded-full bg-emerald-100 px-3 py-1.5 text-sm font-medium text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200">
                {{ t('admin.models.metered') }}
              </span>
              <span
                v-if="isPriceMissing(model)"
                class="inline-flex items-center rounded-full bg-amber-100 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/15 dark:text-amber-200"
              >
                {{ t('admin.models.priceMissing') }}
              </span>
            </div>
          </article>
        </section>

        <section v-else class="rounded-[28px] border border-dashed border-slate-300 bg-white/70 px-6 py-16 text-center dark:border-slate-700 dark:bg-slate-900/60">
          <div class="mx-auto max-w-md space-y-3">
            <h2 class="text-xl font-semibold text-slate-900 dark:text-white">{{ t('admin.models.emptyTitle') }}</h2>
            <p class="text-sm leading-6 text-slate-600 dark:text-slate-300">{{ t('admin.models.emptyDescription') }}</p>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { getModelCatalog } from '@/api/models'
import type { ModelCatalogEntry } from '@/types'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const searchQuery = ref('')
const selectedPlatform = ref('')
const models = ref<ModelCatalogEntry[]>([])

const platformOptions = computed(() =>
  Array.from(new Set(models.value.map(model => model.platform).filter(Boolean))).sort()
)

const platformSelectOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.models.allPlatforms') },
  ...platformOptions.value.map(platform => ({
    value: platform,
    label: platformLabel(platform),
  })),
])

const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return models.value.filter(model => {
    if (selectedPlatform.value && model.platform !== selectedPlatform.value) {
      return false
    }
    if (!query) {
      return true
    }
    return (
      model.id.toLowerCase().includes(query) ||
      model.display_name.toLowerCase().includes(query) ||
      platformLabel(model.platform).toLowerCase().includes(query)
    )
  })
})

const pricedModelsCount = computed(() => models.value.filter(hasCompletePricing).length)

function hasAnyOutputPricing(model: ModelCatalogEntry): boolean {
  return model.output_price != null || model.image_output_price != null
}

function hasCompletePricing(model: ModelCatalogEntry): boolean {
  return model.input_price != null && hasAnyOutputPricing(model)
}

function isPriceMissing(model: ModelCatalogEntry): boolean {
  return model.pricing_fallback || !hasCompletePricing(model)
}

function platformLabel(platform: string): string {
  if (!platform) {
    return t('admin.models.unknownPlatform')
  }
  const key = `admin.groups.platforms.${platform}`
  const translated = t(key)
  return translated === key ? platform : translated
}

function formatPrice(value?: number | null): string {
  if (value == null) {
    return '--'
  }
  return `$${value.toFixed(4)}/M`
}

async function loadCatalog() {
  loading.value = true
  try {
    const data = await getModelCatalog()
    models.value = data.models
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.models.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function copyModelId(modelID: string) {
  try {
    await navigator.clipboard.writeText(modelID)
    appStore.showSuccess(t('admin.models.copySuccess'))
  } catch {
    appStore.showError(t('admin.models.copyFailed'))
  }
}

onMounted(() => {
  void loadCatalog()
})
</script>
