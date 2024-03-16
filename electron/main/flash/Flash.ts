import { app, BrowserWindow, Notification } from 'electron';
import { Worker } from 'worker_threads';
import { pathToFileURL } from 'url';
import { join } from 'path';
import { Progress } from './types';
import { removeFlash } from '../ipc/flash';

export default class Flash<WorkerData> {
  id: string;
  file: string;
  worker: Worker | undefined;
  workerData: WorkerData;

  constructor({
    id,
    file,
    workerData,
  }: {
    id: string;
    file: string;
    workerData: WorkerData;
  }) {
    this.id = id;
    this.file = file;
    this.workerData = workerData;
  }

  start() {
    this.worker = new Worker(
      join(process.env.DIST_ELECTRON, 'main', this.file),
      {
        workerData: this.workerData,
      },
    );

    this.registerEvents();
  }

  async cancel() {
    if (this.worker) {
      await this.worker.terminate();
    }

    removeFlash(this.id);
  }

  registerEvents() {
    if (!this.worker) {
      return;
    }

    this.worker.on('message', (message) => {
      if (message.type === 'progress') {
        this.sendProgress(message.data);
      } else if (message.type === 'result') {
        this.finish();
      }
    });

    this.worker.on('error', (error) => {
      console.log(error);
    });
  }

  finish() {
    this.sendProgress({
      id: this.id,
      activity: '',
      done: true,
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    });

    removeFlash(this.id);

    // Fire off notification
    new Notification({
      title: 'Flash Complete',
      body: 'Your flash is complete!',
    });
  }

  sendProgress(progress: Progress) {
    this.sendMessage(`flash-${this.id}-progress`, progress);
  }

  sendMessage(channel: string, s?: unknown) {
    BrowserWindow.getAllWindows().forEach((window) =>
      window.webContents.send(channel, s),
    );
  }
}
