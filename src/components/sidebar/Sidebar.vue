<template>
  <div class="sidebar drag">
    <div class="sidebar-section-title">External</div>

    <div class="sidebar-drives">
      <div
        v-for="drive in items"
        :key="drive.DeviceIdentifier"
        class="sidebar-drive"
      >
        <font-awesome-icon :icon="['fab', 'usb']" />
        <div class="sidebar-drive-name">
          {{ drive.manufacturer }} {{ drive._name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useDisksStore } from '@/stores/disks';

export default defineComponent({
  name: 'Sidebar',

  computed: {
    ...mapState(useDisksStore, ['items']),
  },
});
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  padding: var(--titlebar-height) 16px 16px 16px;
  height: 100vh;
  background: transparent;
  border-right: 1px solid black;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-section-title {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.4);
}

.sidebar-drive {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sidebar-drive-name {
  font-size: 14px;
  font-weight: 500;
}
</style>
