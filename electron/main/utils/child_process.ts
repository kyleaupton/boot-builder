import {
  exec as oldExec,
  spawn,
  type SpawnOptionsWithoutStdio,
} from 'node:child_process';
import { promisify } from 'node:util';

export const exec = promisify(oldExec);

export const spawnAsync = async (
  command: string,
  args: string[],
  options?: SpawnOptionsWithoutStdio,
) => {
  return new Promise<void>((resolve, reject) => {
    const child = spawn(command, args, options);

    child.on('exit', (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Command failed with code ${code}`));
      }
    });
  });
};
