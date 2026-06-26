<template>
  <div
    class="terminal h-full w-full overflow-y-auto p-4 md:p-6 cursor-text font-mono text-sm leading-relaxed relative"
    :style="{ background: 'var(--rp-base)', color: 'var(--rp-text)' }"
    @click="focusInput"
  >
    <!-- Matrix rain canvas overlay -->
    <canvas
      v-if="matrixActive"
      ref="matrixCanvas"
      class="absolute inset-0 z-10 pointer-events-none"
    ></canvas>

    <!-- Banner on first load -->
    <div v-if="lines.length === 0 && !initialBannerShown" class="mb-8 mt-4">
      <pre class="text-[var(--rp-rose)] leading-tight text-xs sm:text-sm md:text-base"
      > ██████╗  █████╗ ████████╗██╗  ██╗██╗     ███████╗███████╗███████╗
██╔═══██╗██╔══██╗╚══██╔══╝██║  ██║██║     ██╔════╝██╔════╝██╔════╝
██║   ██║███████║   ██║   ███████║██║     █████╗  ███████╗███████╗
██║   ██║██╔══██║   ██║   ██╔══██║██║     ██╔══╝  ╚════██║╚════██║
╚██████╔╝██║  ██║   ██║   ██║  ██║███████╗███████╗███████║███████║
 ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚══════╝</pre>
      <div class="mt-4" style="color: var(--rp-subtle)">
        <div>welcome to my terminal. type <span style="color: var(--rp-foam)">help</span> to get started.</div>
        <div class="mt-2">connected to <span style="color: var(--rp-gold)">oathless.dev</span></div>
      </div>
    </div>

    <!-- Output lines -->
    <div v-for="(line, i) in lines" :key="i" class="mb-1">
      <!-- User input line -->
      <div v-if="line.type === 'input'" class="flex">
        <span style="color: var(--rp-iris)">❯&nbsp;</span>
        <span style="color: var(--rp-text)">{{ line.text }}</span>
      </div>

      <!-- Command output -->
      <pre
        v-else-if="line.type === 'output'"
        class="whitespace-pre-wrap m-0 font-mono text-sm leading-relaxed"
        style="color: var(--rp-subtle)"
        v-text="line.text"
      ></pre>

      <!-- Error output -->
      <pre
        v-else-if="line.type === 'error'"
        class="whitespace-pre-wrap m-0 font-mono text-sm leading-relaxed"
        style="color: var(--rp-love)"
        v-text="line.text"
      ></pre>
    </div>

    <!-- Loading indicator -->
    <div v-if="loading" class="flex" style="color: var(--rp-gold)">
      <span>…</span>
    </div>

    <!-- Input line -->
    <div v-if="!loading" class="flex items-start relative">
      <span class="flex-shrink-0" style="color: var(--rp-iris)">❯&nbsp;</span>
      <div class="flex-1 relative">
        <input
          ref="hiddenInput"
          v-model="inputValue"
          class="w-full bg-transparent border-none outline-none font-mono text-sm leading-relaxed"
          :style="{ color: 'var(--rp-text)', caretColor: 'var(--rp-iris)' }"
          spellcheck="false"
          autocomplete="off"
          autocapitalize="off"
          @keydown="handleKeydown"
        />
        <!-- Tab completion suggestions -->
        <div
          v-if="showSuggestions && suggestions.length"
          class="absolute left-0 mt-1 z-20 rounded border"
          :style="{
            background: 'var(--rp-surface)',
            borderColor: 'var(--rp-overlay)',
            minWidth: '200px'
          }"
        >
          <div
            v-for="(s, idx) in suggestions"
            :key="s"
            class="px-3 py-1 cursor-pointer font-mono text-sm"
            :class="{ 'bg-[var(--rp-overlay)]': idx === selectedSuggestion }"
            :style="{ color: idx === selectedSuggestion ? 'var(--rp-text)' : 'var(--rp-subtle)' }"
            @mousedown.prevent="selectSuggestion(s)"
          >
            {{ s }}
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom spacer for scroll -->
    <div ref="bottomRef"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'

const hiddenInput = ref(null)
const bottomRef = ref(null)
const matrixCanvas = ref(null)
const inputValue = ref('')
const lines = ref([])
const loading = ref(false)
const initialBannerShown = ref(false)
const history = ref([])
const historyIndex = ref(-1)
const tempInput = ref('')
const matrixActive = ref(false)
const matrixTimer = ref(null)

// Tab completion state
const availableCommands = ref([])
const showSuggestions = ref(false)
const suggestions = ref([])
const selectedSuggestion = ref(0)
const tabWordStart = ref(0)

function focusInput() {
  hiddenInput.value?.focus()
}

function scrollToBottom() {
  nextTick(() => {
    bottomRef.value?.scrollIntoView({ behavior: 'smooth' })
  })
}

