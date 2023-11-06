<template>
  <Drop @drop="handleDrop">
    <div class="drive-drop">
      <font-awesome-icon
        class="drive-source-chooser-icon"
        style="font-size: 48px"
        :icon="['fab', 'windows']"
      />

      <div class="drive-source-chooser-title">Drop Windows install here</div>
    </div>
  </Drop>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { getErrorMessage } from '@/utils/error';
import Drop from '@/components/Drop.vue';

export default defineComponent({
  name: 'DriveSelectSourceDrop',

  components: {
    Drop,
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
  },
});
</script>

<style scoped>
.drive-drop {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2em;
  width: 80%;
}
</style>
