<template>
  <div
    class="terminal h-full w-full overflow-y-auto p-4 md:p-6 cursor-text font-mono text-sm leading-relaxed"
    :style="{ background: 'var(--rp-base)', color: 'var(--rp-text)' }"
    @click="focusInput"
  >
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
        <span style="color: var(--rp-iris)">❯ </span>
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
    <div v-if="!loading" class="flex items-start">
      <span class="flex-shrink-0" style="color: var(--rp-iris)">❯ </span>
      <span class="flex-1 relative">
        <span>{{ inputLeft }}</span><span class="cursor-blink" style="color: var(--rp-iris)">█</span>
        <!-- Hidden input for capturing keystrokes -->
        <input
          ref="hiddenInput"
          v-model="inputValue"
          class="absolute inset-0 opacity-0 cursor-text"
          style="caret-color: transparent"
          spellcheck="false"
          autocomplete="off"
          autocapitalize="off"
          @keydown="handleKeydown"
          @keyup="handleKeyup"
        />
      </span>
    </div>

    <!-- Bottom spacer for scroll -->
    <div ref="bottomRef"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, computed } from 'vue'

const hiddenInput = ref(null)
const bottomRef = ref(null)
const inputValue = ref('')
const lines = ref([])
const loading = ref(false)
const initialBannerShown = ref(true)
const history = ref([])
const historyIndex = ref(-1)
const tempInput = ref('')

// Show input starting from the end so cursor sits at end
const inputLeft = computed(() => inputValue.value)

function focusInput() {
  hiddenInput.value?.focus()
}

function scrollToBottom() {
  nextTick(() => {
    bottomRef.value?.scrollIntoView({ behavior: 'smooth' })
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
    } else if (data.type && data.type.startsWith('theme:')) {
      lines.value.push({ type: 'output', text: data.output })
      // Theme change is handled by CSS variable swaps
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
  if (e.key === 'Enter') {
    e.preventDefault()
    submitCommand(inputValue.value)
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

function handleKeyup() {
  scrollToBottom()
}

onMounted(() => {
  focusInput()
  scrollToBottom()
})
</script>
