import { defineStore } from 'pinia';
import { getDisks } from '@/api/disks';

interface State {
  items: Awaited<ReturnType<typeof getDisks>>;
}

export const useDisksStore = defineStore('disks', {
  state: (): State => ({
    items: [],
  }),

  actions: {
    async getDisks() {
      // this.items = (await getDisks()).filter((drive) => drive.isUSB);
      this.items = (await getDisks()).filter(
        (drive) => drive.isRemovable && drive.mountpoints.length > 0,
      );
    },

    registerUsbEvents() {
      window.onUsbAttached(async () => {
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

      window.onUsbDetached(() => {
        this.getDisks();
      });
    },
  },
});
