<template>
  <div class="source-wrapper">
    <Button
      v-if="!value"
      label="Choose file"
      icon="pi pi-upload"
      :disabled="!chosenOs"
      @click="showDialog"
    />

    <InputGroup v-else>
      <InputGroupAddon>
        <font-awesome-icon
          class="drive-section-header-icon"
          :icon="['fas', 'file']"
        />
      </InputGroupAddon>
      <InputText v-model="value" :disabled="true" />
      <Button icon="pi pi-times" severity="danger" @click="value = ''" />
    </InputGroup>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import InputGroup from 'primevue/inputgroup';
import InputGroupAddon from 'primevue/inputgroupaddon';

export default defineComponent({
  name: 'SourceSelector',

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

  methods: {
    async showDialog() {
      const res = await window.api.showOpenIsoDialog();
      this.value = res.filePaths[0];
    },
  },
});
</script>

<style scoped>
.source-wrapper {
  display: flex;
  justify-content: center;
}

.source-selector-header {
  display: flex;
  justify-content: space-between;
  width: 100%;
  gap: 0.75rem;
}

.source-selector-header-button {
  font-size: 0.5rem;
}

.source-wrapper .p-disabled,
.p-component:disabled {
  opacity: 1;
}
</style>
