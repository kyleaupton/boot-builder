<template>
  <!-- If file is chosen -->
  <template v-if="file">
    <div class="drive-source-wrap drive-source-preview">
      <div class="drive-source-preview-file">
        <font-awesome-icon
          class="drive-source-chooser-icon"
          :icon="['fab', 'windows']"
        />

        <div class="drive-source-preview-path">{{ fileName }}</div>

        <font-awesome-icon
          class="drive-source-preview-close"
          :icon="['fas', 'xmark']"
          @click="resetFile"
        />
      </div>
    </div>
  </template>

  <!-- If no file is chosen -->
  <template v-else-if="!loading">
    <div class="drive-source-wrap">
      <!-- Drop -->
      <template v-if="type === 'drop'">
        <Drop @drop="handleDrop">
          <font-awesome-icon
            class="drive-source-chooser-icon"
            style="font-size: 48px"
            :icon="['fab', 'windows']"
          />

          <div class="drive-source-chooser-title">
            Drop Windows install here
          </div>
        </Drop>
      </template>

      <!-- Suggest -->
      <template v-else-if="type === 'suggestion'">
        <div class="drive-source-selection">
          <div class="drive-source-suggestions">
            <div
              v-for="suggestion of suggestions"
              :key="suggestion.path"
              class="drive-source-suggestion"
            >
              <img :src="suggestion.icon" />
              <div>{{ suggestion.path }}</div>
            </div>
          </div>
        </div>
      </template>

      <div>or</div>

      <button @click="handleOpenFile">Choose File</button>
    </div>
  </template>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import Drive from '@/api/Drive';
import { getErrorMessage } from '@/utils/error';
import { t_file } from '@/types/iso';
import Drop from '@/components/Drop.vue';

type t_suggestion = {
  path: string;
  icon: Awaited<ReturnType<typeof window.api.getFileIcon>>;
};

export default defineComponent({
  name: 'DriveSelectLocalSource',

  components: {
    Drop,
  },

  props: {
    drive: {
      type: Object as PropType<Drive>,
      required: true,
    },

    type: {
      type: String as PropType<'drop' | 'suggestion'>,
      required: true,
    },
  },

  emits: ['fileChange'],

  data() {
    return {
      file: undefined as t_file | undefined,
      error: '',
      loading: true,
      suggestions: [] as t_suggestion[],
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

  created() {
    if (this.type === 'suggestion') {
      this.getSuggestionList();
    } else {
      this.loading = false;
    }
  },

  methods: {
    async getSuggestionList() {
      const payload: t_suggestion[] = [];

      /**
       * Currently the only OS using this
       * is macOS. We may expand this in the
       * future but for now we can assume
       * we're looking for `.app` directories
       */
      const installers = (
        (await window.api.ipc.invoke('/utils/fs/readdir', {
          path: '/Applications',
        })) as string[]
      ).filter((x) => /Install macOS .+\.app/.test(x));

      for (const installer of installers) {
        payload.push({
          path: installer,
          icon: await window.api.getFileIcon(installer),
        });
      }

      this.suggestions = payload;
      this.loading = false;
    },

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
          // @ts-ignore - File types are screwed up
          if (!file.path.split('/').pop()?.endsWith('.iso')) {
            throw Error('Only .iso files are allowed');
          }

          this.file = {
            name: file.name,
            // @ts-ignore - File types are screwed up
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
      this.file = undefined;
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

.drive-source-chooser {
  gap: 16px;
  padding: 16px;
  border: 2px dashed var(--text-1);
  border-radius: 4px;
}

.drive-source-chooser-icon {
  font-size: 28px;
}

.drive-source-chooser-title {
  font-size: 18px;
  font-weight: 600;
}

.drive-source-preview-close {
  cursor: pointer;
}

/* .drive-source-preview-close:hover {
  background-color: ;
} */

.drive-source-selection {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
}

.drive-source-suggestion {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}
</style>
