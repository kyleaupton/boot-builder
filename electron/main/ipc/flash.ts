import { ipcMain, IpcMainInvokeEvent } from 'electron';
import FlashWindows from '../flash/FlashWindows';
import FlashMacOS from '../flash/FlashMacOS';

const flashes: { [key: string]: FlashWindows | FlashMacOS } = {};

export const removeFlash = (id: string) => {
  delete flashes[id];
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
      flashes[id] = new FlashWindows({
        id,
        sourcePath,
        targetVolume,
      });

      flashes[id].run();
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
      flashes[id] = new FlashMacOS({
        id,
        sourcePath,
        targetVolume,
      });

      flashes[id].run();
    },
  );
}
