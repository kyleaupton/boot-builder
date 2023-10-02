import { reactive } from 'vue';
import { nanoid } from 'nanoid';
import { t_drive } from '@/types/disks';
import { t_file } from '@/types/iso';
import { t_flashing_progress } from '@/types/flash';

const testing = false;

export default class Drive {
  meta: t_drive;
  id: string;

  // State
  isoFile: undefined | t_file;
  flashing: boolean;
  doneFlashing: boolean;
  flashingProgress: t_flashing_progress;

  constructor(meta: t_drive) {
    this.meta = meta;
    this.id = nanoid();
    this.flashing = false;
    this.doneFlashing = false;
    this.flashingProgress = reactive({
      currentActivity: '',
      stdout: '',
      stderr: '',
    });
  }

  async startFlash() {
    if (!this.isoFile) {
      throw Error('Must define isoPath first');
    }

    this.flashing = true;

    if (testing) {
      await new Promise((resolve) => setTimeout(resolve, 5000));
    } else {
      await window.api.ipc.invoke('/flash', {
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

  _registerIpcEvents() {
    window.api.ipc.recieve(`flash-${this.id}-activity`, this._activityHandler);
    window.api.ipc.recieve(`flash-${this.id}-stdout`, this._stdoutHandler);
    window.api.ipc.recieve(`flash-${this.id}-stderr`, this._stderrHandler);
  }

  _removeIpcEvents() {
    window.api.ipc.removeListener(
      `flash-${this.id}-activity`,
      this._activityHandler,
    );
    window.api.ipc.removeListener(
      `flash-${this.id}-stdout`,
      this._stdoutHandler,
    );
    window.api.ipc.removeListener(
      `flash-${this.id}-stdout`,
      this._stdoutHandler,
    );
  }

  _activityHandler(activity: string) {
    this.flashingProgress.currentActivity = activity;
  }

  _stdoutHandler(stdout: string) {
    this.flashingProgress.stdout += stdout;
  }

  _stderrHandler(stderr: string) {
    this.flashingProgress.stderr += stderr;
  }
}
