import { spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { SerializedFlash } from '@shared/flash';
import { exists } from '@main/utils/fs';
import { exec } from '@main/utils/child_process';
import { type Signal } from '@main/utils/signal';
import { createWorker } from './utils';

type State = Signal<SerializedFlash>;

//
// macOS flash worker
//
export default createWorker(
  async (
    state,
    { sourcePath, targetVolume }: { sourcePath: string; targetVolume: string },
  ) => {
    await validate({ sourcePath });
    await eraseDrive({ state, targetVolume });
    await executeInstallerScript({ state, targetVolume, sourcePath });
  },
);

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
  state,
  targetVolume,
}: {
  state: State;
  targetVolume: string;
}) => {
  // Erase USB as Mac OS Extended
  state.set({
    progress: {
      activity: `Erasing ${targetVolume}`,
    },
  });

  try {
    // https://support.apple.com/en-us/HT201372
    // Acording to the Apple Docs, the drive needs to be pre-formatted
    // as JHFS+...? Would love to verify that sometime.
    // The createinstallmedia script will format it anyways...

    const temporaryDiskName = randomBytes(4).toString('hex');

    // Erase disk
    await exec(
      `diskutil eraseDisk JHFS+ ${temporaryDiskName} GPT ${targetVolume}`,
    );

    // TODO: verify the USB mounted back at the same path
  } catch (e) {
    throw Error(`Failed to erase ${targetVolume}`);
  }
};

const executeInstallerScript = async ({
  state,
  targetVolume,
  sourcePath,
}: {
  state: State;
  targetVolume: string;
  sourcePath: string;
}) => {
  state.set({
    progress: {
      activity: 'Running createinstallmedia script',
    },
  });

  await new Promise<void>((resolve, reject) => {
    const proc = spawn(`${sourcePath}/Contents/Resources/createinstallmedia`, [
      '--volume',
      targetVolume,
    ]);

    proc.stdout.on('data', (data) => {
      console.log(data.toString());
    });

    proc.stderr.on('data', (data) => {
      console.error(data.toString());
    });

    proc.on('error', () => {
      reject(new Error('Failed to run createinstallmedia script'));
    });

    proc.on('exit', (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Command failed with code ${code}`));
      }
    });
  });
};
