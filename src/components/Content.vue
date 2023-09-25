<template>
  <div class="content">
    <!-- Empty -->
    <template v-if="!drive">
      <div class="content-empty">
        <font-awesome-icon
          class="content-empty-icon"
          :icon="['fas', 'floppy-disk']"
        />

        <div class="content-empty-text">
          Insert and select a drive to get started
        </div>
      </div>
    </template>

    <!-- Content -->
    <template v-else>
      <Drive :drive="drive" />
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';

import { useDisksStore } from '@/stores/disks';
import { useLayoutStore } from '@/stores/layout';

import Drive from '@/components/drive/Drive.vue';

export default defineComponent({
  name: 'Content',

  components: {
    Drive,
  },

  computed: {
    ...mapState(useDisksStore, ['items', 'loaded']),
    ...mapState(useLayoutStore, ['chosenDrive']),

    drive() {
      return this.chosenDrive && this.loaded
        ? this.items?.find((x) => x.DeviceIdentifier === this.chosenDrive)
        : null;
    },
  },
});
</script>

<style scoped>
.content {
  background-color: var(--content-color);
  width: 100%;
  height: 100%;
}

.content-empty {
  width: 100%;
  text-align: center;
  padding: 20% 0 0 0;
}

.content-empty-icon {
  color: rgba(255, 255, 255, 0.5);
  font-size: 18px;
}

.content-empty-text {
  font-weight: 500;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
  user-select: none;
}
</style>
