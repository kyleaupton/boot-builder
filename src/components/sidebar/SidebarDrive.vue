<template>
  <div
    class="sidebar-drive no-drag"
    :class="[{ 'sidebar-drive-selected': selected }]"
    @click="handleClick"
  >
    <font-awesome-icon class="sidebar-drive-icon" :icon="['fab', 'usb']" />

    <div class="sidebar-drive-name">
      {{ drive.meta.manufacturer }} {{ drive.meta._name }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { mapState } from 'pinia';
import Drive from '@/api/Drive';
import { useLayoutStore } from '@/stores/layout';

export default defineComponent({
  name: 'SidebarDrive',

  props: {
    drive: {
      type: Object as PropType<Drive>,
      required: true,
    },
  },

  computed: {
    ...mapState(useLayoutStore, ['chosenDrive']),

    selected() {
      return this.chosenDrive === this.drive.meta.DeviceIdentifier;
    },
  },

  methods: {
    handleClick() {
      const store = useLayoutStore();
      store.chosenDrive = this.drive.meta.DeviceIdentifier;
    },
  },
});
</script>

<style scoped>
.sidebar-drive {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px;
  border-radius: 4px;
}

.sidebar-drive-selected {
  background-color: #175dc7;
}

.sidebar-drive-name {
  font-size: 12px;
  font-weight: 550;
}

.sidebar-drive-icon {
  font-size: 12px;
}
</style>
