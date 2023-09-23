<template>
  <div class="titlebar drag"></div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Titlebar',
});
</script>

<style scoped>
.titlebar {
  height: var(--titlebar-height);
  width: 100%;
  background-color: var(--titlebar-color);
  border-bottom: 1px solid black;
}
</style>
