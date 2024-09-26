import { BrowserWindow } from 'electron';
import log from 'electron-log/main';
import { Worker } from 'node:worker_threads';
import { join } from 'node:path';
import { SerializedFlash } from '@shared/flash';
import { removeFlash } from '.';

import { type _Workers } from './workers/workers';

const logger = log.scope('main/Flash');

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
    logger.info('Flash constructor', { id, type, options });
    this.id = id;
    this.type = type;
    this.options = options;

    this.state = {
      id,
      done: false,
      canceled: false,
      error: undefined,
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
    logger.info('Flash start');
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
        logger.info('Flash result');
        this.state.done = true;
        this.sendState();

        // Since we're done, remove flash from main process memory
        removeFlash(this.id);
      } else if (message.type === 'error') {
        logger.info('Flash error');
        this.state.error = message.data;
        this.sendState();

        // Since we're done, remove flash from main process memory
        removeFlash(this.id);
      }
    });

    // TODO: Handle this. What even could make the worker emit an error?
    this.worker.on('error', (error) => {
      logger.error('Flash worker error');
      logger.error(error);
      console.log(error);
    });
  }

  async cancel() {
    logger.info('Flash cancel');
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
