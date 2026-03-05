<template>
  <div class="tutorial-section">
    <!-- 第一步：安装 Node.js -->
    <NodeInstallTutorial :platform="platform" :step-number="1" tool-name="OpenCode" />

    <!-- 第二步：安装 OpenCode -->
    <div class="mb-4 sm:mb-10 sm:mb-6">
      <h4
        class="mb-3 flex items-center text-lg font-semibold text-gray-800 dark:text-gray-300 sm:mb-4 sm:text-xl"
      >
        <span
          class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-xs font-bold text-white sm:mr-3 sm:h-8 sm:w-8 sm:text-sm"
          >2</span
        >
        安装 OpenCode
      </h4>

      <div
        class="mb-4 rounded-xl border border-green-100 bg-gradient-to-r from-green-50 to-emerald-50 p-4 dark:border-green-500/40 dark:from-green-950/30 dark:to-emerald-950/30 sm:mb-6 sm:p-6"
      >
        <h5
          class="mb-2 flex items-center text-base font-semibold text-gray-800 dark:text-gray-200 sm:mb-3 sm:text-lg"
        >
          <svg
            class="mr-2 h-5 w-5 text-green-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
            />
          </svg>
          全局安装
        </h5>
        <p class="mb-3 text-sm text-gray-700 dark:text-gray-300 sm:mb-4 sm:text-base">
          {{ platform === 'windows' ? '打开 PowerShell 或 Windows Terminal' : '打开终端' }}
          ，运行以下命令：
        </p>
        <div
          class="mb-4 overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 sm:p-4 sm:text-sm"
        >
          <div class="mb-2"># 全局安装 OpenCode</div>
          <div class="whitespace-nowrap text-gray-300">
            {{
              platform === 'windows'
                ? 'npm i -g opencode-ai@latest'
                : 'sudo npm i -g opencode-ai@latest'
            }}
          </div>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400">
          这个命令会从 npm 官方仓库下载并安装最新版本的 OpenCode。
        </p>
      </div>

      <!-- 验证安装 -->
      <div
        class="rounded-lg border border-green-200 bg-green-50 p-3 dark:border-green-500/40 dark:bg-green-950/30 sm:p-4"
      >
        <h6 class="mb-2 font-medium text-green-800 dark:text-green-300">验证 OpenCode 安装</h6>
        <p class="mb-3 text-sm text-green-700 dark:text-green-300">
          安装完成后，输入以下命令检查是否安装成功：
        </p>
        <div
          class="overflow-x-auto rounded bg-gray-900 p-2 font-mono text-xs text-green-400 sm:p-3 sm:text-sm"
        >
          <div class="whitespace-nowrap text-gray-300">opencode --version</div>
        </div>
        <p class="mt-2 text-sm text-green-700 dark:text-green-300">
          如果显示版本号，说明 OpenCode 已经成功安装了。
        </p>
      </div>
    </div>

    <!-- 第三步：安装 SDK 依赖 -->
    <div class="mb-4 sm:mb-10 sm:mb-6">
      <h4
        class="mb-3 flex items-center text-lg font-semibold text-gray-800 dark:text-gray-300 sm:mb-4 sm:text-xl"
      >
        <span
          class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-purple-500 text-xs font-bold text-white sm:mr-3 sm:h-8 sm:w-8 sm:text-sm"
          >3</span
        >
        安装 SDK 依赖
      </h4>

      <div
        class="mb-4 rounded-xl border border-purple-100 bg-gradient-to-r from-purple-50 to-pink-50 p-4 dark:border-purple-500/40 dark:from-purple-950/30 dark:to-pink-950/30 sm:mb-6 sm:p-6"
      >
        <h5
          class="mb-2 flex items-center text-base font-semibold text-gray-800 dark:text-gray-200 sm:mb-3 sm:text-lg"
        >
          <svg
            class="mr-2 h-5 w-5 text-purple-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125"
            />
          </svg>
          在配置目录安装 SDK
        </h5>
        <p class="mb-3 text-sm text-gray-700 dark:text-gray-300 sm:mb-4 sm:text-base">
          OpenCode 需要在配置目录中安装对应的 SDK 包：
        </p>
        <div
          class="overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 sm:p-4 sm:text-sm"
        >
          <template v-if="platform === 'windows'">
            <div class="mb-2"># 进入配置目录</div>
            <div class="whitespace-nowrap text-gray-300">
              cd $env:USERPROFILE\.config\opencode
            </div>
            <div class="mb-2 mt-3"># 如果目录不存在，先创建</div>
            <div class="whitespace-nowrap text-gray-300">
              New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.config\opencode | Out-Null
            </div>
          </template>
          <template v-else>
            <div class="mb-2"># 进入配置目录（如果不存在则创建）</div>
            <div class="whitespace-nowrap text-gray-300">
              mkdir -p ~/.config/opencode && cd ~/.config/opencode
            </div>
          </template>
          <div class="mb-2 mt-3"># 安装 Anthropic SDK</div>
          <div class="whitespace-nowrap text-gray-300">npm install @ai-sdk/anthropic</div>
        </div>
      </div>
    </div>

    <!-- 第四步：配置 opencode.json -->
    <div class="mb-4 sm:mb-10 sm:mb-6">
      <h4
        class="mb-3 flex items-center text-lg font-semibold text-gray-800 dark:text-gray-300 sm:mb-4 sm:text-xl"
      >
        <span
          class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-indigo-500 text-xs font-bold text-white sm:mr-3 sm:h-8 sm:w-8 sm:text-sm"
          >4</span
        >
        配置 opencode.json
      </h4>
      <p class="mb-3 text-sm text-gray-700 dark:text-gray-300 sm:mb-4 sm:text-base">
        OpenCode 支持全局配置和项目配置两种方式：
      </p>

      <div class="mb-4 space-y-3">
        <div
          class="rounded-lg border border-blue-200 bg-blue-50 p-3 dark:border-blue-500/40 dark:bg-blue-950/30 sm:p-4"
        >
          <h6 class="mb-2 font-medium text-blue-800 dark:text-blue-300">配置文件位置</h6>
          <ul class="space-y-2 text-sm text-blue-700 dark:text-blue-300">
            <li>
              <strong>全局配置：</strong>
              <code class="rounded bg-blue-100 px-1 dark:bg-blue-900">{{ globalConfigPath }}</code>
            </li>
            <li>
              <strong>项目配置：</strong>
              <code class="rounded bg-blue-100 px-1 dark:bg-blue-900">./opencode.json</code>
              （项目根目录）
            </li>
          </ul>
          <p class="mt-2 text-xs text-blue-600 dark:text-blue-400">
            项目配置优先级高于全局配置。
          </p>
        </div>
      </div>

      <div class="space-y-4">
        <!-- 配置文件内容 -->
        <div
          class="rounded-lg border border-indigo-200 bg-indigo-50 p-3 dark:border-indigo-500/40 dark:bg-indigo-950/30 sm:p-4"
        >
          <h6 class="mb-2 font-medium text-indigo-800 dark:text-indigo-300">
            配置文件示例
          </h6>
          <p class="mb-3 text-sm text-indigo-700 dark:text-indigo-300">
            在配置文件中添加以下内容，将 <code class="rounded bg-indigo-100 px-1 dark:bg-indigo-900">sk-xxx</code> 替换为你的实际 API 密钥：
          </p>
          <div
            class="overflow-x-auto rounded bg-gray-900 p-2 font-mono text-xs text-green-400 sm:p-3 sm:text-sm"
          >
            <div
              v-for="(line, index) in configJsonLines"
              :key="index"
              class="whitespace-pre text-gray-300"
            >
              {{ line }}
            </div>
          </div>
        </div>

        <!-- 一键写入命令 -->
        <div
          class="rounded-lg border border-orange-200 bg-orange-50 p-3 dark:border-orange-500/40 dark:bg-orange-950/30 sm:p-4"
        >
          <h6 class="mb-2 font-medium text-orange-800 dark:text-orange-300">一键写入命令</h6>
          <p class="mb-3 text-sm text-orange-700 dark:text-orange-300">
            复制以下命令快速创建配置文件：
          </p>
          <div
            class="overflow-x-auto rounded bg-gray-900 p-2 font-mono text-xs text-green-400 sm:p-3 sm:text-sm"
          >
            <div class="whitespace-nowrap text-gray-300">{{ configWriteCmd }}</div>
          </div>
        </div>

        <!-- 提示 -->
        <div
          class="rounded-lg border border-yellow-200 bg-yellow-50 p-3 dark:border-yellow-500/40 dark:bg-yellow-950/30 sm:p-4"
        >
          <p class="text-sm text-yellow-700 dark:text-yellow-300">
            <strong>注意：</strong>OpenCode 目前处于早期阶段，体验可能不够稳定。配置方式可能随版本更新变化，请以
            <a
              class="text-blue-600 underline dark:text-blue-400"
              href="https://opencode.ai"
              target="_blank"
              >官方文档</a
            >
            为准。
          </p>
        </div>
      </div>
    </div>

    <!-- 第五步：启动使用 -->
    <div class="mb-6 sm:mb-8">
      <h4
        class="mb-3 flex items-center text-lg font-semibold text-gray-800 dark:text-gray-300 sm:mb-4 sm:text-xl"
      >
        <span
          class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-orange-500 text-xs font-bold text-white sm:mr-3 sm:h-8 sm:w-8 sm:text-sm"
          >5</span
        >
        启动使用
      </h4>
      <div
        class="rounded-xl border border-orange-100 bg-gradient-to-r from-orange-50 to-yellow-50 p-4 dark:border-orange-500/40 dark:from-orange-950/30 dark:to-yellow-950/30 sm:p-6"
      >
        <p class="mb-3 text-sm text-gray-700 dark:text-gray-300 sm:mb-4 sm:text-base">
          配置完成后，在项目目录中启动 OpenCode：
        </p>

        <div class="space-y-4">
          <div>
            <h6 class="mb-2 text-sm font-medium text-gray-800 dark:text-gray-300 sm:text-base">
              启动 OpenCode
            </h6>
            <div
              class="overflow-x-auto rounded bg-gray-900 p-2 font-mono text-xs text-green-400 sm:p-3 sm:text-sm"
            >
              <div class="mb-2"># 进入你的项目目录</div>
              <div class="whitespace-nowrap text-gray-300">
                cd
                {{
                  platform === 'windows' ? 'C:\\path\\to\\your\\project' : '/path/to/your/project'
                }}
              </div>
              <div class="mb-2 mt-2"># 启动 OpenCode</div>
              <div class="whitespace-nowrap text-gray-300">opencode</div>
            </div>
          </div>
        </div>

        <div
          class="mt-4 rounded-lg border border-blue-200 bg-blue-50 p-3 dark:border-blue-500/40 dark:bg-blue-950/30 sm:p-4"
        >
          <h6 class="mb-2 text-sm font-medium text-blue-800 dark:text-blue-300 sm:text-base">
            更多信息
          </h6>
          <ul class="space-y-1 text-xs text-blue-700 dark:text-blue-300 sm:text-sm">
            <li>
              <a
                class="text-blue-600 underline dark:text-blue-400"
                href="https://opencode.ai"
                target="_blank"
                >OpenCode 官网</a
              >
            </li>
            <li>
              <a
                class="text-blue-600 underline dark:text-blue-400"
                href="https://github.com/nicepkg/opencode"
                target="_blank"
                >GitHub 仓库</a
              >
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTutorialUrls } from '@/composables/useTutorialUrls'
import NodeInstallTutorial from './NodeInstallTutorial.vue'

