import { reactive } from 'vue';
import { defineStore } from 'pinia';
import Drive from '@/api/Drive';
import { Item2 } from '../../electron/main/api';

type t_state = {
  loading: boolean; // if we're currently grabbing
  loaded: boolean; // if items is not null
  items: Drive[];
};

export const useDisksStore = defineStore('disks', {
  state: () =>
    ({
      loading: false,
      loaded: false,
      items: [],
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
          return reactive(new Drive({ ...usbData, ...x }));
        }

        throw Error();
      });

      this.loading = false;
      this.loaded = true;
    },

    registerUsbEvents() {
      window.api.ipc.recieve('/usb/attached', async () => {
        const currentNum = this.items?.length || 0;

        const attemptRefresh = async (attempt = 1): Promise<void> => {
          if (attempt <= 5) {
            await this.getDisks();

            if (this.items?.length === currentNum + 1) {
              return;
            } else {
              await new Promise((resolve) => setTimeout(resolve, 1000)); // Wait one second
              attempt++;
              return attemptRefresh(attempt);
            }
          }
        };

        await attemptRefresh();
      });

      window.api.ipc.recieve('/usb/detached', () => {
        this.getDisks();
      });
    },
  },
});
