<template>
  <div class="titlebar drag">
    <img height="50" src="../assets/logo.svg" />
    <p class="titlebar-text">Boot Builder</p>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Titlebar',
});
</script>

<style scoped>
.titlebar {
  width: 100%;
  box-sizing: border-box;
  padding: 2rem 2rem 0 2rem;

  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
}

.titlebar-text {
  font-weight: 600;
}

.drag {
  user-select: none;
  -webkit-user-select: none;
  -webkit-app-region: drag;
}

/* Not currently used. */
/* But this is how one would */
/* allow things to be clickable */
.no-drag {
  -webkit-app-region: no-drag;
}
</style>
