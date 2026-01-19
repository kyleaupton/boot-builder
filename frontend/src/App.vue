<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useDrivesStore, useSourceStore, useJobStore } from '@/stores'
import { useDragDrop } from '@/composables'
import type { AppState } from '@/types'

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

// File drop handling
useDragDrop({
  accept: ['.iso', '.img'],
  onDrop: (files) => {
    if (files.length > 0) {
      sourceStore.setSource(files[0])
    }
  },
})

// Format drive size for display
function formatSize(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) return `${gb.toFixed(1)} GB`
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(0)} MB`
}

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
  drivesStore.startAutoRefresh()
  sourceStore.loadInstallers()
  jobStore.subscribeToEvents()
})
</script>

<template>
  <div class="container" data-file-drop-target>
    <div class="debug-panel">
      <h1>OS Install Maker</h1>
      <p class="state-badge">State: <code>{{ appState }}</code></p>

      <!-- Source Section -->
      <section class="section">
        <h2>Source</h2>
        <div v-if="sourceStore.hasSource" class="source-info">
          <p><strong>File:</strong> {{ sourceStore.filename }}</p>
          <p><strong>Installer:</strong> {{ sourceStore.detectedInstaller?.Name ?? 'Unknown' }}</p>
          <button class="btn-secondary" @click="sourceStore.clearSource">Clear</button>
        </div>
        <div v-else class="dropzone-placeholder">
          <p>Drop an ISO file here</p>
          <p class="hint">or the window will accept drops anywhere</p>
        </div>
      </section>

      <!-- Drive Section -->
      <section class="section">
        <h2>Target Drive</h2>
        <div v-if="drivesStore.hasDrives" class="drive-list">
          <label
            v-for="drive in drivesStore.removableDrives"
            :key="drive.Device"
            class="drive-item"
            :class="{ selected: drivesStore.selectedDriveId === drive.Device }"
          >
            <input
              type="radio"
              :value="drive.Device"
              :checked="drivesStore.selectedDriveId === drive.Device"
              @change="drivesStore.selectDrive(drive.Device)"
            />
            <span class="drive-info">
              <span class="drive-name">{{ drive.Model || 'Unknown Drive' }}</span>
              <span class="drive-meta">{{ drive.Device }} · {{ formatSize(drive.SizeBytes) }}</span>
            </span>
          </label>
        </div>
        <p v-else class="empty-state">No removable drives detected</p>
        <p v-if="drivesStore.isLoading" class="loading">Refreshing...</p>
      </section>

      <!-- Progress Section -->
      <section v-if="appState === 'in-progress'" class="section">
        <h2>Progress</h2>
        <div class="progress-info">
          <p v-if="jobStore.currentStep"><strong>Step:</strong> {{ jobStore.currentStep }}</p>
          <p v-if="jobStore.currentMessage">{{ jobStore.currentMessage }}</p>
          <div class="progress-bar-container">
            <div class="progress-bar" :style="{ width: `${jobStore.progress}%` }"></div>
          </div>
          <p class="progress-percent">{{ jobStore.progress }}%</p>
        </div>
      </section>

      <!-- Complete Section -->
      <section v-if="appState === 'complete'" class="section success">
        <h2>Complete!</h2>
        <p>Your bootable USB has been created successfully.</p>
      </section>

      <!-- Error Section -->
      <section v-if="jobStore.error" class="section error">
        <h2>Error</h2>
        <p>{{ jobStore.error }}</p>
      </section>

      <!-- Actions -->
      <div class="actions">
        <button
          v-if="appState === 'ready'"
          class="btn-primary"
          :disabled="jobStore.isStarting"
          @click="handleStartJob"
        >
          {{ jobStore.isStarting ? 'Starting...' : 'Flash Drive' }}
        </button>

        <button
          v-if="appState === 'complete' || appState === 'error'"
          class="btn-primary"
          @click="handleReset"
        >
          Start Over
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  min-height: 100vh;
  background: var(--background);
  color: var(--foreground);
}

.debug-panel {
  width: 100%;
  max-width: 500px;
}

h1 {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.state-badge {
  margin-bottom: 1.5rem;
  color: var(--muted-foreground);
}

.state-badge code {
  background: var(--muted);
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius);
  font-size: 0.875rem;
}

.section {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--card);
}

.section h2 {
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted-foreground);
  margin-bottom: 0.75rem;
}

.source-info p {
  margin: 0.25rem 0;
}

.dropzone-placeholder {
  text-align: center;
  padding: 2rem;
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  color: var(--muted-foreground);
}

.dropzone-placeholder .hint {
  font-size: 0.75rem;
  margin-top: 0.5rem;
}

.drive-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.drive-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: border-color 0.2s;
}

.drive-item:hover {
  border-color: var(--primary);
}

.drive-item.selected {
  border-color: var(--primary);
  background: var(--accent);
}

.drive-info {
  display: flex;
  flex-direction: column;
}

.drive-name {
  font-weight: 500;
}

.drive-meta {
  font-size: 0.75rem;
  color: var(--muted-foreground);
}

.empty-state {
  color: var(--muted-foreground);
  font-style: italic;
}

.loading {
  font-size: 0.75rem;
  color: var(--muted-foreground);
  margin-top: 0.5rem;
}

.progress-info p {
  margin: 0.25rem 0;
}

.progress-bar-container {
  height: 8px;
  background: var(--muted);
  border-radius: 4px;
  overflow: hidden;
  margin: 0.75rem 0;
}

.progress-bar {
  height: 100%;
  background: var(--primary);
  transition: width 0.3s ease;
}

.progress-percent {
  font-size: 0.875rem;
  color: var(--muted-foreground);
}

.section.success {
  border-color: var(--chart-2);
  background: color-mix(in oklch, var(--chart-2) 10%, transparent);
}

.section.error {
  border-color: var(--destructive);
  background: color-mix(in oklch, var(--destructive) 10%, transparent);
}

.actions {
  margin-top: 1.5rem;
}

.btn-primary,
.btn-secondary {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn-primary {
  background: var(--primary);
  color: var(--primary-foreground);
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--muted);
  color: var(--muted-foreground);
  margin-top: 0.5rem;
}

.btn-secondary:hover {
  background: var(--accent);
}
</style>
