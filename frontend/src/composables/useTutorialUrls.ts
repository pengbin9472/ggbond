import { computed, type ComputedRef } from 'vue'

interface TutorialUrls {
  currentBaseUrl: ComputedRef<string>
  geminiBaseUrl: ComputedRef<string>
  openaiBaseUrl: ComputedRef<string>
  droidClaudeBaseUrl: ComputedRef<string>
  droidOpenaiBaseUrl: ComputedRef<string>
}

export function useTutorialUrls(): TutorialUrls {
  const getBaseUrlPrefix = (): string => {
    const customPrefix = import.meta.env.VITE_API_BASE_PREFIX as string | undefined
    if (customPrefix) {
      return customPrefix.replace(/\/$/, '')
    }

    let origin = ''
    if (window.location.origin) {
      origin = window.location.origin
    } else {
      const protocol = window.location.protocol
      const hostname = window.location.hostname
      const port = window.location.port
      origin = protocol + '//' + hostname
      if (
        port &&
        ((protocol === 'http:' && port !== '80') || (protocol === 'https:' && port !== '443'))
      ) {
        origin += ':' + port
      }
    }

    if (!origin) {
      const currentUrl = window.location.href
      const pathStart = currentUrl.indexOf('/', 8)
      if (pathStart !== -1) {
        origin = currentUrl.substring(0, pathStart)
      } else {
        return ''
      }
    }

    return origin
  }

  const currentBaseUrl = computed(() => getBaseUrlPrefix())
  const geminiBaseUrl = computed(() => getBaseUrlPrefix() + '/gemini')
  const openaiBaseUrl = computed(() => getBaseUrlPrefix() + '/openai')
  const droidClaudeBaseUrl = computed(() => getBaseUrlPrefix() + '/droid/claude')
  const droidOpenaiBaseUrl = computed(() => getBaseUrlPrefix() + '/droid/openai')

  return {
    currentBaseUrl,
    geminiBaseUrl,
    openaiBaseUrl,
    droidClaudeBaseUrl,
    droidOpenaiBaseUrl,
  }
}
