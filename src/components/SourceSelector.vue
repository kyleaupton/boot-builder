<template>
  <!-- <FloatLabel> -->
  <InputGroup :class="{ 'p-disabled': !chosenOs }">
    <InputGroupAddon>
      <font-awesome-icon
        class="drive-section-header-icon"
        :icon="['fas', 'hard-drive']"
      />
    </InputGroupAddon>

    <Dropdown
      v-model="value"
      input-id="source-selector"
      placeholder="Select Source"
    />
  </InputGroup>

  <!-- <label for="source-selector">Select Source</label>
  </FloatLabel> -->
</template>

<script lang="ts">
import { defineComponent } from 'vue';
// import FloatLabel from 'primevue/floatlabel';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';

export default defineComponent({
  name: 'SourceSelector',

  components: {
    // FloatLabel,
    InputGroup,
    InputGroupAddon,
  },

  props: {
    // v-model
    modelValue: {
      type: String,
      required: true,
    },

    chosenOs: {
      type: String,
      default: '',
    },
  },

  emits: ['update:modelValue'],

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
