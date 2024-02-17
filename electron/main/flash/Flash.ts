import { spawn, SpawnOptions } from 'child_process';
import { BrowserWindow, Notification } from 'electron';
import { removeFlash } from '../ipc/flash';

interface IProgress {
  id: string;
  activity: string;
  done: boolean;
  // ETA
  transferred: number;
  speed: number;
  percentage: number;
  eta: number;
}

export default class Flash {
  id: string;

  constructor({ id }: { id: string }) {
    this.id = id;
  }

  async _executeCommand(
    cmd: string,
    args?: string[],
    events?: { onOut?: (data: string) => void, onErr?: (data: string) => void }, // eslint-disable-line
    options?: SpawnOptions,
  ) {
    return new Promise<void>((resolve, reject) => {
      const proc = spawn(cmd, args, options);

      proc.stdout.on('data', (data) => {
        const string = data.toString();
        this._sendCommandOutput(string, 'stdout');
        if (events && events.onOut) events.onOut(string);
      });

      proc.stderr.on('data', (data) => {
        const string = data.toString();
        console.log(string);
        this._sendCommandOutput(string, 'stderr');
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
  }

  _sendMessage(channel: string, s?: unknown) {
    BrowserWindow.getAllWindows().forEach((window) =>
      window.webContents.send(channel, s),
    );
  }

  _sendProgress(progress: IProgress) {
    this._sendMessage(`flash-${this.id}-progress`, progress);
  }

  _sendCommandOutput(s: string, type: 'stdout' | 'stderr') {
    this._sendMessage(`flash-${this.id}-${type}`, s);
  }

  _sendDone() {
    this._sendProgress({
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
}
