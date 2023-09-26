import { t_drive } from '../../electron/main/api/disks';

export default class Drive {
  meta: t_drive;

  constructor(meta: t_drive) {
    this.meta = meta;
  }
}
