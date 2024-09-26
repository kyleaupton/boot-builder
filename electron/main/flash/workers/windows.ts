import { stat } from 'node:fs/promises';
import { resolve } from 'node:path';
import { spawn } from 'node:child_process';
import log from 'electron-log/node';
import { copy } from '@kyleupton/glob-copy';
import { SerializedFlash } from '@shared/flash';
import { exec } from '@main/utils/child_process';
import { exists, dirSize } from '@main/utils/fs';
import { getWimlibPath } from '@main/utils/lib';
import { humanReadableToBytes } from '@main/utils/bytes';
import { type Signal } from '@main/utils/signal';
import { createWorker } from './utils';

const logger = log.scope('main/flash-windows');
type State = Signal<SerializedFlash>;

//
// Windows flash worker
//
export default createWorker(
  async (
    state,
    { sourcePath, targetVolume }: { sourcePath: string; targetVolume: string },
  ) => {
    // Mount target ISO
    const mountedIsoPath = await mountIsoVolume({ state, sourcePath });
    // Validate ISO
    await validateIso({ mountedIsoPath });
    // Erase drive
    const diskName = await eraseDrive({ state, targetVolume, sourcePath });
    // Copy files
    await copyFiles({ state, mountedIsoPath, diskName });
    // Clean up
    await cleanUp({ state, targetVolume, mountedIsoPath });
  },
);

const mountIsoVolume = async ({
  state,
  sourcePath,
}: {
  state: State;
  sourcePath: string;
}): Promise<string> => {
  state.set({
    progress: {
      activity: `Mounting ${sourcePath}`,
    },
  });

  try {
    const { stdout } = await exec(`hdiutil mount ${sourcePath}`);
    return stdout.trim().split(/\s+/)[1];
  } catch (e) {
    throw Error(`Failed to mount ${sourcePath}`);
  }
};

const validateIso = async ({ mountedIsoPath }: { mountedIsoPath: string }) => {
  const needed = ['sources/install.wim'];

  for (const rel of needed) {
    const abs = resolve(mountedIsoPath, rel);

    if (!(await exists(abs))) {
      throw Error(`Invalid Windows ISO file - ${rel} not found`);
    }
  }
};

const eraseDrive = async ({
  state,
  targetVolume,
  sourcePath,
}: {
  state: State;
  targetVolume: string;
  sourcePath: string;
}): Promise<string> => {
  state.set({
    progress: {
      activity: `Erasing ${targetVolume}`,
    },
  });

  let diskName = 'WINDOWS';
  try {
    // Get disk name
    // Windows 10: `Win10_22H2_English_x64v1.iso`
    // Windows 11: `Win11_22H2_English_x64v2.iso`
    if (sourcePath.includes('Win10') || sourcePath.includes('Win11')) {
      diskName = sourcePath.split('/').slice(-1)[0].split('_')[0].toUpperCase();
    }

    // Erase disk
    await exec(`diskutil eraseDisk MS-DOS ${diskName} MBR ${targetVolume}`);

    return diskName;
  } catch (e) {
    throw Error(`Failed to erase ${targetVolume}`);
  }
};

const copyFiles = async ({
  state,
  mountedIsoPath,
  diskName,
}: {
  state: State;
  mountedIsoPath: string;
  diskName: string;
}) => {
  state.set({
    progress: {
      activity: 'Copying files',
    },
  });

  try {
    // Get stats for progress reporting
    const totalSize = await dirSize(mountedIsoPath);
    const bigFileSize = await stat(`${mountedIsoPath}/sources/install.wim`);
    let transferred = 0;

    // Copy over everything minus the big file
    await copy({
      source: mountedIsoPath,
      destination: `/Volumes/${diskName}`,
      options: {
        ignore: ['sources/install.wim'],
      },
      onProgress: (progress) => {
        transferred = progress.transferred;
        const remaining = totalSize - progress.transferred;

        state.set({
          progress: {
            transferred: progress.transferred,
            speed: progress.speed,
            percentage: (progress.transferred / totalSize) * 100,
            eta: remaining / progress.speed,
          },
        });
      },
    });

    // Next copy the big file
    // const wimlibImagex = getWimlibPath();
    const wimlibImagex =
      '/Users/kyleupton/Downloads/wimlib-1.14.4/wimlib-install/bin/wimlib-imagex';
    logger.info(`Using wimlib-imagex at ${wimlibImagex}`);
    const start = Date.now();
    let secondFileTransferred = 0;

    await new Promise<void>((resolve, reject) => {
      const proc = spawn(wimlibImagex, [
        'split',
        `${mountedIsoPath}/sources/install.wim`,
        `/Volumes/${diskName}/sources/install.swm`,
        '3800',
      ]);

      proc.stdout.on('data', (buff) => {
        const data = buff.toString();
        const match = data.match(
          /(?<part>\d{1,7}\s\w{2,4}) of (?<whole>\d{1,7}\s\w{2,4})/,
        );

        if (match && match.groups) {
          const { part } = match.groups;
          const { bytes } = humanReadableToBytes(part);
          const delta = bytes - secondFileTransferred;
          secondFileTransferred = bytes;
          transferred += delta;
          const remaining = bigFileSize.size - bytes;
          const speed = bytes / ((Date.now() - start) / 1000);

          state.set({
            progress: {
              transferred,
              speed,
              percentage: (transferred / totalSize) * 100,
              eta: remaining / speed,
            },
          });
        }
      });

      proc.on('error', (e) => {
        logger.debug('got to proc.on("error")');
        reject(e);
      });

      proc.on('exit', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(`Command failed with code ${code}`));
        }
      });
    });
  } catch (e) {
    logger.debug('got to catch');
    logger.error(e);
    throw e;
  }
};

const cleanUp = async ({
  state,
  targetVolume,
  mountedIsoPath,
}: {
  state: State;
  targetVolume: string;
  mountedIsoPath: string;
}) => {
  state.set({
    progress: {
      activity: 'Cleaning up',
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    },
  });

  try {
    await exec(`diskutil eject ${targetVolume}`);
    await exec(`umount ${mountedIsoPath}`);
  } catch (e) {
    throw Error('Failed to clean up');
  }
};
