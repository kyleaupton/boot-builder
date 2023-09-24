import { defineStore } from 'pinia';
import { AllDisksAndPartition, Item2 } from '../../electron/main/api';

type t_state = {
  loading: boolean; // if we're currently grabbing
  loaded: boolean; // if items is not null
  items: (AllDisksAndPartition & Item2)[] | null;
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

      const items = (await window.api.getDisks()).AllDisksAndPartitions;

      const usbData = await window.api.getUsbData();
      const found: (Item2 & { key: string })[] = [];

      for (const controller of usbData.SPUSBDataType) {
        if (controller._items) {
          for (const hub of controller._items) {
            if (hub._items) {
              for (const device of hub._items) {
                if (device.Media) {
                  const foundItem = device.Media.find((x) =>
                    items.find((y) => y.DeviceIdentifier === x.bsd_name),
                  );

                  if (foundItem) {
                    found.push({ ...device, key: foundItem.bsd_name });
                  }
                }
              }
            }
          }
        }
      }

      this.items = items.map((x) => {
        const usbData = found.find((y) => y.key === x.DeviceIdentifier);

        if (usbData) {
          return {
            ...usbData,
            ...x,
          };
        }

        throw Error();
      });

      this.loading = false;
      this.loaded = true;
    },
  },
});
