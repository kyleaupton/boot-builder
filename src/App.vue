<template>
  <Titlebar />
  <Main />
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapActions } from 'pinia';
import { useDisksStore, useFlashStore } from '@/stores';
import Titlebar from '@/components/Titlebar.vue';
import Main from '@/components/Main.vue';

export default defineComponent({
  name: 'App',

  components: {
    Titlebar,
    Main,
  },

  async created() {
    this.registerUsbEvents();
    this.registerFlashEvents();
    await this.getDisks();
  },

  methods: {
    ...mapActions(useDisksStore, ['getDisks', 'registerUsbEvents']),
    ...mapActions(useFlashStore, ['registerFlashEvents']),
  },
});
</script>

<style>
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
