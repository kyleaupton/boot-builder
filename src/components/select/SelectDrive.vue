<template>
  <InputGroup :class="{ 'p-disabled': isFlashing }">
    <InputGroupAddon>
      <font-awesome-icon :icon="['fas', 'hard-drive']" />
    </InputGroupAddon>

    <Dropdown
      v-model="selectedDrive"
      :options="dropdownOptions"
      option-label="label"
      option-value="value"
      placeholder="Select target"
    />
  </InputGroup>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { storeToRefs } from 'pinia';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import prettyBytes from 'pretty-bytes';
import { useDisksStore, useFlashStore } from '@/stores';

const { items } = storeToRefs(useDisksStore());
const { selectedDrive, isFlashing } = storeToRefs(useFlashStore());

const dropdownOptions = computed(() => {
  return items.value.map((drive) => ({
    label: `${drive.description} (${prettyBytes(drive.size ?? 0)})`,
    value: drive.device,
  }));
});

watch(items, () => {
  if (
    selectedDrive.value !== undefined &&
    !items.value.find((drive) => drive.devicePath === selectedDrive.value)
  ) {
    selectedDrive.value = undefined;
  }
});
</script>

<style scoped></style>
