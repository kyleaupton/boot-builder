import { spawn, SpawnOptions } from 'child_process';
import { parentPort, workerData } from 'worker_threads';
import { Progress } from '@main/flash/types';

export const expose = async <T, V>({
  fn,
}: {
  fn: (param: T) => Promise<V>; // eslint-disable-line
}) => {
  if (parentPort) {
    const result = await fn(workerData);

    parentPort?.postMessage({
      type: 'result',
      data: result,
    });
  }
};

export const sendProgress = (p: Progress) => {
  if (parentPort) {
    parentPort.postMessage({
      type: 'progress',
      data: p,
    });
  }
};

export const executeCommand = async (
  cmd: string,
  args: string[],
  options: SpawnOptions,
  events?: { onOut?: (data: string) => void, onErr?: (data: string) => void }, // eslint-disable-line
) => {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn(cmd, args);

    if (proc.stdout) {
      proc.stdout.on('data', (data) => {
        const string = data.toString();
        // this._sendCommandOutput(string, 'stdout');
        if (events && events.onOut) events.onOut(string);
      });

      proc.stderr.on('data', (data) => {
        const string = data.toString();
        console.log(string);
        // this._sendCommandOutput(string, 'stderr');
        if (events && events.onErr) events.onErr(string);
      });
    }

    proc.on('close', () => {
      resolve();
    });

    proc.on('error', (error) => {
      console.log(error);
      reject(error);
    });
  });
};
