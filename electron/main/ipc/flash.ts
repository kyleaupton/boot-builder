import { createIpcHandlers } from 'typed-electron-ipc';
import { addFlash, getFlash } from '@main/flash';
import FlashWindows from '@main/flash/FlashWindows';
import FlashMacOS from '@main/flash/FlashMacOS';

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

      addFlash(flash);
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

      addFlash(flash);
      flash.start();
    },

    '/flash/cancel': async (event, id: string) => {
      const flash = getFlash(id);

      if (flash) {
        flash.cancel();
      }
    },
  });
