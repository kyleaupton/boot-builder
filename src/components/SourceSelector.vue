<template>
  <div class="source-wrapper">
    <Button
      v-if="!value"
      label="Choose file"
      icon="pi pi-upload"
      :disabled="!chosenOs"
      @click="showDialog"
    />

    <InputGroup
      v-else
      class="source-preview"
      :class="{ 'p-disabled': flashing }"
    >
      <InputGroupAddon>
        <font-awesome-icon
          class="drive-section-header-icon"
          :icon="['fas', 'file']"
        />
      </InputGroupAddon>
      <InputText v-model="namePreview" :disabled="true" />
      <Button
        v-if="!flashing"
        icon="pi pi-times"
        severity="danger"
        @click="value = ''"
      />
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
    flashing: {
      type: Boolean,
      default: false,
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

    namePreview: {
      get() {
        return window.api.path.basename(this.value);
      },
      set() {
        // do nothing
      },
    },
  },

  methods: {
    async showDialog() {
      type res_windows = Awaited<
        ReturnType<typeof window.api.showOpenIsoDialog>
      >;
      type res_mac = Awaited<ReturnType<typeof window.api.showOpenAppDialog>>;

      let res: res_windows | res_mac;

      if (this.chosenOs === 'windows') {
        res = await window.api.showOpenIsoDialog();
      } else if (this.chosenOs === 'macos') {
        res = await window.api.showOpenAppDialog();
      } else {
        throw Error('Unsupprted OS');
      }

      if (!res.filePaths[0]) {
        throw Error('No file selected');
      }

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

.source-preview .p-component:disabled {
  opacity: 1;
}
</style>
