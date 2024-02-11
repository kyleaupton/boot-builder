<template>
  <InputGroup>
    <InputGroupAddon>
      <font-awesome-icon
        class="drive-section-header-icon"
        :icon="['fas', 'hard-drive']"
      />
    </InputGroupAddon>

    <Dropdown v-model="value" :items="disks" placeholder="Select target" />
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
      disks: [],
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

  created() {
    console.log('DriveSelector created');
  },

  methods: {
    async getDisks() {
      console.log('getDisks');
    },
  },
});
</script>

<style scoped></style>
