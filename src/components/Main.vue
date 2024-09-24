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
import { mapState } from 'pinia';
import { nanoid } from 'nanoid';
import { useDisksStore, useFlashStore } from '@/stores';
import { showConfirmDialog } from '@/api/dialog';
import SelectDrive from '@/components/select/SelectDrive.vue';
import SelectOs from '@/components/select/SelectOs.vue';
import SelectSource from '@/components/select/SelectSource.vue';
import FlashProgress from '@/components/flash/FlashProgress.vue';
import FlashFinished from '@/components/flash/FlashFinished.vue';

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
      if (!this.selectedDrive || !this.selectedOs || !this.selectedSource) {
        return;
      }

      const driveData = this.items.find(
        (drive) => drive.devicePath === this.selectedDrive,
      );

      if (!driveData) {
        return;
      }

      const accepted = await showConfirmDialog(
        `Are you sure you want to flash ${driveData.description}?`,
      );

      if (accepted) {
        // TODO: Support macOS
        await window.ipcInvoke('/flash/new', nanoid(), 'windows', {
          sourcePath: this.selectedSource,
          targetVolume: this.selectedDrive,
        });
      }
    },
  },
});
</script>

<style scoped></style>
