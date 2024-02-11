import { defineStore } from 'pinia';
import drives from 'drivelist';

type t_state = {
  items: Array<drives.Drive>;
};

export const useDisksStore = defineStore('disks', {
  state: () =>
    ({
      items: [],
    }) as t_state,

  actions: {
    async getDisks() {
      this.items = (await window.api.getDisks()).filter((drive) => drive.isUSB);
    },

    registerUsbEvents() {
      window.api.ipc.recieve('/usb/attached', async () => {
        const currentNum = this.items.length || 0;

        const attemptRefresh = async (attempt = 1): Promise<void> => {
          if (attempt <= 5) {
            await this.getDisks();

            if (this.items.length === currentNum + 1) {
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
