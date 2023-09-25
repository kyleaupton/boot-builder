<template>
  <div class="sidebar drag">
    <div class="sidebar-section-title">External</div>

    <div class="sidebar-drives">
      <SidebarDrive
        v-for="drive of items"
        :key="drive.DeviceIdentifier"
        :drive="drive"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useDisksStore } from '@/stores/disks';
import SidebarDrive from './SidebarDrive.vue';

export default defineComponent({
  name: 'Sidebar',

  components: {
    SidebarDrive,
  },

  computed: {
    ...mapState(useDisksStore, ['items']),
  },
});
</script>

<style scoped>
.sidebar {
  flex-shrink: 0;
  width: var(--sidebar-width);
  padding: var(--titlebar-height) 12px 12px 12px;
  height: 100vh;
  background: transparent;
  border-right: 1px solid black;
  display: flex;
  flex-direction: column;
  gap: 4px;
  box-sizing: border-box;
}

.sidebar-section-title {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.4);
  margin: 0 6px;
}

.sidebar-drives {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
