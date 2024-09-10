import { createIpcHandlers } from 'typed-electron-ipc';
import FlashWindows from '../flash/FlashWindows';
import FlashMacOS from '../flash/FlashMacOS';

type _Flash = FlashWindows | FlashMacOS;

const flashes = new Map<string, _Flash>();

export const removeFlash = (id: string) => {
  flashes.delete(id);
};

export const flashIpc = () =>
  createIpcHandlers({
    '/flash/windows': async (
      event,
      {
        id,
        sourcePath,
        targetVolume,
      }: { sourcePath: string; targetVolume: string; id: string },
    ) => {
      const flash = new FlashWindows({
        id,
        sourcePath,
        targetVolume,
      });

      flashes.set(id, flash);
      flash.start();
    },

    '/flash/macOS': async (
      event,
      {
        id,
        sourcePath,
        targetVolume,
      }: { id: string; sourcePath: string; targetVolume: string },
    ) => {
      const flash = new FlashMacOS({
        id,
        sourcePath,
        targetVolume,
      });

      flashes.set(id, flash);
      flash.start();
    },

    '/flash/cancel': async (event, id: string) => {
      const flash = flashes.get(id);

      if (flash) {
        flash.cancel();
      }
    },
  });
