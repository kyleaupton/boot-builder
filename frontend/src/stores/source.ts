import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { ListInstallers } from '@bindings/boot-builder/internal/service/jobsservice'
import type { InstallerMeta, Target, SourceInfo } from '@/types'

export const useSourceStore = defineStore('source', () => {
  // State
  const source = ref<SourceInfo | null>(null)
  const installers = ref<InstallerMeta[]>([])
  const selectedTargetIndex = ref<number>(0)
  const isAnalyzing = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const hasSource = computed(() => source.value !== null)

  const filename = computed(() => source.value?.filename ?? null)

  const detectedInstaller = computed(() => {
    if (!source.value?.installerID) return null
    return installers.value.find((i) => i.ID === source.value!.installerID) ?? null
  })

  const selectedTarget = computed((): Target | null => {
    if (!source.value?.detectedTargets?.length) return null
    return source.value.detectedTargets[selectedTargetIndex.value] ?? null
  })

  const fileSizeFormatted = computed(() => {
    if (!source.value?.sizeBytes) return null
    const bytes = source.value.sizeBytes
    const gb = bytes / (1024 * 1024 * 1024)
    if (gb >= 1) {
      return `${gb.toFixed(1)} GB`
    }
    const mb = bytes / (1024 * 1024)
    return `${mb.toFixed(0)} MB`
  })

  // Actions
  async function loadInstallers(): Promise<void> {
    try {
      installers.value = await ListInstallers()
    } catch (e) {
      console.error('Failed to load installers:', e)
    }
  }

  function detectInstallerFromFilename(filename: string): InstallerMeta | null {
    const lower = filename.toLowerCase()

    // Pattern matching - can be extended as more installers are added
    if (lower.includes('ubuntu') || lower.includes('linux')) {
      return installers.value.find((i) => i.ID === 'linux') ?? null
    }
    if (lower.includes('windows') || lower.includes('win10') || lower.includes('win11')) {
      return installers.value.find((i) => i.ID === 'windows') ?? null
    }

    return null
  }

  async function setSource(path: string): Promise<void> {
    isAnalyzing.value = true
    error.value = null

    try {
      // Ensure installers are loaded
      if (installers.value.length === 0) {
        await loadInstallers()
      }

      const filename = path.split('/').pop() ?? path
      const detected = detectInstallerFromFilename(filename)

      source.value = {
        path,
        filename,
        sizeBytes: 0, // Could be populated by backend in future
        installerID: detected?.ID ?? null,
        installerName: detected?.Name ?? null,
        detectedTargets: detected?.Targets ?? [],
      }

      selectedTargetIndex.value = 0
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to analyze source'
      source.value = null
    } finally {
      isAnalyzing.value = false
    }
  }

  function selectTarget(index: number): void {
    if (source.value?.detectedTargets && index < source.value.detectedTargets.length) {
      selectedTargetIndex.value = index
    }
  }

  function clearSource(): void {
    source.value = null
    selectedTargetIndex.value = 0
    error.value = null
  }

  function $reset(): void {
    source.value = null
    selectedTargetIndex.value = 0
    isAnalyzing.value = false
    error.value = null
  }

  return {
    // State
    source,
    installers,
    selectedTargetIndex,
    isAnalyzing,
    error,
    // Getters
    hasSource,
    filename,
    detectedInstaller,
    selectedTarget,
    fileSizeFormatted,
    // Actions
    loadInstallers,
    setSource,
    selectTarget,
    clearSource,
    $reset,
  }
})
