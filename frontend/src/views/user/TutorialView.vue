<template>
  <div class="card p-3 sm:p-6">
    <div class="mb-4 sm:mb-8">
      <h3
        class="mb-3 flex items-center text-xl font-bold text-gray-900 dark:text-gray-100 sm:mb-4 sm:text-2xl"
      >
        <svg class="mr-2 h-6 w-6 text-blue-600 sm:mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342M6.75 15a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm0 0v-3.675A55.378 55.378 0 0 1 12 8.443m-7.007 11.55A5.981 5.981 0 0 0 6.75 15.75v-1.5"/></svg>
        {{ currentToolTitle }} 使用教程
      </h3>
      <p class="text-sm text-gray-600 dark:text-gray-400 sm:text-lg">
        跟着这个教程，你可以轻松在自己的电脑上安装并使用 {{ currentToolTitle }}。
      </p>
    </div>

    <!-- 系统选择标签 -->
    <div class="mb-4 sm:mb-6">
      <div class="flex flex-wrap gap-1 rounded-xl bg-gray-100 p-1 dark:bg-gray-800 sm:gap-2 sm:p-2">
        <button
          v-for="system in tutorialSystems"
          :key="system.key"
          :class="[
            'flex flex-1 items-center justify-center gap-1 rounded-lg px-3 py-2 text-xs font-semibold transition-all duration-300 sm:gap-2 sm:px-6 sm:py-3 sm:text-sm',
            activeTutorialSystem === system.key
              ? 'bg-white text-blue-600 shadow-sm dark:bg-blue-600 dark:text-white dark:shadow-blue-500/40'
              : 'text-gray-600 hover:bg-white/50 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-700 dark:hover:text-white'
          ]"
          @click="activeTutorialSystem = system.key"
        >
          <!-- Windows icon -->
          <svg v-if="system.key === 'windows'" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801"/></svg>
          <!-- macOS icon -->
          <svg v-else-if="system.key === 'macos'" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/></svg>
          <!-- Linux icon -->
          <svg v-else-if="system.key === 'linux'" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 0 0-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.368 1.884 1.43.199.008.395-.024.585-.066l.006-.001a.587.587 0 0 0 .196.023c.106 0 .266-.04.396-.136a.585.585 0 0 0 .198-.336.585.585 0 0 0 .023-.2 3.61 3.61 0 0 0-.054-.402c-.063-.32-.158-.667-.164-1.068-.008-.468.066-.9.262-1.333.065-.134.138-.27.205-.406.134-.27.26-.536.33-.803.165-.602.069-1.208-.327-1.681a.585.585 0 0 0-.247-.192c-.097-.135-.2-.27-.298-.404-.297-.4-.586-.8-.859-1.196-.395-.573-.764-1.064-1.08-1.528-.159-.201-.312-.399-.455-.597l-.002-.003c-.292-.399-.584-.798-.822-1.196-.24-.401-.419-.802-.495-1.196-.038-.203-.044-.403-.025-.603.013-.133.039-.268.072-.399a1.543 1.543 0 0 0 .084-.795c-.044-.267-.166-.5-.333-.664a.6.6 0 0 0-.26-.178c.068-.399.159-.798.156-1.203 0-.398-.113-.8-.358-1.133a2.43 2.43 0 0 0-.652-.597c-.338-.2-.702-.266-1.032-.396-.33-.133-.633-.332-.87-.598-.237-.268-.398-.601-.465-.935a3.55 3.55 0 0 1-.04-.797c.008-.399.058-.797.037-1.196-.02-.4-.112-.8-.318-1.133-.205-.332-.538-.598-.93-.665z"/></svg>
          {{ system.name }}
        </button>
      </div>
    </div>

    <!-- CLI 工具选择标签 -->
    <div class="mb-4 sm:mb-8">
      <div class="flex flex-wrap gap-1 rounded-xl bg-gray-100 p-1 dark:bg-gray-800 sm:gap-2 sm:p-2">
        <button
          v-for="tool in cliTools"
          :key="tool.key"
          :class="[
            'flex flex-1 items-center justify-center gap-1 rounded-lg px-3 py-2 text-xs font-semibold transition-all duration-300 sm:gap-2 sm:px-4 sm:py-3 sm:text-sm',
            activeCliTool === tool.key
              ? 'bg-white text-blue-600 shadow-sm dark:bg-blue-600 dark:text-white dark:shadow-blue-500/40'
              : 'text-gray-600 hover:bg-white/50 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-700 dark:hover:text-white'
          ]"
          @click="activeCliTool = tool.key"
        >
          <!-- Claude Code / Robot icon -->
          <svg v-if="tool.key === 'claude-code'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m3.75-1.5V3m0 18v1.5m-9-1.5V21m3.75-1.5V21M9 7.5h6v4.5a3 3 0 0 1-3 3H9V7.5Z"/></svg>
          <!-- OpenClaw / Paw icon -->
          <span v-else-if="tool.key === 'openclaw'" class="text-sm">🐾</span>
          <!-- Codex / Code icon -->
          <svg v-else-if="tool.key === 'codex'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5"/></svg>
          <!-- Gemini CLI / Google icon -->
          <span v-else-if="tool.key === 'gemini-cli'" class="text-sm font-bold">G</span>
          <!-- Droid CLI / Terminal icon -->
          <svg v-else-if="tool.key === 'droid-cli'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="m6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6a2.25 2.25 0 0 0-2.25-2.25H5.25A2.25 2.25 0 0 0 3 6v12a2.25 2.25 0 0 0 2.25 2.25Z"/></svg>
          {{ tool.name }}
        </button>
      </div>
    </div>

    <!-- 动态组件 -->
    <component :is="currentTutorialComponent" :platform="activeTutorialSystem" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Component } from 'vue'
