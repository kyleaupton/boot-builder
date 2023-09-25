<template>
  <Sidebar />

  <div class="main">
    <Titlebar />
    <Content />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapActions } from 'pinia';

import { useDisksStore } from '@/stores/disks';
// import { useLayoutStore } from '@/stores/layout';

import Sidebar from '@/components/sidebar/Sidebar.vue';
import Titlebar from '@/components/titlebar/Titlebar.vue';
import Content from '@/components/Content.vue';

export default defineComponent({
  name: 'App',

  components: {
    Titlebar,
    Sidebar,
    Content,
  },

  async created() {
    this.registerUsbEvents();
    await this.getDisks();

    // const disksStore = useDisksStore();
    // const layoutStore = useLayoutStore();

    // if (disksStore.items && disksStore.items[0])
    // layoutStore.chosenDrive = disksStore.items[0].DeviceIdentifier;
  },

  methods: {
    ...mapActions(useDisksStore, ['getDisks', 'registerUsbEvents']),
  },
});
</script>

<style>
:root {
  /* Titlebar */
  --titlebar-height: 52px;
  --titlebar-color: #403734;

  /* Sidebar */
  --sidebar-width: 300px;

  /* Content */
  --content-color: #2f211d;
  --text-1: rgba(255, 255, 255, 0.87);

  font-family: -apple-system, BlinkMacSystemFont, sans-serif;

  color-scheme: light dark;
  color: var(--text-1);

  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
}

body {
  margin: 0;
  display: flex;
  place-items: center;
  min-width: 320px;
  min-height: 100vh;
}

h1 {
  font-size: 3.2em;
  line-height: 1.1;
}

button {
  border-radius: 4px;
  border: 1px solid transparent;
  font-size: 1em;
  background-color: #665c56;
  cursor: pointer;
  padding: 4px 12px;
}

button:active {
  background-color: #938c88;
}

code {
  background-color: #1a1a1a;
  padding: 2px 4px;
  margin: 0 4px;
  border-radius: 4px;
}

.card {
  padding: 2em;
}

.main {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  max-height: 100dvh;
}

#app {
  width: 100dvw;
  height: 100dvh;
  max-width: 100dvw;
  max-height: 100dvh;
  display: flex;
}

.drag {
  user-select: none;
  -webkit-user-select: none;
  -webkit-app-region: drag;
}

.no-drag {
  -webkit-app-region: no-drag;
}

.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}

.logo.electron:hover {
  filter: drop-shadow(0 0 2em #9feaf9);
}

.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}

.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
