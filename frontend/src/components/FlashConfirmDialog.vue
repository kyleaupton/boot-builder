<script setup lang="ts">
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

defineProps<{
  open: boolean
  driveName: string
  driveSize: string
}>()

defineEmits<{
  'update:open': [value: boolean]
  confirm: []
}>()
</script>

<template>
  <AlertDialog :open="open" @update:open="$emit('update:open', $event)">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Erase drive and flash?</AlertDialogTitle>
        <AlertDialogDescription>
          All data on <strong>{{ driveName }}</strong> ({{ driveSize }}) will be permanently erased. This cannot be undone.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="$emit('update:open', false)">Cancel</AlertDialogCancel>
        <AlertDialogAction
          class="destructive-action"
          @click="$emit('confirm')"
        >
          Continue
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<style scoped>
.destructive-action {
  background: var(--destructive);
  color: white;
}

.destructive-action:hover {
  background: var(--destructive);
  opacity: 0.9;
}
</style>
