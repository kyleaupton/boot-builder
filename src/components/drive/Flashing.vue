<template>
  <div class="flashing">
    <div>Flashing {{ drive.meta._name }} with {{ file.name }}</div>

    <div class="flashing-progress">
      <LoadingBar />

      <div>{{ statusText }}</div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import Drive from '@/api/Drive';
import { t_file } from '@/types/iso';
import LoadingBar from '@/components/LoadingBar.vue';

export default defineComponent({
  name: 'Flashing',

  components: {
    LoadingBar,
  },

  props: {
    drive: {
      type: Object as PropType<Drive>,
      required: true,
    },

    file: {
      type: Object as PropType<t_file>,
      required: true,
    },

    statusText: {
      type: String,
      default: '',
    },
  },
});
</script>

<style scoped>
.flashing {
  display: flex;
  flex-direction: column;
  width: 100%;
  justify-content: center;
  align-items: center;
  gap: 2rem;
}

.flashing-progress {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
