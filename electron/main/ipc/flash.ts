import { ipcMain, BrowserWindow, IpcMainInvokeEvent } from 'electron';
import { stat } from 'fs/promises';
import { spawn } from 'child_process';
import { resolve } from 'path';
import { copy } from '@kyleupton/node-rsync';
import { getPath } from '../utils/lib';
import { exec } from '../utils/child_process';
import { dirSize } from '../utils/fs';
import { humanReadableToBytes } from '../utils/bytes';

const exists = async (path: string) => {
  try {
    await stat(path);
    return true;
  } catch (e) {
    return false;
  }
};

const run = (
  cmd: string,
  args: string[],
  {
    onStdout,
    onStderr,
  }: { onStdout: (data: string) => void; onStderr: (data: string) => void }, // eslint-disable-line
) => {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn(cmd, args);

    proc.stdout.on('data', (data) => {
      onStdout(data.toString());
    });

    proc.stderr.on('data', (data) => {
      onStderr(data.toString());
    });

    proc.on('close', () => {
      resolve();
    });

    proc.on('error', (error) => {
      reject(error);
    });
  });
};

export default function start() {
  ipcMain.handle(
    '/flash',
    async (
      event: IpcMainInvokeEvent,
      { isoFile, volume, id }: { isoFile: string; volume: string; id: string },
    ) => {
      const sender = BrowserWindow.fromWebContents(event.sender);

      const procHandlers = {
        onStdout: (data: string) => {
          console.log(data);
          sender.webContents.send(`flash-${id}-stdout`, data);
        },
        onStderr: (data: string) => {
          console.log(data);
          sender.webContents.send(`flash-${id}-stderr`, data);
        },
      };

      //
      // Mount ISO volume
      //
      sender.webContents.send(`flash-${id}-activity`, `Mounting ${isoFile}`);
      const { stdout } = await exec(`hdiutil mount ${isoFile}`);
      const isoMountedPath = stdout.trim().split(/\s+/)[1];

      //
      // Validate
      //
      const needed = ['sources/install.wim'];
      for (const rel of needed) {
        const abs = resolve(isoMountedPath, rel);

        if (!(await exists(abs))) {
          throw Error(`Validation: ${rel} not found`);
        }
      }

      //
      // Erase drive
      //
      sender.webContents.send(`flash-${id}-activity`, `Erasing ${volume}`);

      // Get disk name
      // Windows 10: `Win10_22H2_English_x64v1.iso`
      // Windows 11: `Win11_22H2_English_x64v2.iso`
      const diskName = isoFile
        .split('/')
        .slice(-1)[0]
        .split('_')[0]
        .toUpperCase();

      await run(
        'diskutil',
        ['eraseDisk', 'MS-DOS', diskName, 'MBR', volume],
        procHandlers,
      );

      //
      // Copy files
      //
      sender.webContents.send(`flash-${id}-activity`, `Copying files`);

      // Get stats for progress reporting
      const totalSize = await dirSize(isoMountedPath);
      const bigFileSize = await stat(`${isoMountedPath}/sources/install.wim`);
      let transferred = 0;

      // Copy over everything minus the big file
      await copy({
        source: `${isoMountedPath}/`,
        destination: `/Volumes/${diskName}`,
        options: {
          archive: true,
          exclude: 'sources/install.wim',
        },
        onProgress: (progress) => {
          const remaining = totalSize - progress.transferred;
          transferred += progress.transferred;

          const payload = {
            transferred: progress.transferred,
            speed: progress.speed,
            percentage: (progress.transferred / totalSize) * 100,
            eta: remaining / progress.speed,
          };

          sender.webContents.send(`flash-${id}-copy-eta`, payload);
        },
      });

      // Next copy the big file
      const wimlibImagex = getPath('wimlib');

      const start = Date.now();
      await run(
        wimlibImagex,
        [
          'split',
          `${isoMountedPath}/sources/install.wim`,
          `/Volumes/${diskName}/sources/install.swm`,
          '3800',
        ],
        {
          onStdout: (data) => {
            const match = data.match(
              /(?<part>\d{1,7}\s\w{2,4}) of (?<whole>\d{1,7}\s\w{2,4})/,
            );

            if (match && match.groups) {
              const { part } = match.groups;
              const { bytes } = humanReadableToBytes(part);
              transferred += bytes;
              const remaining = bigFileSize.size - bytes;
              const speed = bytes / ((Date.now() - start) / 1000);

              const payload = {
                transferred,
                speed,
                percentage: (transferred / totalSize) * 100,
                eta: remaining / speed,
              };

              sender.webContents.send(`flash-${id}-copy-eta`, payload);
            }

            console.log(data);
            sender.webContents.send(`flash-${id}-stdout`, data);
          },
          onStderr: () => {},
        },
      );

      // Clean up
      sender.webContents.send(`flash-${id}-activity`, `Cleaning up`);

      // await run('umount', [isoMountedPath], procHandlers);
      // await run('umount', [isoMountedPath], procHandlers);
    },
  );
}
