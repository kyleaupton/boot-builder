/* eslint-disable no-unused-vars */
/* eslint-disable @typescript-eslint/no-explicit-any */

import { type SerializedFlash } from '@shared/flash';
import { type Signal } from '@main/utils/signal';

export type WorkerHandler = (...args: any[]) => Promise<any>;

export type Workers = Record<string, WorkerHandler>;

export const createWorker = <T, V>(
  fun: (s: Signal<SerializedFlash>, options: T) => V,
) => {
  return fun;
};

export const createWorkers = <T extends Workers>(workers: T) => {
  return workers;
};
