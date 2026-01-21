<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import FlashConfirmDialog from '@/components/FlashConfirmDialog.vue'
import { useDrivesStore } from '@/stores'

const props = defineProps<{
  loading?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  click: []
}>()

const drivesStore = useDrivesStore()
const dialogOpen = ref(false)

function handleClick() {
  if (props.loading || props.disabled) return
  dialogOpen.value = true
}

function handleConfirm() {
  dialogOpen.value = false
  emit('click')
}

function formatSize(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) return `${gb.toFixed(1)} GB`
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(0)} MB`
}
</script>

<template>
  <FlashConfirmDialog
    v-model:open="dialogOpen"
    :drive-name="drivesStore.selectedDrive?.Model || 'Unknown Drive'"
    :drive-size="drivesStore.selectedDrive ? formatSize(drivesStore.selectedDrive.SizeBytes) : ''"
    @confirm="handleConfirm"
  />

  <Button
    size="lg"
    class="flash-button"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <template v-if="loading">
      <div class="spinner" />
      <span>Starting...</span>
    </template>
    <template v-else>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
      </svg>
      <span>Flash Drive</span>
    </template>
  </Button>
</template>

<style scoped>
.flash-button {
  width: 100%;
  height: 3rem;
  font-size: 1rem;
  font-weight: 600;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
