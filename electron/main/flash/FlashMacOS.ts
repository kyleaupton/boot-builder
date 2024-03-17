import Flash from './Flash';
import { FlashMacOSWorkerOptions } from './workers/macOS';

export default class FlashMacOS extends Flash<FlashMacOSWorkerOptions> {
  sourcePath: string;
  targetVolume: string;

  constructor({ id, sourcePath, targetVolume }: FlashMacOSWorkerOptions) {
    super({
      id,
      file: 'macOS.js',
      workerData: { id, sourcePath, targetVolume },
    });
    this.sourcePath = sourcePath;
    this.targetVolume = targetVolume;
  }
}
