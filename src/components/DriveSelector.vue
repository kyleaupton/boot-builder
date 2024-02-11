<template>
  <InputGroup>
    <InputGroupAddon>
      <font-awesome-icon
        class="drive-section-header-icon"
        :icon="['fas', 'hard-drive']"
      />
    </InputGroupAddon>

    <Dropdown
      v-model="value"
      :options="dropdownOptions"
      option-label="label"
      option-value="value"
      placeholder="Select target"
    />
  </InputGroup>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';
import prettyBytes from 'pretty-bytes';
import { useDisksStore } from '@/stores/disks';

export default defineComponent({
  name: 'DriveSelector',

  components: {
    InputGroup,
    InputGroupAddon,
  },

  props: {
    // v-model
    modelValue: {
      type: String,
      required: true,
    },
  },

  emits: ['update:modelValue'],

  computed: {
    ...mapState(useDisksStore, ['items']),

    // v-model
    value: {
      get() {
        return this.modelValue;
      },
      set(value: string) {
        this.$emit('update:modelValue', value);
      },
    },

    dropdownOptions() {
      return this.items.map((drive) => ({
        label: `${drive.description} (${prettyBytes(drive.size ?? 0)})`,
        value: drive.devicePath,
      }));
    },
  },
});
</script>

<style scoped></style>
