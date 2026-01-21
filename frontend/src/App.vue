<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useColorMode } from '@vueuse/core'
import { useDrivesStore, useSourceStore, useJobStore } from '@/stores'
import type { AppState } from '@/types'
import { Button } from '@/components/ui/button'
import SourceDropzone from '@/components/SourceDropzone.vue'
import DriveSelector from '@/components/DriveSelector.vue'
import ProgressPanel from '@/components/ProgressPanel.vue'
import StatusAlert from '@/components/StatusAlert.vue'
import FlashButton from '@/components/FlashButton.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import { Toaster } from '@/components/ui/sonner'

// Initialize stores
const drivesStore = useDrivesStore()
const sourceStore = useSourceStore()
const jobStore = useJobStore()

// Combine store states into overall app state
const appState = computed((): AppState => {
  if (jobStore.isFailed || jobStore.error) return 'error'
  if (jobStore.isRunning || jobStore.isPending) return 'in-progress'
  if (jobStore.isComplete) return 'complete'
  if (sourceStore.hasSource && drivesStore.selectedDrive) return 'ready'
  if (sourceStore.hasSource) return 'source-only'
  return 'empty'
})

// Step number for the stepper based on app state
const currentStepNumber = computed((): 1 | 2 | 3 => {
  switch (appState.value) {
    case 'empty':
    case 'source-only':
      return sourceStore.hasSource ? 2 : 1
    case 'ready':
      return 2
    case 'in-progress':
    case 'complete':
    case 'error':
      return 3
    default:
      return 1
  }
})

// Start job handler
async function handleStartJob() {
  if (!sourceStore.source || !drivesStore.selectedDrive || !sourceStore.detectedInstaller) {
    return
  }

  try {
    await jobStore.startJob({
      InstallerID: sourceStore.detectedInstaller.ID,
      SourceLocal: sourceStore.source.path,
      DriveID: drivesStore.selectedDrive.Device,
    })
  } catch (e) {
    console.error('Failed to start job:', e)
  }
}

// Reset to start over
function handleReset() {
  jobStore.clearCurrentJob()
  sourceStore.clearSource()
  drivesStore.selectDrive(null)
}

onMounted(() => {
  const mode = useColorMode()
  mode.value = 'dark';

  drivesStore.startAutoRefresh()
  sourceStore.loadInstallers()
  jobStore.subscribeToEvents()
})
</script>

<template>
  <div class="app-container">
    <header class="app-header">
      <h1 class="app-title">Boot Builder</h1>
      <StepIndicator :current-step="currentStepNumber" />
    </header>

    <Toaster position="bottom-center" />

    <!-- Main Content -->
    <main class="app-main">
      <SourceDropzone />
      <DriveSelector
        v-if="sourceStore.hasSource && appState !== 'in-progress' && appState !== 'complete' && appState !== 'error'"
      />
      <ProgressPanel v-if="appState === 'in-progress'" />
      <StatusAlert
        v-if="appState === 'complete' || appState === 'error'"
        :status="appState"
        :error="jobStore.error"
      />
    </main>

    <footer class="app-footer">
      <FlashButton
        v-if="appState === 'ready'"
        :loading="jobStore.isStarting"
        @click="handleStartJob"
      />

      <Button
        v-if="appState === 'complete' || appState === 'error'"
        size="lg"
        variant="secondary"
        class="reset-button"
        @click="handleReset"
      >
        Start Over
      </Button>
    </footer>
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--background);
  color: var(--foreground);
}

.app-header {
  padding: 1rem 1.5rem;
  text-align: center;
  --wails-draggable: drag;
}

.app-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0;
  color: var(--foreground);
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0 1.5rem;
  max-width: 500px;
  width: 100%;
  margin: 0 auto;
}

.app-footer {
  padding: 1.5rem;
  max-width: 500px;
  width: 100%;
  margin: 0 auto;
}

.reset-button {
  width: 100%;
  height: 3rem;
  font-size: 1rem;
}
</style>
