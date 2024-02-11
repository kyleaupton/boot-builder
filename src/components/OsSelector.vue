<template>
  <InputGroup>
    <InputGroupAddon>
      <font-awesome-icon :icon="['fas', 'display']" />
    </InputGroupAddon>

    <Dropdown
      v-model="value"
      :options="os"
      placeholder="Select OS"
      option-label="label"
      option-value="value"
    />
  </InputGroup>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';

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

  data() {
    return {
      os: [
        {
          label: 'Windows',
          value: 'windows',
        },
        {
          label: 'macOS',
          value: 'macos',
        },
      ],
    };
  },

  computed: {
    value: {
      get() {
        return this.modelValue;
      },
      set(value: string) {
        this.$emit('update:modelValue', value);
      },
    },
  },
});
</script>

<style scoped></style>
