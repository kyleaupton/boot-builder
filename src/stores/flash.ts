import { defineStore } from 'pinia';
import { SerializedFlash } from '@shared/flash';
import { showConfirmDialog } from '@/api/dialog';

type State = {
  selectedDrive: string | undefined;
  selectedOs: string | undefined;
  selectedSource: string | undefined;

  // Stores the current flash state
  flash: SerializedFlash | undefined;
  // flash: {
  //   id: 'FAXuK8yKE1MTjWS8jSxn-',
  //   activity: '',
  //   done: false,
  //   canceled: true,
  //   transferred: -1,
  //   speed: -1,
  //   percentage: -1,
  //   eta: -1,
  // },
};

export const useFlashStore = defineStore('flash', {
  state: (): State => ({
    selectedDrive: undefined,
    selectedOs: undefined,
    selectedSource: undefined,
    flash: undefined,
  }),

  getters: {
    isFlashing: (state) => {
      return state.flash !== undefined && !state.flash.done;
    },

    isFinished: (state) => {
      return (
        state.flash !== undefined &&
        (state.flash.done || state.flash.canceled || state.flash.error)
      );
    },
  },

  actions: {
    async cancelFlash() {
      if (this.flash) {
        const accepted = await showConfirmDialog(
          'Are you sure you want to cancel the flashing process?',
        );

        if (!accepted) {
          return;
        }

        window.ipcInvoke('/flash/cancel', this.flash.id);
      }
    },

    reset() {
      this.selectedDrive = undefined;
      this.selectedOs = undefined;
      this.selectedSource = undefined;
      this.flash = undefined;
    },

    registerFlashEvents() {
      window.onFlashUpdate((flash) => {
        this.flash = flash;
      });
    },
  },
});
