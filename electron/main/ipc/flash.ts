import { ipcMain, BrowserWindow, IpcMainInvokeEvent } from 'electron';
import { spawn } from 'child_process';
import { getPath } from '../utils/lib';
import { exec } from '../utils/child_process';

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

      const procHandlers = {
        onStdout: (data: string) => {
          sender.webContents.send(`flash-${id}-stdout`, data);
        },
        onStderr: (data: string) => {
          sender.webContents.send(`flash-${id}-stderr`, data);
        },
      };

      // Erase drive
      sender.webContents.send(`flash-${id}-activity`, `Erasing ${volume}`);
      const diskName = 'WIN10';

      await run(
        'diskutil',
        ['eraseDisk', 'MS-DOS', diskName, 'MBR', volume],
        procHandlers,
      );

      // Mount ISO volume
      sender.webContents.send(`flash-${id}-activity`, `Mounting ${isoFile}`);
      const { stdout } = await exec(`hdiutil mount ${isoFile}`);
      const isoMountedPath = stdout.trim().split(/\s+/)[1];

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
