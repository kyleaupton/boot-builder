<template>
  <div class="flash-done-container">
    <font-awesome-icon
      :style="finished?.style"
      :icon="finished?.icon"
      style="--fa-animation-iteration-count: 1"
      bounce
    />
    <div class="flash-done-text">{{ finished?.text }}</div>

    <div v-if="flash?.error" class="error-message">
      {{
        flash.error ||
        'An unknown error occurred while flashing the drive. Please try again.'
      }}
    </div>

    <div class="flash-done-actions">
      <Button label="Flash another" @click="reset" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState, mapActions } from 'pinia';
import { useFlashStore } from '@/stores';

export default defineComponent({
  name: 'FlashFinished',

  computed: {
    ...mapState(useFlashStore, ['flash']),

    finished() {
      if (this.flash?.done) {
        return {
          icon: ['fas', 'circle-check'],
          text: 'Flash done!',
          style: { fontSize: '58px', color: '#4bb543' },
        };
      } else if (this.flash?.canceled) {
        return {
          icon: ['fas', 'ban'],
          text: 'Flash canceled',
          style: { fontSize: '58px', color: '#ff0000' },
        };
      } else if (this.flash?.error) {
        return {
          icon: ['fas', 'triangle-exclamation'],
          text: 'Flash failed',
          style: { fontSize: '58px', color: 'var(--color-destructive)' },
        };
      }

      return undefined;
    },
  },

  methods: {
    ...mapActions(useFlashStore, ['reset']),
  },
});
</script>

<style scoped></style>
