<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Events } from '@wailsio/runtime';
import { StartJob } from '../bindings/boot-builder/internal/service/jobsservice';
import { ListDrives} from '../bindings/boot-builder/internal/service/drivesservice'

import '@/styles.css'

const currentProgress = ref(null)

const prettyProgress = computed(() => {
  if (currentProgress.value === null) {
    return 'No progress';
  }

  return JSON.stringify(currentProgress.value, null, 2);
})

const startTest = async () => {
  const job = await StartJob({
    InstallerID: 'linux.ubuntu',
    SourceLocal: '/Users/kyleupton/Downloads/ubuntu-24.04.3-live-server-amd64.iso',
    DriveID: '/dev/disk4',
  })

  console.log('job', job)
}

onMounted(async () => {
  const drives = await ListDrives()
  console.log('drives', drives)

  Events.On('job:event', (event) => {
    console.log('event', event)
    currentProgress.value = event.data
  })
})
</script>

<template>
  <div class="container">
    <button @click="startTest">Start Test</button>
    <pre>{{ prettyProgress }}</pre>
  </div>
</template>

<style scoped>
.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  width: 100vw;
}

pre {
  width: 80%;
  height: 80%;
  overflow-y: scroll;
  background-color: #f0f0f0;
  padding: 10px;
  border-radius: 10px;
  font-size: 14px;
  font-family: monospace;
}
</style>
