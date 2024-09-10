import FlashMacOS from './macos/FlashMacOS';
import FlashWindows from './windows/FlashWindows';

type Flash = FlashMacOS | FlashWindows;
const flashes = new Map<string, Flash>();

export const addFlash = (flash: Flash) => {
  flashes.set(flash.id, flash);
};

export const getFlash = (id: string) => {
  return flashes.get(id);
};

export const removeFlash = (id: string) => {
  flashes.delete(id);
};