const props = defineProps<{
  platform: 'windows' | 'macos' | 'linux'
}>()

const { currentBaseUrl } = useTutorialUrls()

const globalConfigPath = computed(() =>
  props.platform === 'windows'
    ? '%USERPROFILE%\\.config\\opencode\\opencode.json'
    : '~/.config/opencode/opencode.json'
)

const configJson = computed(() => {
  const baseUrl = currentBaseUrl.value
  return {
    $schema: 'https://opencode.ai/config.json',
    provider: {
      'crs-claude': {
        npm: '@ai-sdk/anthropic',
        options: {
          apiKey: 'sk-xxx',
          baseURL: baseUrl,
        },
        models: {
          'claude-opus-4-6-20260205': { name: 'Claude Opus 4.6', attachment: true },
        },
      },
    },
    model: 'crs-claude/claude-opus-4-6-20260205',
  }
})

const configJsonLines = computed(() => {
  return JSON.stringify(configJson.value, null, 2).split('\n')
})

const configJsonContent = computed(() => JSON.stringify(configJson.value, null, 2))

const configWriteCmd = computed(() => {
  if (props.platform === 'windows') {
    const escaped = configJsonContent.value.replace(/"/g, '`"').replace(/\n/g, '`n')
    return `New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\\.config\\opencode" | Out-Null; "${escaped}" | Set-Content -Path "$env:USERPROFILE\\.config\\opencode\\opencode.json" -Force`
  }
  const escaped = configJsonContent.value.replace(/\n/g, '\\n')
  return `mkdir -p ~/.config/opencode && printf '${escaped}\\n' > ~/.config/opencode/opencode.json`
})
</script>
