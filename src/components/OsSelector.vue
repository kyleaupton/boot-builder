<template>
  <InputGroup :class="{ 'p-disabled': flashing }">
    <InputGroupAddon>
      <font-awesome-icon :icon="['fas', 'display']" />
    </InputGroupAddon>

    <Dropdown
      v-model="value"
      :options="os"
      option-label="label"
      option-value="value"
      placeholder="Select OS"
    >
      <template #value="slotProps">
        <!--  -->
        <div v-if="slotProps.value" class="os-selector-option">
          <font-awesome-icon
            :icon="os.find((x) => x.value === slotProps.value)?.icon || []"
          />
          <div>
            {{
              os.find((x) => x.value === slotProps.value)?.label ||
              slotProps.value
            }}
          </div>
        </div>
        <span v-else>
          {{ slotProps.placeholder }}
        </span>
      </template>
      <template #option="slotProps">
        <div class="os-selector-option">
          <font-awesome-icon :icon="slotProps.option.icon" />
          <div>{{ slotProps.option.label }}</div>
        </div>
      </template>
    </Dropdown>
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
    flashing: {
      type: Boolean,
      default: false,
    },
  },

  emits: ['update:modelValue'],

  data() {
    return {
      os: [
        {
          label: 'Windows',
          value: 'windows',
          icon: ['fab', 'windows'],
        },
        {
          label: 'macOS',
          value: 'macos',
          icon: ['fab', 'apple'],
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

<style scoped>
.os-selector-option {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
</style>
