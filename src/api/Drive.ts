import { reactive } from 'vue';
import { t_drive } from '@/types/disks';
import { t_file } from '@/types/iso';
import { t_flashing_progress } from '@/types/flash';

const testing = false;

export default class Drive {
  meta: t_drive;
  id: string;

  // State
  os: 'Windows' | 'macOS' | 'Linux' | undefined;
  sourcePath: t_file | undefined;
  flashing: boolean;
  doneFlashing: boolean;
  flashingProgress: t_flashing_progress;

  constructor(meta: t_drive) {
    this.meta = meta;
    this.id = meta.serial_num;
    this.flashing = false;
    this.doneFlashing = false;
    this.flashingProgress = reactive({
      currentActivity: '',
      stdout: '',
      stderr: '',
      error: '',
    });

    this._stdoutHandler = this._stdoutHandler.bind(this);
    this._stderrHandler = this._stderrHandler.bind(this);
    this._activityHandler = this._activityHandler.bind(this);
    this._etaHandler = this._etaHandler.bind(this);
    this._doneHandler = this._doneHandler.bind(this);

    this._registerIpcEvents();
  }

  async startFlash() {
    if (!this.sourcePath || !this.os) {
      throw Error('Must define isoPath & os first');
    }

    this.flashing = true;

    if (testing) {
      await new Promise((resolve) => setTimeout(resolve, 5000));
    } else {
      try {
        let url: string;

        if (this.os === 'Windows') {
          url = '/flash/windows';
        } else if (this.os === 'macOS') {
          url = '/flash/macOS';
        } else if (this.os === 'Linux') {
          url = '/flash/linux';
        } else {
          throw Error('Invalid os chosen');
        }

        await window.api.ipc.invoke(url, {
          sourcePath: this.sourcePath.path,
          volume: `/dev/${this.meta.DeviceIdentifier}`,
          id: this.id,
        });
      } catch (e) {
        this.flashingProgress.error =
          e instanceof Error ? e.message : String(e);
      }
    }
  }

  /**
   * Since we cannot modify a instance variable on
   * a vue component prop, we need this to allow
   * for modification of the file in such a case.
   */
  setOs(os: typeof this.os) {
    this.os = os;
  }

  setSourcePath(file: t_file | undefined) {
    this.sourcePath = file;
  }

  _registerIpcEvents() {
    window.api.ipc.recieve(`flash-${this.id}-activity`, this._activityHandler);
    window.api.ipc.recieve(`flash-${this.id}-stdout`, this._stdoutHandler);
    window.api.ipc.recieve(`flash-${this.id}-stderr`, this._stderrHandler);
    window.api.ipc.recieve(`flash-${this.id}-copy-eta`, this._etaHandler);
    window.api.ipc.recieve(`flash-${this.id}-done`, this._doneHandler);
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

    window.api.ipc.removeListener(
      `flash-${this.id}-copy-eta`,
      this._etaHandler,
    );

    window.api.ipc.removeListener(`flash-${this.id}-done`, this._doneHandler);
  }

  _activityHandler(activity: string) {
    if (!this.flashing) {
      this.flashing = true;
    }

    this.flashingProgress.currentActivity = activity;
  }

  _stdoutHandler(stdout: string) {
    if (!this.flashing) {
      this.flashing = true;
    }

    this.flashingProgress.stdout += stdout;
  }

  _stderrHandler(stderr: string) {
    if (!this.flashing) {
      this.flashing = true;
    }

    this.flashingProgress.stderr += stderr;
  }

  _etaHandler(progress: {
    transferred: number;
    speed: number;
    percentage: number;
    eta: number;
  }) {
    if (!this.flashing) {
      this.flashing = true;
    }

    const time = progress.eta;

    const h = Math.floor(time / 3600)
      .toString()
      .padStart(2, '0');

    const m = Math.floor((time % 3600) / 60)
      .toString()
      .padStart(2, '0');

    const s = Math.floor(time % 60)
      .toString()
      .padStart(2, '0');

    const eta = h !== '00' ? `${h}:${m}:${s}` : `${m}:${s}`;

    this.flashingProgress.copy = {
      ...progress,
      etaHuman: eta,
    };
  }

  _doneHandler() {
    this.flashing = false;
    this.doneFlashing = true;
    this._removeIpcEvents();
  }
}
