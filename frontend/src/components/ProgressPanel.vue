<script setup lang="ts">
import { computed } from 'vue'
import { useJobStore } from '@/stores'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'

const jobStore = useJobStore()

const progress = computed(() => jobStore.progress)
const currentStep = computed(() => jobStore.currentStep)
const currentMessage = computed(() => jobStore.currentMessage)
const isActive = computed(() => jobStore.isRunning || jobStore.isPending)
</script>

<template>
  <Card class="progress-card">
    <CardHeader class="progress-header">
      <div class="progress-header-row">
        <CardTitle class="progress-title">Progress</CardTitle>
        <div v-if="isActive" class="active-indicator">
          <div class="pulse-dot" />
        </div>
      </div>
    </CardHeader>
    <CardContent class="progress-content">
      <!-- Step info -->
      <div v-if="currentStep" class="step-info">
        <span class="step-label">{{ currentStep }}</span>
      </div>

      <!-- Progress bar -->
      <div class="progress-bar-wrapper">
        <Progress :model-value="progress" class="progress-bar" />
        <span class="progress-percent">{{ progress.toFixed(1) }}%</span>
      </div>

      <!-- Status message -->
      <p v-if="currentMessage" class="status-message">
        {{ currentMessage }}
      </p>
    </CardContent>
  </Card>
</template>

<style scoped>
.progress-card {
  overflow: hidden;
}

.progress-header {
  padding-bottom: 0;
}

.progress-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.progress-title {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted-foreground);
}

.active-indicator {
  display: flex;
  align-items: center;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background: var(--primary);
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(0.85);
  }
}

.progress-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.step-info {
  display: flex;
  align-items: center;
}

.step-label {
  font-weight: 500;
  font-size: 0.875rem;
}

.progress-bar-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.progress-bar {
  flex: 1;
  height: 8px;
}

.progress-percent {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--muted-foreground);
  min-width: 3rem;
  text-align: right;
}

.status-message {
  font-size: 0.875rem;
  color: var(--muted-foreground);
  margin: 0;
}
</style>
