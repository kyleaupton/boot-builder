import { defineStore } from 'pinia';

type t_state = {
  chosenDrive: string;
};

export const useLayoutStore = defineStore('layout', {
  state: () =>
    ({
      chosenDrive: '',
    }) as t_state,

  actions: {},
});
