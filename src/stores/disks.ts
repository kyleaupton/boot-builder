import { defineStore } from 'pinia';
import { AllDisksAndPartition } from '../../electron/main/api';

type t_state = {
  loading: boolean; // if we're currently grabbing
  loaded: boolean; // if items is not null
  items: AllDisksAndPartition[] | null;
};

export const useDisksStore = defineStore('disks', {
  state: () =>
    ({
      loading: false,
      loaded: false,
      items: null,
    }) as t_state,

  actions: {
    async getDisks() {
      this.loading = true;

      this.items = (await window.api.getDisks()).AllDisksAndPartitions;

      this.loading = false;
      this.loaded = true;
    },
  },
});
