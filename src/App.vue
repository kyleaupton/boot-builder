<template>
  <Titlebar />

  <div class="content">
    <div v-if="!finished" class="form-container">
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
          <Button
            class="progress-cancel"
            label="Cancel"
            severity="danger"
            @click="cancel"
          />
        </div>

        <ProgressBar
          v-if="(flash?.progress.percentage ?? -1) > -1"
          :value="flash?.progress.percentage ?? 0"
          :style="{ height: '8px' }"
          :show-value="false"
        />

        <ProgressBar v-else mode="indeterminate" :style="{ height: '8px' }" />

        <div class="progress-lower">
          <div>{{ flash?.progress.activity || '' }}</div>
          <div :style="{ display: 'flex', gap: '8px' }">
            <div v-if="flash?.progress.eta != null">{{ humanReadableEta }}</div>
            <div
              v-if="
                flash?.progress.percentage != null &&
                flash.progress.percentage > -1
              "
            >
              {{ Math.floor(flash.progress.percentage) }}%
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="flash-done-container">
      <font-awesome-icon
        :style="finished.style"
        :icon="finished.icon"
        style="--fa-animation-iteration-count: 1"
        bounce
      />
      <div class="flash-done-text">{{ finished.text }}</div>

      <div v-if="flash?.error" class="error-message">
        {{
          flash.error ||
          'An unknown error occurred while flashing the drive. Please try again.'
        }}
      </div>

      <div class="flash-done-actions">
        <Button label="Flash another" @click="reset" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState, mapActions } from 'pinia';
import { nanoid } from 'nanoid';

import { useDisksStore, useFlashStore } from '@/stores';
import { showConfirmDialog } from '@/api/dialog';

import Titlebar from './components/Titlebar.vue';
import DriveSelector from './components/DriveSelector.vue';
import OsSelector from './components/OsSelector.vue';
import SourceSelector from './components/SourceSelector.vue';

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
    };
  },

  computed: {
    ...mapState(useDisksStore, ['items']),
    ...mapState(useFlashStore, ['flash']),

    flashing() {
      return !!this.flash;
    },

    flashDisabled() {
      return !this.chosenDrive || !this.chosenOs || !this.chosenSource;
    },

    chosenDriveData() {
      return this.items.find((drive) => drive.devicePath === this.chosenDrive);
    },

    humanReadableEta() {
      if (this.flash && this.flash.progress.eta > -1) {
        const h = Math.floor(this.flash.progress.eta / 3600)
          .toString()
          .padStart(2, '0');

        const m = Math.floor((this.flash.progress.eta % 3600) / 60)
          .toString()
          .padStart(2, '0');

        const s = Math.floor(this.flash.progress.eta % 60)
          .toString()
          .padStart(2, '0');

        const eta = h !== '00' ? `${h}:${m}:${s}` : `${m}:${s}`;
        return `ETA: ${eta}`;
      }

      return '';
    },

    finished() {
      if (this.flash?.done) {
        return {
          icon: ['fas', 'circle-check'],
          text: 'Flash done!',
          style: { fontSize: '58px', color: '#4bb543' },
        };
      } else if (this.flash?.canceled) {
        return {
          icon: ['fas', 'ban'],
          text: 'Flash canceled',
          style: { fontSize: '58px', color: '#ff0000' },
        };
      } else if (this.flash?.error) {
        return {
          icon: ['fas', 'triangle-exclamation'],
          text: 'Flash failed',
          style: { fontSize: '58px', color: 'var(--color-destructive)' },
        };
      }

      return undefined;
    },
  },

  watch: {
    chosenOs() {
      this.chosenSource = '';
    },
  },

  async created() {
    this.registerUsbEvents();
    this.registerFlashEvents();
    await this.getDisks();
  },

  methods: {
    ...mapActions(useDisksStore, ['getDisks', 'registerUsbEvents']),
    ...mapActions(useFlashStore, ['registerFlashEvents']),

    async startFlash() {
      if (this.chosenDriveData) {
        const accepted = await showConfirmDialog(
          `Are you sure you want to flash ${this.chosenDriveData.description}?`,
        );

        if (accepted) {
          await window.ipcInvoke('/flash/new', nanoid(), 'windows', {
            sourcePath: this.chosenSource,
            targetVolume: this.chosenDriveData.device,
          });
        }
      }
    },

    reset() {
      this.chosenDrive = '';
      this.chosenOs = '';
      this.chosenSource = '';
    },

    async cancel() {
      if (this.flash) {
        const accepted = await showConfirmDialog(
          'Are you sure you want to cancel the flashing process?',
        );

        if (accepted) {
          await window.ipcInvoke('/flash/cancel', this.flash.id);
        }
      }
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

  /* Colors */
  --color-destructive: hsl(0 84.2% 60.2%);
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
  padding: 10% 0;
  /* margin: auto 0; */
}

.flash-done-text {
  font-size: 24px;
  font-weight: 500;
}

.error-message {
  border-radius: 6px;
  text-align: center;
  background-color: var(--surface-card);
  padding: 24px 12px;
  font-family: monospace;

  overflow: auto;
}

.flash-done-actions {
  flex-shrink: 0;
}
</style>
