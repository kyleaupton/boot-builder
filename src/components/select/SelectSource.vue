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
        <font-awesome-icon
          class="drive-section-header-icon"
          :icon="['fas', 'file']"
        />
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

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import { showOpenIsoDialog, showOpenAppDialog } from '@/api/dialog';
import { useFlashStore } from '@/stores';

export default defineComponent({
  name: 'SourceSelector',

  components: {
    InputGroup,
    InputGroupAddon,
  },

  data() {
    return {
      namePreview: '',
    };
  },

  computed: {
    ...mapState(useFlashStore, ['selectedSource', 'selectedOs', 'isFlashing']),
  },

  methods: {
    async showDialog() {
      let res;
      if (this.selectedOs === 'windows') {
        res = await showOpenIsoDialog();
      } else if (this.selectedOs === 'macos') {
        res = await showOpenAppDialog();
      } else {
        throw Error('Unsupprted OS');
      }

      if (!res.filePaths[0]) {
        throw Error('No file selected');
      }

      this.selectedSource = res.filePaths[0];
    },

    resetSource() {
      this.selectedSource = undefined;
    },
  },
});
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
