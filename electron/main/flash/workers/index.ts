import { spawn, SpawnOptions } from 'child_process';
import { parentPort, workerData } from 'worker_threads';
import { Progress } from '../types';

export const expose = async ({
  fn,
}: {
  fn: (...args: Array<unknown>) => Promise<unknown>; // eslint-disable-line
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
  args?: string[],
  events?: { onOut?: (data: string) => void, onErr?: (data: string) => void }, // eslint-disable-line
  options?: SpawnOptions,
) => {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn(cmd, args, options);

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

    proc.on('close', () => {
      resolve();
    });

    proc.on('error', (error) => {
      console.log(error);
      reject(error);
    });
  });
};
