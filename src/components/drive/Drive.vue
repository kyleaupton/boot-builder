<template>
  <div class="drive">
    <!-- Header section -->
    <div class="drive-section drive-section-header">
      <!-- Left side header -->
      <div class="drive-section-header-left">
        <font-awesome-icon
          class="drive-section-header-icon"
          :icon="['fas', 'hard-drive']"
        />

        <div>
          <div class="drive-section-header-title">
            {{ drive.meta.manufacturer }} {{ drive.meta._name }}
          </div>

          <div>{{ drive.meta.DeviceIdentifier }}</div>
        </div>
      </div>

      <!-- Right side header -->
      <div class="drive-section-header-right">
        <div class="drive-section-header-size">{{ prettySize }}</div>
      </div>
    </div>

    <!-- Center section -->
    <div class="drive-section drive-section-center">
      <!-- Choosing file -->
      <template v-if="!drive.flashing">
        <IsoSelector @file-change="handleFileChange" />

        <div v-if="file" :style="{ display: 'grid', placeItems: 'center' }">
          <button @click="handleStartFlash">Start Flash</button>
        </div>
      </template>

      <!-- Flashing -->
      <Flashing
        v-else-if="file"
        :drive="drive"
        :file="file"
        status-text="Flashing"
      />
    </div>

    <!-- Footer section -->
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import prettyBytes from 'pretty-bytes';
import Drive from '@/api/Drive';
import { t_file } from '@/types/iso';
import IsoSelector from '@/components/iso-selector/IsoSelector.vue';
import Flashing from '@/components/drive/Flashing.vue';

export default defineComponent({
  name: 'Drive',

  components: {
    IsoSelector,
    Flashing,
  },

  props: {
    drive: {
      type: Object as PropType<Drive>,
      required: true,
    },
  },

  data() {
    return {
      file: null as t_file | null,
    };
  },

  computed: {
    prettySize() {
      return prettyBytes(this.drive?.meta.Size || 0);
    },
  },

  methods: {
    handleFileChange(file: t_file | null) {
      this.file = file;
    },

    async handleStartFlash() {
      if (this.drive && this.file) {
        // await request "are you sure????"
        this.drive.startFlash();
      }
    },
  },
});
</script>

<style scoped>
.drive {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding: 24px;
  height: calc(100% - var(--titlebar-height));
}

.drive-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}

.drive-section-center {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.drive-section-header-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.drive-section-header-right {
  flex-shrink: 0;
}

.drive-section-header-icon {
  font-size: 48px;
}

.drive-section-header-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.drive-section-header-size {
  padding: 8px 16px;
  border: 1px solid var(--text-1);
  border-radius: 4px;
  text-align: center;
}
</style>
