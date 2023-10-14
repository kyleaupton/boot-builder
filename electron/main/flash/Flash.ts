import { spawn, SpawnOptions } from 'child_process';
import { BrowserWindow } from 'electron';
import { removeFlash } from '../ipc/flash';

interface IProgress {
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

  _sendActivity(s: string) {
    this._sendMessage(`flash-${this.id}-activity`, s);
  }

  _sendEta(payload: IProgress) {
    this._sendMessage(`flash-${this.id}-copy-eta`, payload);
  }

  _sendCommandOutput(s: string, type: 'stdout' | 'stderr') {
    this._sendMessage(`flash-${this.id}-${type}`, s);
  }

  _sendDone() {
    this._sendMessage(`flash-${this.id}-done`);
    removeFlash(this.id);
  }
}
