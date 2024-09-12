import { SerializedFlash } from '@shared/flash';

import windows from './windows';

/* eslint-disable-next-line */
export type WorkerHandler = (...args: any[]) => Promise<any>;

export type Workers = Record<string, WorkerHandler>;

export const createWorker = <T, V>(
  /* eslint-disable-next-line */
  fun: (state: SerializedFlash, options: T) => V,
) => {
  return fun;
};

const createWorkers = <T extends Workers>(workers: T) => {
  return workers;
};

export const workers = createWorkers({
  windows,
});

export type _Workers = typeof workers;
