<template>
  <Titlebar />

  <div v-if="!progress?.done" class="form-container">
    <DriveSelector v-model="chosenDrive" :flashing="flashing" />
    <OsSelector v-model="chosenOs" :flashing="flashing" />
    <SourceSelector
      v-model="chosenSource"
      :chosen-os="chosenOs"
      :flashing="flashing"
    />

    <Button
      v-if="!flashing"
      label="Flash"
      :disabled="flashDisabled"
      @click="startFlash"
    />

    <div v-else class="progress-container">
      <div class="progress-upper">
        <div class="progress-flashing">Flashing...</div>
        <Button class="progress-cancel" label="Cancel" severity="danger" />
      </div>

      <ProgressBar
        v-if="(progress?.percentage ?? -1) > -1"
        :value="progress?.percentage ?? 0"
        :style="{ height: '8px' }"
        :show-value="false"
      />

      <ProgressBar v-else mode="indeterminate" :style="{ height: '8px' }" />

      <div class="progress-lower">
        <div>{{ progress?.activity || '' }}</div>
        <div :style="{ display: 'flex', gap: '8px' }">
          <div v-if="progress?.eta != null">{{ humanReadableEta }}</div>
          <div v-if="progress?.percentage != null && progress.percentage > -1">
            {{ Math.floor(progress.percentage) }}%
          </div>
        </div>
      </div>
    </div>
  </div>

  <div v-else class="flash-done-container">
    <font-awesome-icon
      class="flash-done-icon"
      :icon="['fas', 'circle-check']"
      style="--fa-animation-iteration-count: 1"
      bounce
    />

    <div class="flash-done-text">Flash done!</div>

    <Button label="Flash another" @click="reset" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState, mapActions } from 'pinia';
import { nanoid } from 'nanoid';

import { useDisksStore } from './stores/disks';

import Titlebar from './components/Titlebar.vue';
import DriveSelector from './components/DriveSelector.vue';
import OsSelector from './components/OsSelector.vue';
import SourceSelector from './components/SourceSelector.vue';

interface IProgress {
  id: string;
  activity: string;
  done: boolean;
  // ETA
  transferred: number;
  speed: number;
  percentage: number;
  eta: number;
}

export default defineComponent({
  name: 'App',

  components: {
    Titlebar,
    DriveSelector,
    OsSelector,
    SourceSelector,
  },

  data() {
    return {
      chosenDrive: '',
      chosenOs: '',
      chosenSource: '',
      flashing: false,
      progress: undefined as IProgress | undefined,
    };
  },

  computed: {
    ...mapState(useDisksStore, ['items']),

    flashDisabled() {
      return !this.chosenDrive || !this.chosenOs || !this.chosenSource;
    },

    chosenDriveData() {
      return this.items.find((drive) => drive.devicePath === this.chosenDrive);
    },

    humanReadableEta() {
      if (this.progress && this.progress.eta > -1) {
        const h = Math.floor(this.progress.eta / 3600)
          .toString()
          .padStart(2, '0');

        const m = Math.floor((this.progress.eta % 3600) / 60)
          .toString()
          .padStart(2, '0');

        const s = Math.floor(this.progress.eta % 60)
          .toString()
          .padStart(2, '0');

        const eta = h !== '00' ? `${h}:${m}:${s}` : `${m}:${s}`;
        return `ETA: ${eta}`;
      }

      return '';
    },
  },

  watch: {
    chosenOs() {
      this.chosenSource = '';
    },
  },

  async created() {
    this.registerUsbEvents();
    await this.getDisks();
  },

  methods: {
    ...mapActions(useDisksStore, ['getDisks', 'registerUsbEvents']),

    async startFlash() {
      if (this.chosenDriveData) {
        const res = await window.api.showConfirmDialog(
          `Are you sure you want to flash ${this.chosenDriveData.description}?`,
        );

        // response == 0 means "Yes"
        // response == 1 means "No"

        if (res.response === 0) {
          this.flashing = true;

          const id = nanoid();
          this.registerFlashEvents(id);

          // Start flashing
          let url: string;
          switch (this.chosenOs) {
            case 'windows':
              url = '/flash/windows';
              break;
            case 'macos':
              url = '/flash/macOS';
              break;
            default:
              throw new Error('Invalid OS');
          }

          await window.api.ipc.invoke(url, {
            sourcePath: this.chosenSource,
            targetVolume: this.chosenDriveData.device,
            id,
          });
        }
      }
    },

    registerFlashEvents(id: string) {
      window.api.ipc.recieve(`flash-${id}-progress`, this.handleProgress);
    },

    removeFlashEvents(id: string) {
      window.api.ipc.removeListener(
        `flash-${id}-progress`,
        this.handleProgress,
      );
    },

    handleProgress(progress: IProgress) {
      this.progress = progress;

      if (progress.done) {
        this.flashing = false;
        this.removeFlashEvents(progress.id);
      }
    },

    reset() {
      this.chosenDrive = '';
      this.chosenOs = '';
      this.chosenSource = '';
      this.flashing = false;
      this.progress = undefined;
    },
  },
});
</script>

<style>
:root {
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
}

#app {
  width: 100dvw;
  height: 100dvh;
  max-width: 100dvw;
  max-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  background-color: var(--surface-ground);
}

body {
  margin: 0;
  display: flex;
  place-items: center;
  min-width: 320px;
  min-height: 100vh;
}

.form-container {
  flex-grow: 2;

  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 36px;
  width: 50%;
}

.drag {
  user-select: none;
  -webkit-user-select: none;
  -webkit-app-region: drag;
}

.no-drag {
  -webkit-app-region: no-drag;
}

.progress-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-upper {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.progress-cancel {
  font-size: 12px;
  padding: 0.25rem 0.5rem;
}

.progress-lower {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.progress-flashing {
  font-size: 18px;
  font-weight: 500;
}

.flash-done-container {
  flex-grow: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  width: 100%;
  height: 100%;
  padding: 15% 0;
  /* margin: auto 0; */
}

.flash-done-icon {
  font-size: 58px;
  color: #4bb543;
}

.flash-done-text {
  font-size: 24px;
  font-weight: 500;
}
</style>
