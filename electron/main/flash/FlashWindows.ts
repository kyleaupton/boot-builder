import { app } from 'electron';
import Flash from './Flash';
import {
  FlashWindowsWorkerOptions,
  FlashWindowsWorkerOptionsFinal,
} from './workers/windows';

export default class FlashWindows extends Flash<FlashWindowsWorkerOptionsFinal> {
  constructor({ id, sourcePath, targetVolume }: FlashWindowsWorkerOptions) {
    super({
      id,
      file: 'windows.js', // Path of built worker file
      workerData: { id, sourcePath, targetVolume, isPackaged: app.isPackaged },
    });
  }
}
