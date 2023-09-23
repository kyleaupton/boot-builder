<script setup lang="ts">
import Sidebar from '@/components/sidebar/Sidebar.vue';
import Titlebar from '@/components/titlebar/Titlebar.vue';
</script>

<template>
  <Sidebar />

  <div class="main">
    <Titlebar />

    <div class="content">
      <div>test</div>
    </div>
  </div>
</template>

<style>
:root {
  /* Titlebar */
  --titlebar-height: 52px;
  --titlebar-color: #403734;

  /* Sidebar */
  --sidebar-width: 350px;

  /* Content */
  --content-color: #2f211d;

  font-family: Inter, system-ui, Avenir, Helvetica, Arial, sans-serif;
  line-height: 1.5;
  font-weight: 400;

  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);

  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
}

a {
  font-weight: 500;
  color: #646cff;
  text-decoration: inherit;
}

a:hover {
  color: #535bf2;
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
  border-radius: 8px;
  border: 1px solid transparent;
  padding: 0.6em 1.2em;
  font-size: 1em;
  font-weight: 500;
  font-family: inherit;
  background-color: #1a1a1a;
  cursor: pointer;
  transition: border-color 0.25s;
}

button:hover {
  border-color: #646cff;
}

button:focus,
button:focus-visible {
  outline: 4px auto -webkit-focus-ring-color;
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
  width: 100%;
  height: 100vh;
}

.content {
  background-color: var(--content-color);
  width: 100%;
  height: 100%;
}

#app {
  width: 100vw;
  display: flex;
}

.drag {
  user-select: none;
  -webkit-user-select: none;
  -webkit-app-region: drag;
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
