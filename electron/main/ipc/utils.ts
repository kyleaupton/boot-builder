import { nativeImage, ipcMain, IpcMainInvokeEvent } from 'electron';
import fs from 'fs/promises';
import os from 'os';

export default function start() {
  ipcMain.handle('/utils/os/home', () => {
    return os.homedir();
  });

  ipcMain.handle(
    '/utils/fs/readdir',
    (event: IpcMainInvokeEvent, { path }: { path: string }) => {
      return fs.readdir(path);
    },
  );

  ipcMain.handle(
    '/utils/app/fileIcon',
    async (event: IpcMainInvokeEvent, { path }: { path: string }) => {
      const result = await nativeImage.createThumbnailFromPath(path, {
        width: 200,
        height: 200,
      });
      return result.toDataURL();
    },
  );
}
