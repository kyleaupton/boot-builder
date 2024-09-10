import { randomBytes } from 'crypto';
import { exists } from '@main/utils/fs';
import { expose, sendProgress, executeCommand } from '.';

const validate = async ({ sourcePath }: { sourcePath: string }) => {
  // USB must be at least 14GB in size
  // sourcePath must have the relative `Contents/Resources/createinstallmedia`
  const scriptExists = await exists(
    `${sourcePath}/Contents/Resources/createinstallmedia`,
  );

  if (!scriptExists) {
    throw Error('createinstallmedia script not present');
  }
};

const eraseDrive = async ({
  id,
  targetVolume,
}: {
  id: string;
  targetVolume: string;
}) => {
  // Erase USB as Mac OS Extended
  // sendProgress(`Erasing ${targetVolume}`);
  sendProgress({
    id,
    activity: `Erasing ${targetVolume}`,
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
  });

  try {
    // https://support.apple.com/en-us/HT201372
    // Acording to the Apple Docs, the drive needs to be pre-formatted
    // as JHFS+...? Would love to verify that sometime.
    // The createinstallmedia script will format it anyways...

    const temporaryDiskName = randomBytes(4).toString('hex');

    // Erase disk
    await executeCommand(
      'diskutil',
      ['eraseDisk', 'JHFS+', temporaryDiskName, 'GPT', targetVolume],
      {},
    );

    // TODO: verify the USB mounted back at the same path
  } catch (e) {
    throw Error(`Failed to erase ${targetVolume}`);
  }
};

const executeInstallerScript = async ({
  id,
  targetVolume,
  sourcePath,
}: {
  id: string;
  targetVolume: string;
  sourcePath: string;
}) => {
  sendProgress({
    id,
    activity: 'Running createinstallmedia script',
    done: false,
    transferred: -1,
    speed: -1,
    percentage: -1,
    eta: -1,
  });

  await executeCommand(
    `${sourcePath}/Contents/Resources/createinstallmedia`,
    ['--volume', targetVolume],
    {},
    {
      onOut: (data) => {
        // Parse output of script to get progress and generate eta
        console.log(data);
      },
      onErr: (error) => {
        console.log(error);
      },
    },
  );
};

export interface FlashMacOSWorkerOptions {
  id: string;
  sourcePath: string;
  targetVolume: string;
}

expose<FlashMacOSWorkerOptions, void>({
  fn: async ({ id, sourcePath, targetVolume }: FlashMacOSWorkerOptions) => {
    await validate({ sourcePath });
    await eraseDrive({ id, targetVolume });
    await executeInstallerScript({ id, targetVolume, sourcePath });
  },
});
