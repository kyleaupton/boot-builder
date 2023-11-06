<template>
  <div class="drop">
    <div
      class="drop-cover"
      :class="{ 'drop-cover-dragging': isDragging }"
    ></div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Drop',

  props: {
    title: {
      type: String,
      default: 'Drop Files to Upload',
    },
    subtitle: {
      type: String,
      default: '',
    },
  },

  emits: ['drop'],

  data() {
    return {
      isDragging: false,
    };
  },

  mounted() {
    this.$el.addEventListener('dragenter', this.drag);
    this.$el.addEventListener('dragover', this.drag);
    this.$el.addEventListener('dragleave', this.dragleave);
    this.$el.addEventListener('drop', this.drop);
  },

  beforeUnmount() {
    this.$el.removeEventListener('dragenter', this.drag);
    this.$el.removeEventListener('dragover', this.drag);
    this.$el.removeEventListener('dragleave', this.dragleave);
    this.$el.removeEventListener('drop', this.drop);
  },

  methods: {
    drag(e: DragEvent) {
      e.stopPropagation();
      e.preventDefault();

      this.isDragging = true;
    },

    dragleave() {
      this.isDragging = false;
    },

    drop(e: DragEvent) {
      e.stopPropagation();
      e.preventDefault();

      this.$emit('drop', e);

      this.isDragging = false;
    },
  },
});
</script>

<style scoped>
.drop {
  position: relative;
  width: 100%;
  display: grid;
  place-content: center;
}

.drop-cover {
  /* */
  position: absolute;
  /* */

  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  background: rgba(255, 255, 255, 0.4);
  opacity: 0;
  gap: 8px;
  pointer-events: none;
  border-radius: 4px;
  /* backdrop-filter: blur(3px); */
}

.drop-cover-dragging {
  opacity: 1;
}
</style>
