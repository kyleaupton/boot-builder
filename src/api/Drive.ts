import { t_drive } from '../../electron/main/api/disks';

export default class Drive {
  meta: t_drive;

  // State
  flashing: boolean;

  constructor(meta: t_drive) {
    this.meta = meta;
    this.flashing = false;
  }

  async startFlash() {
    this.flashing = true;

    // await window.api.create({
    //   isoFile: '/Users/kyleupton/Downloads/Win10_22H2_English_x64v1.iso',
    //   volume: `/dev/${this.meta.DeviceIdentifier}`,
    // });

    // this.flashing = false;
  }
}
