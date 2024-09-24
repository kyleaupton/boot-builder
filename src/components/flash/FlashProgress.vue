<template>
  <div class="progress-container">
    <div class="progress-upper">
      <div class="progress-flashing">Flashing...</div>
      <Button
        class="progress-cancel"
        label="Cancel"
        severity="danger"
        @click="cancelFlash"
      />
    </div>

    <ProgressBar
      v-if="(flash?.progress.percentage ?? -1) > -1"
      :value="flash?.progress.percentage ?? 0"
      :style="{ height: '8px' }"
      :show-value="false"
    />

    <ProgressBar v-else mode="indeterminate" :style="{ height: '8px' }" />

    <div class="progress-lower">
      <div>{{ flash?.progress.activity || '' }}</div>
      <div :style="{ display: 'flex', gap: '8px' }">
        <div v-if="flash?.progress.eta != null">{{ humanReadableEta }}</div>
        <div
          v-if="
            flash?.progress.percentage != null && flash.progress.percentage > -1
          "
        >
          {{ Math.floor(flash.progress.percentage) }}%
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState, mapActions } from 'pinia';
import { useFlashStore } from '@/stores';

export default defineComponent({
  name: 'FlashProgress',

  computed: {
    ...mapState(useFlashStore, ['flash']),

    humanReadableEta() {
      if (this.flash && this.flash.progress.eta > -1) {
        const h = Math.floor(this.flash.progress.eta / 3600)
          .toString()
          .padStart(2, '0');

        const m = Math.floor((this.flash.progress.eta % 3600) / 60)
          .toString()
          .padStart(2, '0');

        const s = Math.floor(this.flash.progress.eta % 60)
          .toString()
          .padStart(2, '0');

        const eta = h !== '00' ? `${h}:${m}:${s}` : `${m}:${s}`;
        return `ETA: ${eta}`;
      }

      return '';
    },
  },

  methods: {
    ...mapActions(useFlashStore, ['cancelFlash']),
  },
});
</script>

<style scoped></style>
