<template>
  <div class="source-local">
    <div class="source-local-title-container">
      <div class="source-local-title">{{ suggestionOptions.title }}</div>
      <div class="source-local-desc">{{ suggestionOptions.desc }}</div>
    </div>

    <!-- Drop -->
    <Drop>
      <div class="drive-source-suggestions">
        <div v-if="loadingSuggestions">Loading...</div>

        <template v-else>
          <div
            v-for="suggestion of suggestions"
            :key="suggestion.path"
            class="source-local-suggestion"
          >
            <img class="source-local-suggestion-image" :src="suggestion.icon" />
            <div class="source-local-suggestion-name">
              {{ suggestion.name }}
            </div>
          </div>
        </template>
      </div>
    </Drop>

    <!-- Or -->
    <div>Or</div>

    <!-- Choose button-->
    <button>{{ buttonText }}</button>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import Drop from '@/components/Drop.vue';

export type t_suggestion_options = {
  path: string;
  regex: RegExp;
  title: string;
  desc: string;
};

type t_suggestion = {
  name: string;
  path: string;
  icon: string;
};

export default defineComponent({
  name: 'DriveSelectSourceLocal',

  components: {
    Drop,
  },

  props: {
    suggestionOptions: {
      type: Object as PropType<t_suggestion_options>,
      required: true,
    },
  },

  data() {
    return {
      suggestions: [] as t_suggestion[],
      loadingSuggestions: true,
    };
  },

  computed: {
    buttonText() {
      return 'Browse';
    },
  },

  created() {
    this.getSuggestionList();
  },

  methods: {
    async getSuggestionList() {
      const payload: t_suggestion[] = [];

      let lookupPath = this.suggestionOptions.path;
      if (this.suggestionOptions.path === '$downloads$') {
        lookupPath = await window.api.ipc.invoke('/utils/app/getPath', {
          name: 'downloads',
        });
      }

      const suggestions = (
        (await window.api.ipc.invoke('/utils/fs/readdir', {
          path: lookupPath,
        })) as string[]
      ).filter((x) => this.suggestionOptions.regex.test(x));

      for (const suggestion of suggestions) {
        const fullPath = window.api.path.join(lookupPath, suggestion);
        const ext = window.api.path.extname(suggestion);

        if (ext === '.app') {
          // If the suggestion is an macOS application
          // we want to do something special to get the icon

          payload.push({
            name: suggestion.split('.app')[0], // Remove .app bit
            path: fullPath,
            // Special call to get the icon of a macOS application
            icon: await window.api.ipc.invoke('/utils/macApp/getIcon', {
              path: fullPath,
            }),
          });
        } else {
          payload.push({
            name: suggestion,
            path: fullPath,
            // Normal app.getFileIcon() call
            icon: await window.api.ipc.invoke('/utils/app/getFileIcon', {
              path: fullPath,
            }),
          });
        }
      }

      this.suggestions = payload;
      this.loadingSuggestions = false;
    },
  },
});
</script>

<style scoped>
.source-local {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.source-local-title {
  font-weight: 600;
  text-align: center;
  font-size: 18px;
  margin-bottom: 12px;
}

.source-local-desc {
  text-align: center;
  font-size: 14px;
  color: var(--text-2);
}

.drive-source-suggestions {
  min-height: 200px;
  min-width: 300px;
  background-color: var(--titlebar-color);
  padding: 12px;
  border: 1px solid black;
  border-radius: 4px;
}

.source-local-suggestion {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.source-local-suggestion-image {
  width: 32px;
}

.source-local-suggestion-name {
  font-weight: 600;
  flex-grow: 2;
  text-align: center;
}
</style>
