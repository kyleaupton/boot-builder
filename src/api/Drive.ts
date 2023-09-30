import { t_drive } from '@/types/disks';
import { t_file } from '@/types/iso';

const testing = false;

export default class Drive {
  meta: t_drive;

  // State
  isoFile: undefined | t_file;
  flashing: boolean;
  doneFlashing: boolean;

  constructor(meta: t_drive) {
    this.meta = meta;
    this.flashing = false;
    this.doneFlashing = false;
  }

  async startFlash() {
    if (!this.isoFile) {
      throw Error('Must define isoPath first');
    }

    this.flashing = true;

    if (testing) {
      await new Promise((resolve) => setTimeout(resolve, 5000));
    } else {
      await window.api.create({
        isoFile: this.isoFile.path,
        volume: `/dev/${this.meta.DeviceIdentifier}`,
      });
    }

    this.flashing = false;
    this.doneFlashing = true;
  }

  /**
   * Since we cannot modify a instance variable on
   * a vue component prop, we need this to allow
   * for modification of the file in such a case.
   */
  setIsoFile(file: t_file | undefined) {
    this.isoFile = file;
  }
}
