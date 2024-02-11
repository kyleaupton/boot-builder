import { defineStore } from 'pinia';
import { Item2, SpusbdataType, Medum, t_drive } from '@/types/disks';

type t_state = {
  items: Array<t_drive>;
};

export const useDisksStore = defineStore('disks', {
  state: () =>
    ({
      items: [],
    }) as t_state,

  actions: {
    async getDisks() {
      const items = (await window.api.getDisks()).AllDisksAndPartitions;
      const usbData = (await window.api.getUsbData()).SPUSBDataType;
      const found: Item2[] = [];

      const isDevice = (item: any): item is Item2 => { // eslint-disable-line
        return item._name && item.serial_num && item.bcd_device;
      };

      const isController = (item: any): item is SpusbdataType => { // eslint-disable-line
        return item.pci_device && item._items;
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

          return { ...usbData, ...item };
        });
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
