<template>
  <div class="drive-os-selection">
    <div class="drive-os-selection-title">OS Selection</div>
    <div class="drive-os-selection-extra">
      What are you making installation media for?
    </div>

    <div class="drive-os-selection-targets">
      <div
        v-for="target of mediaTargets"
        :key="target.id"
        class="drive-os-selection-target"
        @click="$emit('selectedOs', target.id)"
      >
        <font-awesome-icon
          class="drive-os-selection-target-icon"
          :icon="target.icon"
        />

        <div style="width: 100%">
          <div class="drive-os-selection-target-title">{{ target.id }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'DriveSelectOS',

  emits: ['selectedOs'],

  data() {
    return {
      mediaTargets: [
        {
          id: 'Windows',
          icon: ['fab', 'windows'],
        },
        {
          id: 'macOS',
          icon: ['fab', 'apple'],
        },
        // {
        //   id: 'Linux',
        //   icon: ['fab', 'linux'],
        // },
      ],
    };
  },
});
</script>

<style scoped>
.drive-os-selection {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.drive-os-selection-title {
  font-weight: 600;
  text-align: center;
  margin-bottom: 12px;
  font-size: 18px;
}

.drive-os-selection-extra {
  text-align: center;
  font-size: 14px;
  color: var(--text-2);
  margin-bottom: 42px;
}

.drive-os-selection-targets {
  align-self: center;
  display: flex;
  gap: 12px;
}

.drive-os-selection-target {
  width: 100px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  align-items: center;
  cursor: pointer;
  border-radius: 8px;
  padding: 16px;
  transition: background 0.1s ease;
}

.drive-os-selection-target:hover {
  background: var(--hover-1);
}

.drive-os-selection-target-icon {
  height: 42px;
}

.drive-os-selection-target-title {
  text-align: center;
  font-weight: 600;
}
</style>
