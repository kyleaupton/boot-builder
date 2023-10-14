import { ipcMain, IpcMainInvokeEvent } from 'electron';
import FlashWindows from '../flash/FlashWindows';

const flashes: { [key: string]: FlashWindows } = {};

export const removeFlash = (id: string) => {
  delete flashes[id];
};

export default function start() {
  ipcMain.handle(
    '/flash',
    async (
      event: IpcMainInvokeEvent,
      { isoFile, volume, id }: { isoFile: string; volume: string; id: string },
    ) => {
      flashes[id] = new FlashWindows({
        id,
        isoFile,
        targetVolume: volume,
      });

      flashes[id].run();
    },
  );
}
