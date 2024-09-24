<template>
  <InputGroup :class="{ 'p-disabled': isFlashing }">
    <InputGroupAddon>
      <font-awesome-icon :icon="['fas', 'display']" />
    </InputGroupAddon>

    <Dropdown
      v-model="selectedOs"
      :options="options"
      option-label="label"
      option-value="value"
      placeholder="Select OS"
    >
      <template #value="slotProps">
        <!--  -->
        <div v-if="slotProps.value" class="os-selector-option">
          <font-awesome-icon
            :icon="options.find((x) => x.value === slotProps.value)?.icon || []"
          />
          <div>
            {{
              options.find((x) => x.value === slotProps.value)?.label ||
              slotProps.value
            }}
          </div>
        </div>
        <span v-else>
          {{ slotProps.placeholder }}
        </span>
      </template>
      <template #option="slotProps">
        <div class="os-selector-option">
          <font-awesome-icon :icon="slotProps.option.icon" />
          <div>{{ slotProps.option.label }}</div>
        </div>
      </template>
    </Dropdown>
  </InputGroup>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import { useFlashStore } from '@/stores';

const { selectedOs, isFlashing } = storeToRefs(useFlashStore());

const options = [
  {
    label: 'Windows',
    value: 'windows',
    icon: ['fab', 'windows'],
  },
  {
    label: 'macOS',
    value: 'macos',
    icon: ['fab', 'apple'],
  },
];
</script>

<style scoped>
.os-selector-option {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
</style>
