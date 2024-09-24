<template>
  <Titlebar />
  <Main />
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useDisksStore, useFlashStore } from '@/stores';
import Titlebar from '@/components/Titlebar.vue';
import Main from '@/components/Main.vue';

const disksStore = useDisksStore();
const flashStore = useFlashStore();

onMounted(() => {
  disksStore.getDisks();
  disksStore.registerUsbEvents();
  flashStore.registerFlashEvents();
});
</script>
