import { BrowserWindow } from 'electron';
import { Worker } from 'node:worker_threads';
import { join } from 'node:path';
import { SerializedFlash } from '@shared/flash';
import { removeFlash } from '.';

import { type _Workers } from './workers/workers';

export default class Flash<
  T extends keyof _Workers,
  V extends Parameters<_Workers[T]>[1],
> {
  id: string;
  type: T;
  options: V;
  state: SerializedFlash;
  worker: Worker | undefined;

  constructor(id: string, type: T, options: NoInfer<V>) {
    this.id = id;
    this.type = type;
    this.options = options;

    this.state = {
      id,
      status: '',
      done: false,
      canceled: false,
      progress: {
        activity: '',
        transferred: -1,
        speed: -1,
        percentage: -1,
        eta: -1,
      },
    };
  }

  start() {
    this.worker = new Worker(
      join(process.env.DIST_ELECTRON, 'main', 'entry.js'),
      {
        workerData: {
          type: this.type,
          state: this.state,
          options: this.options,
        },
      },
    );

    this.worker.on('message', (message) => {
      if (message.type === 'state') {
        this.state = message.data;
        this.sendState();
      } else if (message.type === 'result') {
        this.state.done = true;
        this.sendState();

        // Since we're done, remove flash from main process memory
        removeFlash(this.id);
      } else if (message.type === 'error') {
        this.state.status = message.data;
        this.sendState();

        // Since we're done, remove flash from main process memory
        removeFlash(this.id);
      }
    });

    this.worker.on('error', (error) => {
      console.log(error);
    });
  }

  async cancel() {
    if (this.worker) {
      await this.worker.terminate();
      this.state.canceled = true;
      this.sendState();
    }
  }

  sendState() {
    BrowserWindow.getAllWindows().forEach((window) => {
      window.webContents.send('/flash/update', this.state);
    });
  }
}
