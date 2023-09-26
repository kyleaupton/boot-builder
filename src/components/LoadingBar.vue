<template>
  <div class="loading-bar" :class="{ 'loading-bar-complete': complete }">
    <div class="loading-bar-inner" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'LoadingBar',

  props: {
    complete: {
      type: Boolean,
      default: false,
    },
  },
});
</script>

<style scoped>
.loading-bar {
  height: 6px;
  width: 100%;
  border-radius: 8px;
  background-color: grey;
  overflow: hidden;
}

.loading-bar-inner {
  height: 100%;
  width: 20%;
  background-color: #157ae3;
  border: 1px solid #0e559e;
  border-radius: 8px;
  /* box-shadow: 1px 1px 4px #0e559e; */
  animation-name: slide;
  animation-duration: 4s;
  animation-iteration-count: infinite;
  animation-timing-function: ease-in-out;
}

@keyframes slide {
  0% {
    margin-left: -18%;
  }

  50% {
    margin-left: 98%;
  }

  100% {
    margin-left: -18%;
  }
}
</style>
