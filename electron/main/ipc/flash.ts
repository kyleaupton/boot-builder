import { ipcMain, IpcMainInvokeEvent } from 'electron';
import FlashWindows from '../flash/FlashWindows';
import FlashMacOS from '../flash/FlashMacOS';

type _Flash = FlashWindows | FlashMacOS;

const flashes = new Map<string, _Flash>();

export const removeFlash = (id: string) => {
  flashes.delete(id);
};

export default function start() {
  ipcMain.handle(
    '/flash/windows',
    async (
      event: IpcMainInvokeEvent,
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
  );

  ipcMain.handle(
    '/flash/macOS',
    (
      event: IpcMainInvokeEvent,
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
  );

  ipcMain.handle(
    '/flash/cancel',
    (event: IpcMainInvokeEvent, { id }: { id: string }) => {
      const flash = flashes.get(id);

      if (flash) {
        flash.cancel();
      }
    },
  );
}
