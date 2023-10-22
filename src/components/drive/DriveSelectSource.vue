<template>
  <!-- Select OS -->
  <template v-if="!selectedTarget">
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
          @click="selectedTarget = target.id"
        >
          <font-awesome-icon
            class="drive-os-selection-target-icon"
            :icon="target.icon"
          />

          <div class="drive-os-selection-target-title-wrap">
            <div class="drive-os-selection-target-title">{{ target.id }}</div>
          </div>
        </div>
      </div>
    </div>
  </template>

  <!-- Selected local file preview -->
  <template v-else-if="localFile">
    <div class="drive-source-wrap drive-source-preview">
      <div class="drive-source-preview-file">
        <font-awesome-icon
          class="drive-source-chooser-icon"
          :icon="['fab', 'windows']"
        />

        <div class="drive-source-preview-path">{{ localFile.name }}</div>

        <font-awesome-icon
          class="drive-source-preview-close"
          :icon="['fas', 'xmark']"
          @click="localFile = null"
        />
      </div>

      <button>Start Flash</button>
    </div>
  </template>

  <!-- Windows -->
  <template v-else-if="selectedTarget === 'Windows'">
    <div v-if="!windowsState.download" class="drive-os-selection">
      <div class="drive-os-selection-title">ISO Selection</div>
      <div class="drive-os-selection-extra">
        Are you bringing your own ISO file?
      </div>

      <div class="drive-os-selection-targets">
        <div
          class="drive-os-selection-target drive-os-selection-target-download"
        >
          <font-awesome-icon
            class="drive-os-selection-target-icon"
            :icon="['fas', 'down-long']"
          />

          <div class="drive-os-selection-target-title-wrap">
            <div class="drive-os-selection-target-title">Download</div>
            <div
              style="
                font-size: 12px;
                text-align: center;
                color: var(--text-2);
                margin-top: 4px;
              "
            >
              Coming soon
            </div>
          </div>
        </div>

        <div class="drive-os-selection-target" @click="handleSelectIso">
          <font-awesome-icon
            class="drive-os-selection-target-icon"
            :icon="['fas', 'desktop']"
          />

          <div class="drive-os-selection-target-title-wrap">
            <div class="drive-os-selection-target-title">Local File</div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="">got here</div>
  </template>

  <!-- macOS -->
  <template v-else-if="selectedTarget === 'macOS'">
    <DriveSelectLocalSource :drive="drive" type="suggestion" />
  </template>
</template>

<script lang="ts">
import { PropType, defineComponent } from 'vue';
import Drive from '@/api/Drive';
import DriveSelectLocalSource from './DriveSelectLocalSource.vue';

export default defineComponent({
  name: 'DriveSelectOS',

  components: {
    DriveSelectLocalSource,
  },

  props: {
    drive: {
      type: Object as PropType<Drive>,
      required: true,
    },
  },

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

      selectedTarget: '',

      localFile: null as { name: string; path: string; size: number } | null,

      windowsState: {
        download: false,
      },
    };
  },

  methods: {
    async handleSelectIso() {
      const path = (await window.api.showOpenIsoDialog()).filePaths[0];
      this.localFile = await window.api.getFileFromPath(path);
      console.log(this.localFile);
    },
  },
});
</script>

<style scoped>
.drive-source-wrap {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 16px;
}

.drive-source-preview-file {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  text-align: center;
  margin-top: 32px;
}

.drive-source-preview-path {
  font-weight: 600;
}

.drive-source-preview-close {
  cursor: pointer;
}

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

.drive-os-selection-target:not(.drive-os-selection-target-download):hover {
  background: var(--hover-1);
}

.drive-os-selection-target-icon {
  height: 42px;
}

.drive-os-selection-target-title-wrap {
  flex-grow: 2;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.drive-os-selection-target-title {
  text-align: center;
  font-weight: 600;
}

.drive-os-selection-target-download {
  cursor: not-allowed;
}
</style>
