// Worker thread for windows flash
import { stat } from 'fs/promises';
import { resolve, dirname, basename } from 'path';
import { copy } from '@kyleupton/glob-copy';
import { exec } from '@main/utils/child_process';
import { exists, dirSize } from '@main/utils/fs';
import { getPath } from '@main/utils/lib';
import { humanReadableToBytes } from '@main/utils/bytes';
import { createWorker } from '.';
import Worker, { IWorker } from './Worker';
import { SerializedFlash } from '@shared/flash';

const mountIsoVolume = async ({
  id,
  sourcePath,
}: {
  id: string;
  sourcePath: string;
}): Promise<string> => {
  sendProgress({
    id,
    activity: `Mounting ${sourcePath}`,
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
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
  id,
  targetVolume,
  sourcePath,
}: {
  id: string;
  targetVolume: string;
  sourcePath: string;
}): Promise<string> => {
  sendProgress({
    id,
    activity: `Erasing ${targetVolume}`,
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
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
    await executeCommand(
      'diskutil',
      ['eraseDisk', 'MS-DOS', diskName, 'MBR', targetVolume],
      {},
    );

    return diskName;
  } catch (e) {
    throw Error(`Failed to erase ${targetVolume}`);
  }
};

const copyFiles = async ({
  id,
  mountedIsoPath,
  diskName,
  isPackaged,
}: {
  id: string;
  mountedIsoPath: string;
  diskName: string;
  isPackaged: boolean;
}) => {
  const activity = 'Copying files';

  sendProgress({
    id,
    activity,
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
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

        sendProgress({
          id,
          activity,
          done: false,
          transferred: progress.transferred,
          speed: progress.speed,
          percentage: (progress.transferred / totalSize) * 100,
          eta: remaining / progress.speed,
        });
      },
    });

    // Next copy the big file
    const wimlibImagex = getPath(
      { name: 'wimlib', bin: 'wimlib-imagex' },
      { isPackaged },
    );
    const dir = dirname(wimlibImagex);
    const name = basename(wimlibImagex);
    const start = Date.now();
    let secondFileTransferred = 0;

    let progress = {
      id,
      activity,
      done: false,
      transferred: 0,
      speed: 0,
      percentage: 0,
      eta: 0,
    };

    const intervalId = setInterval(() => {
      sendProgress(progress);
    }, 1000);

    await executeCommand(
      name,
      [
        'split',
        `${mountedIsoPath}/sources/install.wim`,
        `/Volumes/${diskName}/sources/install.swm`,
        '3800',
      ],
      {
        cwd: dir,
      },
      {
        onOut: (data) => {
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

            progress = {
              id,
              activity,
              done: false,
              transferred,
              speed,
              percentage: (transferred / totalSize) * 100,
              eta: remaining / speed,
            };
          }
        },
      },
    );

    clearInterval(intervalId);
  } catch (e) {
    console.log(e);
    throw Error('Failed to copy files');
  }
};

const cleanUp = async ({
  id,
  targetVolume,
  mountedIsoPath,
}: {
  id: string;
  targetVolume: string;
  mountedIsoPath: string;
}) => {
  sendProgress({
    id,
    activity: 'Cleaning up',
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
  });

  try {
    await executeCommand('diskutil', ['eject', targetVolume], {});
    await executeCommand('umount', [mountedIsoPath], {});
  } catch (e) {
    throw Error('Failed to clean up');
  }
};

export interface FlashWindowsWorkerOptions {
  id: string;
  sourcePath: string;
  targetVolume: string;
}

export interface FlashWindowsWorkerOptionsFinal
  extends FlashWindowsWorkerOptions {
  isPackaged: boolean;
}

// expose(
//   async ({
//     id,
//     sourcePath,
//     targetVolume,
//     isPackaged,
//   }: FlashWindowsWorkerOptionsFinal) => {
//     // Mount target ISO
//     const mountedIsoPath = await mountIsoVolume({ id, sourcePath });
//     // Validate ISO
//     await validateIso({ mountedIsoPath });
//     // Erase drive
//     const diskName = await eraseDrive({ id, targetVolume, sourcePath });
//     // Copy files
//     await copyFiles({ id, mountedIsoPath, diskName, isPackaged });
//     // Clean up
//     await cleanUp({ id, targetVolume, mountedIsoPath });
//   },
// );

// export default class FlashWindowsWorker extends Worker implements IWorker {
//   constructor() {
//     super();
//   }

//   run() {
//     return;
//   }
// }

export default createWorker(
  async (
    state,
    { sourcePath, targetVolume }: { sourcePath: string; targetVolume: string },
  ) => {
    console.log('FlashWindowsWorker', sourcePath, targetVolume);
    for (let i = 0; i < 100; i++) {
      state.progress = {
        activity: 'Copying files',
        transferred: i * 100,
        speed: i * 100,
        percentage: i,
        eta: i * 100,
      };
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
    return;
  },
);
