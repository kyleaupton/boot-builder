import { t_drive } from '../../electron/main/api/disks';

export default class Drive {
  meta: t_drive;

  // State
  flashing: boolean;
  doneFlashing: boolean;

  constructor(meta: t_drive) {
    this.meta = meta;
    this.flashing = false;
    this.doneFlashing = false;
  }

  async startFlash() {
    this.flashing = true;

    // await window.api.create({
    //   isoFile: '/Users/kyleupton/Downloads/Win10_22H2_English_x64v1.iso',
    //   volume: `/dev/${this.meta.DeviceIdentifier}`,
    // });

    // this.flashing = false;

    await new Promise((resolve) => setTimeout(resolve, 5000));

    this.flashing = false;
    this.doneFlashing = true;
  }
}
