<template>
  <div class="content">
    <div v-if="!isFinished" class="form-container">
      <SelectDrive />
      <SelectOs />
      <SelectSource />

      <Button
        v-if="!isFlashing"
        label="Flash"
        :disabled="invalidData"
        @click="startFlash"
      />

      <FlashProgress v-else />
    </div>

    <FlashFinished v-else />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import log from 'electron-log/renderer';
import { mapState } from 'pinia';
import { nanoid } from 'nanoid';
import { useDisksStore, useFlashStore } from '@/stores';
import { showConfirmDialog } from '@/api/dialog';
import SelectDrive from '@/components/select/SelectDrive.vue';
import SelectOs from '@/components/select/SelectOs.vue';
import SelectSource from '@/components/select/SelectSource.vue';
import FlashProgress from '@/components/flash/FlashProgress.vue';
import FlashFinished from '@/components/flash/FlashFinished.vue';

const logger = log.scope('renderer/Main');

export default defineComponent({
  name: 'Main',

  components: {
    SelectDrive,
    SelectOs,
    SelectSource,
    FlashProgress,
    FlashFinished,
  },

  computed: {
    ...mapState(useDisksStore, ['items']),
    ...mapState(useFlashStore, [
      'selectedDrive',
      'selectedOs',
      'selectedSource',
      'isFlashing',
      'isFinished',
    ]),

    invalidData() {
      return !this.selectedDrive || !this.selectedOs || !this.selectedSource;
    },
  },

  methods: {
    async startFlash() {
      logger.info('Starting flash process');

      if (!this.selectedDrive || !this.selectedOs || !this.selectedSource) {
        logger.info('Invalid data');
        return;
      }

      const driveData = this.items.find(
        (drive) => drive.device === this.selectedDrive,
      );

      if (!driveData) {
        logger.error('Drive not found');
        return;
      }

      const accepted = await showConfirmDialog(
        `Are you sure you want to flash ${driveData.description}?`,
      );

      if (!accepted) {
        logger.info('User canceled flash');
        return;
      }

      // TODO: Support macOS
      await window.ipcInvoke('/flash/new', nanoid(), 'windows', {
        sourcePath: this.selectedSource,
        targetVolume: this.selectedDrive,
      });
    },
  },
});
</script>

<style scoped>
.content {
  padding: 0 4em;
  overflow: hidden;
  height: 100%;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.form-container {
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 36px;
}
</style>
