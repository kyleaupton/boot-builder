<template>
  <div class="iso-chooser">
    <font-awesome-icon
      class="iso-chooser-icon"
      :icon="['fas', 'compact-disc']"
    />

    <div class="iso-chooser-title">Drag and drop .iso here</div>

    <div>or</div>

    <button>Browse...</button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'IsoSelector',
});
</script>

<style scoped>
.iso-chooser {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
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
</style>