function doCompletion() {
  const input = hiddenInput.value
  if (!input) return

  const pos = input.selectionStart
  const val = inputValue.value
  // Find word boundaries around cursor
  let start = pos
  while (start > 0 && val[start - 1] !== ' ') start--
  let end = pos
  while (end < val.length && val[end] !== ' ') end++

  const partial = val.slice(start, end).toLowerCase()
  const matches = availableCommands.value.filter(c => c.startsWith(partial))

  if (matches.length === 0) {
    showSuggestions.value = false
    return
  }

  if (matches.length === 1) {
    // Single match: autocomplete
    const before = val.slice(0, start)
    const after = val.slice(end)
    inputValue.value = before + matches[0] + after
    showSuggestions.value = false
    // ponytail: setTimeout for cursor position after Vue reactivity
    nextTick(() => {
      input.setSelectionRange(start + matches[0].length, start + matches[0].length)
    })
    return
  }

  // Multiple matches: show suggestions, cycle on repeated Tab
  if (showSuggestions.value && suggestions.value.length === matches.length) {
    selectedSuggestion.value = (selectedSuggestion.value + 1) % matches.length
    return
  }

  suggestions.value = matches
  selectedSuggestion.value = 0
  tabWordStart.value = start
  showSuggestions.value = true
}

function selectSuggestion(cmd) {
  const val = inputValue.value
  const start = tabWordStart.value
  const input = hiddenInput.value
  let end = input ? input.selectionStart : val.length
  while (end < val.length && val[end] !== ' ') end++

  inputValue.value = val.slice(0, start) + cmd + val.slice(end)
  showSuggestions.value = false
  nextTick(() => {
    if (input) input.setSelectionRange(start + cmd.length, start + cmd.length)
    focusInput()
  })
}

function dismissSuggestions() {
  showSuggestions.value = false
}

// Matrix rain animation
function startMatrixRain() {
  const prevTheme = localStorage.getItem('theme') || 'rose-pine'
  matrixActive.value = true
  applyTheme('matrix')

  nextTick(() => {
    const canvas = matrixCanvas.value
    if (!canvas) return

    const container = canvas.parentElement
    const dpr = window.devicePixelRatio || 1
    canvas.width = container.clientWidth * dpr
    canvas.height = container.clientHeight * dpr
    canvas.style.width = container.clientWidth + 'px'
    canvas.style.height = container.clientHeight + 'px'

    const ctx = canvas.getContext('2d')
    ctx.scale(dpr, dpr)
    const chars = '日ﾊﾐﾋｰｳｼﾅﾓﾆｻﾜﾂｵﾘｱﾎﾃﾏｹﾒｴｶｷﾑﾕﾗｾﾈｽﾀﾇﾍ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    const fontSize = 14
    const columns = Math.floor(container.clientWidth / fontSize)
    const drops = new Array(columns).fill(0)

    const draw = () => {
      ctx.fillStyle = 'rgba(0, 0, 0, 0.05)'
      ctx.fillRect(0, 0, container.clientWidth, container.clientHeight)
      ctx.fillStyle = '#00ff41'
      ctx.font = `${fontSize}px monospace`

      for (let i = 0; i < drops.length; i++) {
        const char = chars[Math.floor(Math.random() * chars.length)]
        ctx.fillText(char, i * fontSize, drops[i] * fontSize)

        if (drops[i] * fontSize > container.clientHeight && Math.random() > 0.975) {
          drops[i] = 0
        }
        drops[i]++
      }
    }

    const interval = setInterval(draw, 40)

    matrixTimer.value = setTimeout(() => {
      clearInterval(interval)
      let opacity = 0
      const fadeOut = setInterval(() => {
        opacity += 0.05
        ctx.fillStyle = `rgba(0, 0, 0, ${opacity})`
        ctx.fillRect(0, 0, container.clientWidth, container.clientHeight)
        if (opacity >= 1) {
          clearInterval(fadeOut)
          applyTheme(prevTheme)
          lines.value = []
          matrixActive.value = false
        }
      }, 30)
    }, 8000)
  })
}

