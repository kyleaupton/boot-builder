import { ipcMain, BrowserWindow, IpcMainInvokeEvent } from 'electron';
import { stat } from 'fs/promises';
import { spawn } from 'child_process';
import { resolve } from 'path';
import { getPath } from '../utils/lib';
import { exec } from '../utils/child_process';

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
      // Browser window of sender
      const sender = BrowserWindow.fromWebContents(event.sender);

      // Mount ISO volume
      sender.webContents.send(`flash-${id}-activity`, `Mounting ${isoFile}`);
      const { stdout } = await exec(`hdiutil mount ${isoFile}`);
      const isoMountedPath = stdout.trim().split(/\s+/)[1];

      // Validate
      const needed = ['sources/install.wim'];
      for (const rel of needed) {
        const abs = resolve(isoMountedPath, rel);

        if (!(await exists(abs))) {
          throw Error(`Validation: ${rel} not found`);
        }
      }

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

      // Erase drive
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

      // Copy over everything minus the big file
      sender.webContents.send(`flash-${id}-activity`, `Copying files`);
      await run(
        'rsync',
        [
          '--progress',
          '-vha',
          '--exclude=sources/install.wim',
          `${isoMountedPath}/`,
          `/Volumes/${diskName}/`,
          '--info=progress2',
        ],
        procHandlers,
      );

      // Next copy the big file
      const wimlibImagex = getPath('wimlib');

      await run(
        wimlibImagex,
        [
          'split',
          `${isoMountedPath}/sources/install.wim`,
          `/Volumes/${diskName}/sources/install.swm`,
          '3800',
        ],
        procHandlers,
      );
    },
  );
}
