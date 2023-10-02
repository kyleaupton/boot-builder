<template>
  <div class="flashing">
    <div>Flashing {{ drive.meta._name }} with {{ drive.isoFile?.name }}</div>

    <div class="flashing-progress">
      <LoadingBar />

      <div class="flashing-progress-extra">{{ statusText }}</div>
    </div>

    <div class="flashing-details">
      <div
        class="flashing-details-header"
        @click="showingDetails = !showingDetails"
      >
        <font-awesome-icon
          class="flashing-details-header-icon"
          :icon="['fas', 'chevron-up']"
          :rotation="chevronRotation"
        />
        <div class="flashing-details-header-title">Show Details</div>
      </div>

      <div v-if="showingDetails" class="flashing-details-text-wrap">
        <textarea v-model="stdout" class="flashing-details-text" disabled />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import Drive from '@/api/Drive';
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
  },

  data() {
    return {
      showingDetails: false,
    };
  },

  computed: {
    stdout: {
      get() {
        return this.drive.flashingProgress.stdout;
      },
      set() {},
    },

    statusText() {
      return this.drive.flashingProgress.currentActivity;
    },

    chevronRotation() {
      return this.showingDetails ? 180 : 90;
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
  gap: 8px;
}

.flashing-progress-extra {
  font-size: 14px;
  color: var(--text-2);
}

.flashing-details {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.flashing-details-header {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: default;
}

.flashing-details-text {
  width: 100%;
  height: calc(100vh * 0.4);
  resize: none;
}

.flashing-details-header-icon {
  font-size: 12px;
  color: var(--text-2);
}

.flashing-details-header-title {
  font-size: 12px;
  color: var(--text-2);
}
</style>