import ClaudeCodeTutorial from '@/components/tutorial/ClaudeCodeTutorial.vue'
import OpenClawTutorial from '@/components/tutorial/OpenClawTutorial.vue'
import GeminiCliTutorial from '@/components/tutorial/GeminiCliTutorial.vue'
import CodexTutorial from '@/components/tutorial/CodexTutorial.vue'
import DroidCliTutorial from '@/components/tutorial/DroidCliTutorial.vue'

interface TutorialSystem {
  key: string
  name: string
}

interface CliTool {
  key: string
  name: string
  component: Component
}

// 当前系统选择
const activeTutorialSystem = ref('windows')

// 当前 CLI 工具选择
const activeCliTool = ref('claude-code')

// 系统列表
const tutorialSystems: TutorialSystem[] = [
  { key: 'windows', name: 'Windows' },
  { key: 'macos', name: 'macOS' },
  { key: 'linux', name: 'Linux / WSL2' }
]

// CLI 工具列表
const cliTools: CliTool[] = [
  { key: 'claude-code', name: 'Claude Code', component: ClaudeCodeTutorial },
  { key: 'openclaw', name: 'OpenClaw', component: OpenClawTutorial },
  { key: 'codex', name: 'Codex', component: CodexTutorial },
  { key: 'gemini-cli', name: 'Gemini CLI', component: GeminiCliTutorial },
  { key: 'droid-cli', name: 'Droid CLI', component: DroidCliTutorial }
]

// 当前工具标题
const currentToolTitle = computed(() => {
  const tool = cliTools.find((t) => t.key === activeCliTool.value)
  return tool ? tool.name : 'CLI 工具'
})

// 当前教程组件
const currentTutorialComponent = computed(() => {
  const tool = cliTools.find((t) => t.key === activeCliTool.value)
  return tool ? tool.component : null
})
</script>

<style scoped>
.tutorial-container {
  min-height: calc(100vh - 300px);
}
</style>
