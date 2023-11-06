import { app, ipcMain, IpcMainInvokeEvent, nativeImage } from 'electron';
import fs from 'fs/promises';
import { join } from 'path';
import crypto from 'crypto';
import glob from 'fast-glob';
import { exec } from '../utils/child_process';

export default function start() {
  ipcMain.handle(
    '/utils/fs/readdir',
    (event: IpcMainInvokeEvent, { path }: { path: string }) => {
      return fs.readdir(path);
    },
  );

  ipcMain.handle(
    '/utils/app/getFileIcon',
    async (event: IpcMainInvokeEvent, { path }: { path: string }) => {
      return (await app.getFileIcon(path)).toDataURL();
    },
  );

  ipcMain.handle(
    '/utils/app/getPath',
    async (event: IpcMainInvokeEvent, { name }: { name: 'downloads' }) => {
      return app.getPath(name);
    },
  );

  ipcMain.handle(
    '/utils/macApp/getIcon',
    async (event: IpcMainInvokeEvent, { path }: { path: string }) => {
      // Ensure path is macOS App
      // - Path ends with .app
      // - path/Resources/*.icns must exist
      if (!path.endsWith('.app')) {
        throw Error('Invalid path');
      }

      const iconFiles = await glob('Contents/Resources/*.icns', { cwd: path });
      if (iconFiles.length < 1) {
        throw Error('No icns files found in app content');
      }

      // For now just grab the first icon file
      const icns = iconFiles[0];

      // Create temp path where the PNG will live
      const tempPath = join(
        app.getPath('temp'),
        `${crypto.randomBytes(6).toString('hex')}.png`,
      );

      // Form path of source .icns file
      const sourceIconPath = join(path, icns);

      // Do the converstion
      await exec(`sips -s format png "${sourceIconPath}" --out ${tempPath}`);

      // Generate base64 with electron's nativeImage
      const base64 = nativeImage.createFromPath(tempPath).toDataURL();

      // Delete temp file
      await fs.unlink(tempPath);

      return base64;
    },
  );
}
