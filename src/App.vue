<template>
  <Titlebar />

  <div class="form-container">
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

    <ProgressBar v-else :value="progress" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapActions } from 'pinia';

import { useDisksStore } from '@/stores/disks';

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
      flashing: false,
      progress: 0,
    };
  },

  computed: {
    flashDisabled() {
      return !this.chosenDrive || !this.chosenOs || !this.chosenSource;
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
      this.flashing = true;

      const id = setInterval(() => (this.progress += 3), 1000);
      await new Promise((resolve) => setTimeout(resolve, 30000));
      window.clearInterval(id);

      this.flashing = false;
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
</style>
