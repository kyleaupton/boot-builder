import { reactive } from 'vue';
import { defineStore } from 'pinia';
import Drive from '@/api/Drive';
import { Item2, SpusbdataType, Medum } from '@/types/disks';

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
      const usbData = (await window.api.getUsbData()).SPUSBDataType;
      const found: Item2[] = [];

      const isDevice = (item: any): item is Item2 => { // eslint-disable-line
        return item._name && item.serial_num && item.Media && !item._items;
      };

      const isController = (item: any): item is SpusbdataType => { // eslint-disable-line
        return item._items;
      };

      const checkController = (controller: SpusbdataType) => {
        if (controller._items) {
          for (const item of controller._items) {
            if (isDevice(item)) {
              found.push(item);
            } else if (isController(item)) {
              checkController(item);
            }
          }
        }
      };

      // I want to recurse through and find any key that is `_items` and is an array
      for (const baseItem of usbData) {
        checkController(baseItem);
      }

      this.items = found
        .filter((item) => item.Media && item.Media.length === 1)
        .map((item) => {
          const _media = item.Media as Medum[];
          const { bsd_name } = _media[0];
          const usbData = items.find((x) => x.DeviceIdentifier === bsd_name);

          if (!usbData) {
            throw Error();
          }

          return reactive(new Drive({ ...usbData, ...item }));
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
