<template>
  <div class="source-wrapper">
    <Button
      v-if="!selectedSource"
      label="Choose file"
      icon="pi pi-upload"
      :disabled="!selectedOs"
      @click="showDialog"
    />

    <InputGroup
      v-else
      class="source-preview"
      :class="{ 'p-disabled': isFlashing }"
    >
      <InputGroupAddon>
        <font-awesome-icon :icon="['fas', 'file']" />
      </InputGroupAddon>
      <InputText v-model="namePreview" :disabled="true" />
      <Button
        v-if="!isFlashing"
        icon="pi pi-times"
        severity="danger"
        @click="resetSource"
      />
    </InputGroup>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { storeToRefs } from 'pinia';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import { showOpenIsoDialog, showOpenAppDialog } from '@/api/dialog';
import { useFlashStore } from '@/stores';

const { selectedSource, selectedOs, isFlashing } = storeToRefs(useFlashStore());
const namePreview = ref<string | undefined>(undefined);

const showDialog = async () => {
  let res;
  if (selectedOs.value === 'windows') {
    res = await showOpenIsoDialog();
  } else if (selectedOs.value === 'macos') {
    res = await showOpenAppDialog();
  } else {
    throw Error('Unsupprted OS');
  }

  if (!res.filePaths[0]) {
    throw Error('No file selected');
  }

  selectedSource.value = res.filePaths[0];

  namePreview.value = await window.ipcInvoke(
    '/path/basename',
    selectedSource.value,
  );
};

const resetSource = () => {
  selectedSource.value = undefined;
};
</script>

<style scoped>
.source-wrapper {
  display: flex;
  justify-content: center;
}

.source-selector-header {
  display: flex;
  justify-content: space-between;
  width: 100%;
  gap: 0.75rem;
}

.source-selector-header-button {
  font-size: 0.5rem;
}

.source-preview .p-component:disabled {
  opacity: 1;
}
</style>
