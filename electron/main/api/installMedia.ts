import { spawn } from 'child_process';
import { exec } from '../utils/child_process';

const run = (cmd: string, args: string[]) => {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn(cmd, args, { stdio: ['inherit', 'inherit', 'inherit'] });

    proc.on('close', () => {
      resolve();
    });

    proc.on('error', () => {
      reject();
    });
  });
};

export const create = async ({
  isoFile,
  volume,
}: {
  isoFile: string;
  volume: string;
}) => {
  const diskName = 'WIN10';

  // Erase drive
  await run('diskutil', ['eraseDisk', 'MS-DOS', diskName, 'MBR', volume]);

  // Mount ISO volume
  const { stdout } = await exec(`hdiutil mount ${isoFile}`);
  const isoMountedPath = stdout.trim().split(/\s+/)[1];

  // Copy over everything minus the big file
  await run('rsync', [
    '--progress',
    '-vha',
    '--exclude=sources/install.wim',
    `${isoMountedPath}/`,
    `/Volumes/${diskName}/`,
  ]);

  // Next copy the big file
  await run('wimlib-imagex', [
    'split',
    `${isoMountedPath}/sources/install.wim`,
    `/Volumes/${diskName}/sources/install.swm`,
    '3800',
  ]);
};
