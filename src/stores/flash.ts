import { defineStore } from 'pinia';
import { SerializedFlash } from '@shared/flash';

type State = {
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
    flash: undefined,
  }),

  actions: {
    registerFlashEvents() {
      window.onFlashUpdate((flash) => {
        this.flash = flash;
      });
    },
  },
});
