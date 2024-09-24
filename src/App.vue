<template>
  <Titlebar />
  <Main />
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useDisksStore, useFlashStore } from '@/stores';
import Titlebar from '@/components/Titlebar.vue';
import Main from '@/components/Main.vue';

const disksStore = useDisksStore();
const flashStore = useFlashStore();
const { isFinished } = storeToRefs(flashStore);

onMounted(() => {
  disksStore.getDisks();
  disksStore.registerUsbEvents();
  flashStore.registerFlashEvents();
});

watch(isFinished, () => {
  disksStore.getDisks();
});
</script>
