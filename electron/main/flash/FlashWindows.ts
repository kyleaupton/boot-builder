import Flash from './Flash';
import { FlashWindowsWorkerOptions } from './workers/windows';

export default class FlashWindows extends Flash<FlashWindowsWorkerOptions> {
  constructor({ id, sourcePath, targetVolume }: FlashWindowsWorkerOptions) {
    super({
      id,
      file: 'windows.js',
      workerData: { id, sourcePath, targetVolume },
    });
  }
}
