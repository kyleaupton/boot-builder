import Flash from './Flash';

/**
 * Map of flashes by id. This is used to keep track of flashes that are in progress.
 */
const flashes = new Map<string, Flash<any, any>>();

export const addFlash = (flash: Flash<any, any>) => {
  flashes.set(flash.id, flash);
};

export const getFlash = (id: string) => {
  return flashes.get(id);
};

export const removeFlash = (id: string) => {
  flashes.delete(id);
};

export { Flash };