async function submitCommand(cmd) {
  const trimmed = cmd.trim()
  if (!trimmed) return

  // Add to history
  history.value.push(trimmed)
  historyIndex.value = history.value.length

  // Show the command in output
  lines.value.push({ type: 'input', text: trimmed })

  // Handle clear locally
  if (trimmed === 'clear') {
    lines.value = []
    inputValue.value = ''
    return
  }

  // Handle history locally
  if (trimmed === 'history') {
    if (history.value.length === 0) {
      lines.value.push({ type: 'output', text: '(no commands yet)' })
    } else {
      const output = history.value.map((h, i) => `  ${i + 1}  ${h}`).join('\n')
      lines.value.push({ type: 'output', text: output })
    }
    inputValue.value = ''
    return
  }

  // Handle history | grep locally (history lives in browser)
  if (trimmed.startsWith('history | grep ')) {
    const pattern = trimmed.slice('history | grep '.length).trim().toLowerCase()
    const output = history.value
      .map((h, i) => `  ${i + 1}  ${h}`)
      .filter(line => line.toLowerCase().includes(pattern))
      .join('\n')
    lines.value.push({ type: 'output', text: output || '(no matches)' })
    inputValue.value = ''
    return
  }

  loading.value = true
  scrollToBottom()

  try {
    const res = await fetch('/api/command', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: trimmed })
    })

    const data = await res.json()

    if (data.type === 'clear') {
      lines.value = []
    } else if (data.type === 'matrix') {
      lines.value.push({ type: 'output', text: data.output })
      startMatrixRain()
    } else if (data.type && data.type.startsWith('theme:')) {
      const theme = data.type.split(':')[1]
      applyTheme(theme)
      lines.value.push({ type: 'output', text: data.output })
    } else if (data.type === 'error') {
      lines.value.push({ type: 'error', text: data.output })
    } else {
      lines.value.push({ type: 'output', text: data.output })
    }
  } catch {
    lines.value.push({ type: 'error', text: 'connection lost — is the backend running?' })
  } finally {
    loading.value = false
    inputValue.value = ''
    scrollToBottom()
    // Re-focus after the v-if re-mounts the input
    await nextTick()
    focusInput()
  }
}

function handleKeydown(e) {
  // Suggestion navigation
  if (showSuggestions.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selectedSuggestion.value = (selectedSuggestion.value + 1) % suggestions.value.length
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      selectedSuggestion.value = (selectedSuggestion.value - 1 + suggestions.value.length) % suggestions.value.length
      return
    }
    if (e.key === 'Escape') {
      e.preventDefault()
      dismissSuggestions()
      return
    }
    if (e.key === 'Tab') {
      e.preventDefault()
      selectSuggestion(suggestions.value[selectedSuggestion.value])
      return
    }
    dismissSuggestions()
  }

  if (e.key === 'Enter') {
    e.preventDefault()
    dismissSuggestions()
    submitCommand(inputValue.value)
    return
  }

  if (e.key === 'Tab') {
    e.preventDefault()
    doCompletion()
    return
  }

  if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (history.value.length > 0 && historyIndex.value > 0) {
      if (historyIndex.value === history.value.length) {
        tempInput.value = inputValue.value
      }
      historyIndex.value--
      inputValue.value = history.value[historyIndex.value]
    }
    return
  }

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (historyIndex.value < history.value.length - 1) {
      historyIndex.value++
      inputValue.value = history.value[historyIndex.value]
    } else if (historyIndex.value === history.value.length - 1) {
      historyIndex.value = history.value.length
      inputValue.value = tempInput.value
    }
    return
  }

  // Ctrl+L = clear
  if (e.ctrlKey && e.key === 'l') {
    e.preventDefault()
    lines.value = []
    return
  }
}

// Theme palettes
const themes = {
  'rose-pine': {
    base: '#191724', surface: '#1f1d2e', overlay: '#26233a',
    muted: '#6e6a86', subtle: '#908caa', text: '#e0def4',
    love: '#eb6f92', gold: '#f6c177', rose: '#ebbcba',
    pine: '#31748f', foam: '#9ccfd8', iris: '#c4a7e7'
  },
  green: {
    base: '#0d1117', surface: '#161b22', overlay: '#21262d',
    muted: '#484f58', subtle: '#8b949e', text: '#c9d1d9',
    love: '#ff7b72', gold: '#d2991d', rose: '#79c0ff',
    pine: '#3fb950', foam: '#56d364', iris: '#3fb950'
  },
  amber: {
    base: '#1a1100', surface: '#241a00', overlay: '#2d2200',
    muted: '#665000', subtle: '#997a00', text: '#ffd866',
    love: '#ff6600', gold: '#ffaa00', rose: '#ffcc33',
    pine: '#ff8800', foam: '#ffbb33', iris: '#ff9900'
  },
  matrix: {
    base: '#0d0d0d', surface: '#1a1a1a', overlay: '#262626',
    muted: '#0a3d0a', subtle: '#0f5c0f', text: '#00ff41',
    love: '#ff0040', gold: '#ccff00', rose: '#00cc33',
    pine: '#009933', foam: '#33ff66', iris: '#00ff41'
  }
}

function applyTheme(name) {
  const t = themes[name] || themes['rose-pine']
  const root = document.documentElement
  for (const [key, value] of Object.entries(t)) {
    root.style.setProperty(`--rp-${key}`, value)
  }
  localStorage.setItem('theme', name)
}

// Load saved theme on mount
const savedTheme = localStorage.getItem('theme')
if (savedTheme && themes[savedTheme]) {
  applyTheme(savedTheme)
}

onMounted(async () => {
  // Fetch available commands for tab completion
  try {
    const res = await fetch('/api/commands')
    availableCommands.value = await res.json()
  } catch {
    // ponytail: fallback to empty, tab just won't work
  }

  focusInput()
  scrollToBottom()
})
</script>
