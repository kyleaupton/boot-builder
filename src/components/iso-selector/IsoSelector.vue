<template>
  <!-- If file is chosen -->
  <template v-if="file">
    <div class="iso-wrap iso-preview">
      <div class="iso-preview-file">
        <font-awesome-icon
          class="iso-chooser-icon"
          :icon="['fab', 'windows']"
        />
        <div class="iso-preview-path">{{ fileName }}</div>
        <font-awesome-icon
          class="iso-preview-close"
          :icon="['fas', 'xmark']"
          @click="resetFile"
        />
      </div>
    </div>
  </template>

  <!-- If no file is chosen -->
  <template v-else>
    <Drop @drop="handleDrop">
      <div class="iso-wrap iso-chooser">
        <font-awesome-icon
          class="iso-chooser-icon"
          style="font-size: 48px"
          :icon="['fab', 'windows']"
        />
        <div class="iso-chooser-title">Drop Windows install here</div>
        <div>or</div>
        <button @click="handleOpenFile">Choose File</button>
      </div>
    </Drop>
  </template>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { getErrorMessage } from '@/utils/error';
import { t_file } from '@/types/iso';
import Drop from '@/components/Drop.vue';

export default defineComponent({
  name: 'IsoSelector',

  components: {
    Drop,
  },

  emits: ['fileChange'],

  data() {
    return {
      file: null as t_file | null,
      error: '',
    };
  },

  computed: {
    fileName() {
      return this.file ? this.file.path.split('/').pop() : '';
    },
  },

  watch: {
    file() {
      this.$emit('fileChange', this.file);
    },
  },

  methods: {
    handleDrop(e: DragEvent) {
      if (e.dataTransfer) {
        try {
          const { files } = e.dataTransfer;

          /**
           * EDGE CASE: More than one file
           */
          if (files.length > 1) {
            throw Error('Only one file is allowed');
          }

          const file = files[0];

          /**
           * EDGE CASE: Something that is not a .iso file
           */
          if (!file.path.split('/').pop()?.endsWith('.iso')) {
            throw Error('Only .iso files are allowed');
          }

          this.file = {
            name: file.name,
            path: file.path,
            size: file.size,
          };
        } catch (e) {
          console.error(e);
          this.error = getErrorMessage(e);
        }
      }
    },

    async handleOpenFile() {
      const path = (await window.api.showOpenIsoDialog()).filePaths[0];
      this.file = await window.api.getFileFromPath(path);
    },

    resetFile() {
      this.file = null;
    },
  },
});
</script>

<style scoped>
.iso-wrap {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 16px;
}

.iso-preview-file {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  text-align: center;
  margin-top: 32px;
}

.iso-preview-path {
  font-weight: 600;
}

.iso-chooser {
  gap: 16px;
  padding: 16px;
  border: 2px dashed var(--text-1);
  border-radius: 4px;
}

.iso-chooser-icon {
  font-size: 28px;
}

.iso-chooser-title {
  font-size: 18px;
  font-weight: 600;
}

.iso-preview-close {
  cursor: pointer;
}

/* .iso-preview-close:hover {
  background-color: ;
} */
</style>
