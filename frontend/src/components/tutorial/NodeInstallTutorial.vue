<template>
  <div class="mb-4 sm:mb-10 sm:mb-6">
    <h4
      class="mb-3 flex items-center text-lg font-semibold text-gray-800 dark:text-gray-300 sm:mb-4 sm:text-xl"
    >
      <span
        class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-blue-500 text-xs font-bold text-white sm:mr-3 sm:h-8 sm:w-8 sm:text-sm"
        >{{ stepNumber }}</span
      >
      安装 Node.js 环境
    </h4>
    <p class="mb-4 text-sm text-gray-600 dark:text-gray-400 sm:mb-4 sm:mb-6 sm:text-base">
      {{ toolName }} 需要 Node.js 环境才能运行。
    </p>

    <!-- Windows -->
    <div v-if="platform === 'windows'" class="node-install-section">
      <div
        class="mb-4 rounded-xl border border-blue-100 bg-gradient-to-r from-blue-50 to-indigo-50 p-4 dark:border-blue-500/40 dark:from-blue-950/30 dark:to-indigo-950/30 sm:mb-6 sm:p-6"
      >
        <h5
          class="mb-2 flex items-center text-base font-semibold text-gray-800 dark:text-gray-200 sm:mb-3 sm:text-lg"
        >
          <svg class="mr-2 h-5 w-5 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
            <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
          </svg>
          Windows 安装方法
        </h5>
        <div class="mb-3 sm:mb-4">
          <p class="mb-2 text-sm text-gray-700 dark:text-gray-300 sm:mb-3 sm:text-base">
            方法一：官网下载（推荐）
          </p>
          <ol
            class="ml-2 list-inside list-decimal space-y-1 text-xs text-gray-600 dark:text-gray-400 sm:ml-4 sm:space-y-2 sm:text-sm"
          >
            <li>
              打开浏览器访问
              <code
                class="rounded bg-gray-100 px-1 py-1 text-xs dark:bg-gray-800 dark:text-yellow-400 sm:px-2 sm:text-sm"
                >https://nodejs.org/</code
              >
            </li>
            <li>点击 "LTS" 版本进行下载（推荐长期支持版本）</li>
            <li>
              下载完成后双击
              <code
                class="rounded bg-gray-100 px-1 py-1 text-xs dark:bg-gray-800 dark:text-yellow-400 sm:px-2 sm:text-sm"
                >.msi</code
              >
              文件
            </li>
            <li>按照安装向导完成安装，保持默认设置即可</li>
          </ol>
        </div>
        <div class="mb-3 sm:mb-4">
          <p class="mb-2 text-sm text-gray-700 dark:text-gray-300 sm:mb-3 sm:text-base">
            方法二：使用包管理器
          </p>
          <p class="mb-2 text-xs text-gray-600 dark:text-gray-400 sm:text-sm">
            如果你安装了 Chocolatey 或 Scoop，可以使用命令行安装：
          </p>
          <div
            class="overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 dark:border dark:border-gray-700 dark:bg-gray-900 sm:p-4 sm:text-sm"
          >
            <div class="mb-2"># 使用 Chocolatey</div>
            <div class="whitespace-nowrap text-gray-300">choco install nodejs</div>
            <div class="mb-2 mt-3"># 或使用 Scoop</div>
            <div class="whitespace-nowrap text-gray-300">scoop install nodejs</div>
          </div>
        </div>
        <div
          class="rounded-lg border border-blue-200 bg-blue-50 p-3 dark:border-blue-500/40 dark:bg-blue-950/30 sm:p-4"
        >
          <h6 class="mb-2 text-sm font-medium text-blue-800 dark:text-blue-300 sm:text-base">
            Windows 注意事项
          </h6>
          <ul class="space-y-1 text-xs text-blue-700 dark:text-blue-300 sm:text-sm">
            <li>• 建议使用 PowerShell 而不是 CMD</li>
            <li>• 如果遇到权限问题，尝试以管理员身份运行</li>
            <li>• 某些杀毒软件可能会误报，需要添加白名单</li>
          </ul>
        </div>
      </div>
      <VerifyInstall terminal="PowerShell 或 CMD" />
    </div>

    <!-- macOS -->
    <div v-else-if="platform === 'macos'" class="node-install-section">
      <div
        class="mb-4 rounded-xl border border-gray-200 bg-gradient-to-r from-gray-50 to-gray-50 p-4 dark:border-gray-700 dark:from-gray-800 dark:to-gray-800 sm:mb-6 sm:p-6"
      >
        <h5
          class="mb-2 flex items-center text-base font-semibold text-gray-800 dark:text-gray-200 sm:mb-3 sm:text-lg"
        >
          <svg class="mr-2 h-5 w-5 text-gray-700 dark:text-gray-400" fill="currentColor" viewBox="0 0 24 24">
            <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11" />
          </svg>
          macOS 安装方法
        </h5>
        <div class="mb-4">
          <p class="mb-3 text-gray-700 dark:text-gray-300">方法一：使用 Homebrew（推荐）</p>
          <p class="mb-2 text-xs text-gray-600 dark:text-gray-400 sm:text-sm">
            如果你已经安装了 Homebrew，使用它安装 Node.js 会更方便：
          </p>
          <div
            class="overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 dark:border dark:border-gray-700 dark:bg-gray-900 sm:p-4 sm:text-sm"
          >
            <div class="mb-2"># 更新 Homebrew</div>
            <div class="whitespace-nowrap text-gray-300">brew update</div>
            <div class="mb-2 mt-3"># 安装 Node.js</div>
            <div class="whitespace-nowrap text-gray-300">brew install node</div>
          </div>
        </div>
        <div class="mb-4">
          <p class="mb-3 text-gray-700 dark:text-gray-300">方法二：官网下载</p>
          <ol
            class="ml-2 list-inside list-decimal space-y-1 text-xs text-gray-600 dark:text-gray-400 sm:ml-4 sm:space-y-2 sm:text-sm"
          >
            <li>
              访问
              <code
                class="rounded bg-gray-100 px-1 py-1 text-xs dark:bg-gray-700 sm:px-2 sm:text-sm"
                >https://nodejs.org/</code
              >
            </li>
            <li>下载适合 macOS 的 LTS 版本</li>
            <li>
              打开下载的
              <code
                class="rounded bg-gray-100 px-1 py-1 text-xs dark:bg-gray-700 sm:px-2 sm:text-sm"
                >.pkg</code
              >
              文件
            </li>
            <li>按照安装程序指引完成安装</li>
          </ol>
        </div>
        <div
          class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-600 dark:bg-gray-800 sm:p-4"
        >
          <h6 class="mb-2 text-sm font-medium text-gray-800 dark:text-gray-300 sm:text-base">
            macOS 注意事项
          </h6>
          <ul class="space-y-1 text-xs text-gray-700 dark:text-gray-300 sm:text-sm">
            <li>
              • 如果遇到权限问题，可能需要使用
              <code class="rounded bg-gray-200 px-1 text-xs dark:bg-gray-700 sm:text-sm">sudo</code>
            </li>
            <li>• 首次运行可能需要在系统偏好设置中允许</li>
            <li>• 建议使用 Terminal 或 iTerm2</li>
          </ul>
        </div>
      </div>
      <VerifyInstall terminal="Terminal" />
    </div>

    <!-- Linux -->
    <div v-else class="node-install-section">
      <div
        class="mb-4 rounded-xl border border-orange-100 bg-gradient-to-r from-orange-50 to-amber-50 p-4 dark:border-orange-500/40 dark:from-orange-950/30 dark:to-amber-950/30 sm:mb-6 sm:p-6"
      >
        <h5
          class="mb-2 flex items-center text-base font-semibold text-gray-800 dark:text-gray-200 sm:mb-3 sm:text-lg"
        >
          <svg class="mr-2 h-5 w-5 text-orange-600" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.368 1.884 1.43.868.07 1.724-.334 2.073-.87.254-.4.38-.865.544-1.332.108-.267.217-.602.347-.7.227-.2.635-.2 1.035-.267.39-.07.8-.132 1.08-.467.14-.2.22-.467.22-.668 0-.466-.22-.865-.397-1.196-.22-.4-.22-.735-.22-1.067 0-.466-.089-.867-.31-1.267-.11-.2-.24-.4-.37-.6-.13-.2-.267-.467-.267-.733 0-.6.338-.933.588-1.267.075-.1.272-.4.296-.467.24-.68-.07-1.467-.673-2.134-.278-.334-.694-.535-1.017-.802-.39-.267-.746-.467-1.127-.733-.38-.267-.645-.467-.923-.7a6.058 6.058 0 01-.832-.934c-.56-.8-.777-1.734-.777-2.667 0-.399.068-.8.186-1.133.12-.4.28-.733.382-1.067.1-.333.182-.667.165-1.066-.03-.466-.178-.867-.39-1.267-.21-.467-.476-.867-.76-1.334-.28-.466-.525-.866-.716-1.333-.19-.467-.332-.933-.332-1.533 0-.2.015-.468.073-.668.057-.2.136-.398.27-.533.09-.134.18-.267.27-.4.067-.134.1-.267.13-.4.03-.134.03-.2-.07-.267a.835.835 0 00-.47-.134z" />
          </svg>
          Linux / WSL2 安装方法
        </h5>
        <div class="mb-4">
          <p class="mb-3 text-gray-700 dark:text-gray-300">方法一：使用 nvm（推荐）</p>
          <p class="mb-2 text-xs text-gray-600 dark:text-gray-400 sm:text-sm">
            nvm 可以方便地管理多个 Node.js 版本：
          </p>
          <div
            class="overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 dark:border dark:border-gray-700 dark:bg-gray-900 sm:p-4 sm:text-sm"
          >
            <div class="mb-2"># 安装 nvm</div>
            <div class="whitespace-nowrap text-gray-300">
              curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
            </div>
            <div class="mb-2 mt-3"># 重新加载 shell 配置</div>
            <div class="whitespace-nowrap text-gray-300">source ~/.bashrc</div>
            <div class="mb-2 mt-3"># 安装最新 LTS 版本</div>
            <div class="whitespace-nowrap text-gray-300">nvm install --lts</div>
          </div>
        </div>
        <div class="mb-4">
          <p class="mb-3 text-gray-700 dark:text-gray-300">方法二：使用包管理器</p>
          <div
            class="overflow-x-auto rounded-lg bg-gray-900 p-3 font-mono text-xs text-green-400 dark:border dark:border-gray-700 dark:bg-gray-900 sm:p-4 sm:text-sm"
          >
            <div class="mb-2"># Ubuntu/Debian</div>
            <div class="whitespace-nowrap text-gray-300">
              curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            </div>
            <div class="whitespace-nowrap text-gray-300">sudo apt-get install -y nodejs</div>
            <div class="mb-2 mt-3"># Fedora</div>
            <div class="whitespace-nowrap text-gray-300">sudo dnf install nodejs</div>
            <div class="mb-2 mt-3"># Arch Linux</div>
            <div class="whitespace-nowrap text-gray-300">sudo pacman -S nodejs npm</div>
          </div>
        </div>
        <div
          class="rounded-lg border border-orange-200 bg-orange-50 p-3 dark:border-orange-500/40 dark:bg-orange-950/30 sm:p-4"
        >
          <h6 class="mb-2 text-sm font-medium text-orange-800 dark:text-orange-300 sm:text-base">
            Linux / WSL2 注意事项
          </h6>
          <ul class="space-y-1 text-xs text-orange-700 dark:text-orange-300 sm:text-sm">
            <li>• WSL2 用户建议在 Linux 子系统中安装，而不是 Windows</li>
            <li>• 使用 nvm 可以避免权限问题</li>
            <li>• 确保 shell 配置文件正确加载了 nvm</li>
          </ul>
        </div>
      </div>
      <VerifyInstall terminal="终端" />
    </div>
  </div>
</template>

<script setup lang="ts">
import VerifyInstall from './VerifyInstall.vue'

defineProps({
  platform: {
    type: String,
    required: true,
    validator: (value: string) => ['windows', 'macos', 'linux'].includes(value)
  },
  stepNumber: {
    type: [Number, String],
    default: 1
  },
  toolName: {
    type: String,
    default: 'CLI 工具'
  }
})
</script>
