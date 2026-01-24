<script setup lang="ts">
import { computed } from 'vue'
import { useJobStore } from '@/stores'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Check, Circle, Loader2, X } from 'lucide-vue-next'

const jobStore = useJobStore()
const steps = computed(() => jobStore.steps)
const isActive = computed(() => jobStore.isRunning || jobStore.isPending)
</script>

<template>
  <Card class="progress-card">
    <CardHeader class="progress-header">
      <div class="progress-header-row">
        <CardTitle class="progress-title">Creating Bootable USB</CardTitle>
        <div v-if="isActive" class="active-indicator">
          <div class="pulse-dot" />
        </div>
      </div>
    </CardHeader>
    <CardContent class="progress-content">
      <div class="steps-list">
        <div
          v-for="step in steps"
          :key="step.key"
          class="step-item"
          :class="step.status"
        >
          <!-- Step header row -->
          <div class="step-header">
            <span class="step-icon">
              <Check v-if="step.status === 'completed'" class="icon-completed" />
              <Loader2 v-else-if="step.status === 'running'" class="icon-running" />
              <X v-else-if="step.status === 'failed'" class="icon-failed" />
              <Circle v-else class="icon-pending" />
            </span>
            <span class="step-name">{{ step.name }}</span>
          </div>

          <!-- Expanded details for running step with progress -->
          <div
            v-if="step.status === 'running' && step.hasProgress"
            class="step-details"
          >
            <div class="step-progress-row">
              <Progress :model-value="step.progress" class="step-progress-bar" />
              <span class="step-percent">{{ step.progress.toFixed(1) }}%</span>
            </div>
            <p v-if="step.message" class="step-message">{{ step.message }}</p>
          </div>
        </div>
      </div>
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

/* Step list styles */
.steps-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.step-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.step-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

.step-icon :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.icon-completed {
  color: hsl(var(--primary));
}

.icon-running {
  color: hsl(var(--primary));
  animation: spin 1s linear infinite;
}

.icon-failed {
  color: hsl(var(--destructive));
}

.icon-pending {
  color: hsl(var(--muted-foreground));
}

.step-name {
  font-size: 0.875rem;
}

.step-item.completed .step-name {
  color: hsl(var(--muted-foreground));
}

.step-item.pending .step-name {
  color: hsl(var(--muted-foreground));
}

.step-item.running .step-name {
  font-weight: 500;
}

.step-details {
  margin-left: 2rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.step-progress-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.step-progress-bar {
  flex: 1;
  height: 6px;
}

.step-percent {
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
  min-width: 3rem;
  text-align: right;
  color: hsl(var(--muted-foreground));
}

.step-message {
  font-size: 0.75rem;
  color: hsl(var(--muted-foreground));
  margin: 0;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
