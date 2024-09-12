import { createIpcHandlers } from 'typed-electron-ipc';
import { addFlash, getFlash, Flash } from '@main/flash';
import { _Workers } from '@main/flash/workers';

// {
//   id,
//   sourcePath,
//   targetVolume,
// }: { sourcePath: string; targetVolume: string; id: string },

export const flashIpc = () =>
  createIpcHandlers({
    '/flash/new': async (
      event,
      id: string,
      type: keyof _Workers,
      options: Parameters<_Workers[keyof _Workers]>[1],
    ) => {
      const flash = new Flash(id, type, options);
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
