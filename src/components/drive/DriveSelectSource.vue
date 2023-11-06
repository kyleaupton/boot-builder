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
          v-for="os of operatingSystems"
          :key="os.id"
          class="drive-os-selection-target"
          @click="selectedTarget = os.id"
        >
          <font-awesome-icon
            class="drive-os-selection-target-icon"
            :icon="os.icon"
          />

          <div class="drive-os-selection-target-title-wrap">
            <div class="drive-os-selection-target-title">{{ os.id }}</div>
          </div>
        </div>
      </div>
    </div>
  </template>

  <!-- Windows source type -->
  <template v-else-if="selectedOS && !selectedOS.sourceType">
    <div class="drive-os-selection">
      <div class="drive-os-selection-title">ISO Selection</div>
      <div class="drive-os-selection-extra">
        Are you bringing your own ISO file?
      </div>

      <div class="drive-os-selection-targets">
        <div class="drive-os-selection-target" @click="setSourceType('remote')">
          <font-awesome-icon
            class="drive-os-selection-target-icon"
            :icon="['fas', 'down-long']"
          />

          <div class="drive-os-selection-target-title-wrap">
            <div class="drive-os-selection-target-title">Download</div>
          </div>
        </div>

        <div class="drive-os-selection-target" @click="setSourceType('local')">
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
  </template>

  <!-- Source preview -->
  <!-- Final screen before flash -->
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

  <!-- Source selection -->
  <template v-else-if="selectedOS">
    <!-- Windows remote source selection -->
    <DriveSelectSourceRemote v-if="selectedOS.sourceType === 'remote'" />

    <!-- or Local selection -->
    <DriveSelectSourceLocal
      v-if="selectedOS.sourceType === 'local'"
      :suggestion-options="selectedOS.suggestionOptions"
    />
  </template>
</template>

<script lang="ts">
import { PropType, defineComponent } from 'vue';
import Drive from '@/api/Drive';

import DriveSelectSourceRemote from './DriveSelectSourceRemote.vue';
import DriveSelectSourceLocal, {
  t_suggestion_options,
} from './DriveSelectSourceLocal.vue';

export default defineComponent({
  name: 'DriveSelectOS',

  components: {
    DriveSelectSourceRemote,
    DriveSelectSourceLocal,
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
      operatingSystems: [
        {
          id: 'Windows',
          icon: ['fab', 'windows'],
          sourceType: '',
          suggestionOptions: {
            title: 'ISO Selection',
            desc: "Drag n' drop, browse, or choose a suggested ISO",
            path: '$downloads$',
            regex: /.+\.iso/,
          } as t_suggestion_options,
        },
        {
          id: 'macOS',
          icon: ['fab', 'apple'],
          sourceType: 'local',
          suggestionOptions: {
            title: 'Installer Selection',
            desc: 'Choose the desired macOS installer',
            path: '/Applications',
            regex: /Install macOS .+\.app/,
          } as t_suggestion_options,
        },
        // {
        //   id: 'Linux',
        //   icon: ['fab', 'linux'],
        //   suggestionOptions: {} as t_suggestion_options,
        // },
      ],

      selectedTarget: '',

      localFile: null as { name: string; path: string; size: number } | null,
    };
  },

  computed: {
    selectedOS() {
      if (this.selectedTarget) {
        return this.operatingSystems.find((x) => x.id === this.selectedTarget);
      }

      return undefined;
    },
  },

  methods: {
    setSourceType(sourceType: string) {
      const index = this.operatingSystems.findIndex(
        (x) => x.id === this.selectedTarget,
      );

      if (index > -1) {
        this.operatingSystems[index].sourceType = sourceType;
      }
    },

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
