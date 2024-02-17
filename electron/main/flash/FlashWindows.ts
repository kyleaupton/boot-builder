import { stat } from 'fs/promises';
import { basename, dirname, resolve } from 'path';
import { copy } from '@kyleupton/glob-copy';
import Flash from './Flash';
import { exec } from '../utils/child_process';
import { exists, dirSize } from '../utils/fs';
import { humanReadableToBytes } from '../utils/bytes';
import { getPath } from '../utils/lib';

export default class FlashWindows extends Flash {
  sourcePath: string;
  targetVolume: string;
  mountedIsoPath: string;
  diskName: string;

  constructor({
    id,
    sourcePath,
    targetVolume,
  }: {
    id: string;
    sourcePath: string;
    targetVolume: string;
  }) {
    super({ id });
    this.sourcePath = sourcePath;
    this.targetVolume = targetVolume;
    this.mountedIsoPath = '';
    this.diskName = '';
  }

  async run() {
    await this.mountIsoVolume();
    await this.validateIso();
    await this.eraseDrive();
    await this.copyFiles();

    this._sendDone();
  }

  async mountIsoVolume() {
    this._sendProgress({
      id: this.id,
      activity: `Mounting ${this.sourcePath}`,
      done: false,
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    });

    try {
      const { stdout } = await exec(`hdiutil mount ${this.sourcePath}`);
      this.mountedIsoPath = stdout.trim().split(/\s+/)[1];
    } catch (e) {
      throw Error(`Failed to mount ${this.sourcePath}`);
    }
  }

  async validateIso() {
    const needed = ['sources/install.wim'];

    for (const rel of needed) {
      const abs = resolve(this.mountedIsoPath, rel);

      if (!(await exists(abs))) {
        throw Error(`Invalid Windows ISO file - ${rel} not found`);
      }
    }
  }

  async eraseDrive() {
    this._sendProgress({
      id: this.id,
      activity: `Erasing ${this.targetVolume}`,
      done: false,
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    });

    try {
      // Get disk name
      // Windows 10: `Win10_22H2_English_x64v1.iso`
      // Windows 11: `Win11_22H2_English_x64v2.iso`
      if (this.diskName.includes('Win10') || this.diskName.includes('Win11')) {
        this.diskName = this.sourcePath
          .split('/')
          .slice(-1)[0]
          .split('_')[0]
          .toUpperCase();
      } else {
        this.diskName = 'WINDOWS';
      }

      // Erase disk
      await this._executeCommand('diskutil', [
        'eraseDisk',
        'MS-DOS',
        this.diskName,
        'MBR',
        this.targetVolume,
      ]);
    } catch (e) {
      throw Error(`Failed to erase ${this.targetVolume}`);
    }
  }

  async copyFiles() {
    const activity = 'Copying files';

    this._sendProgress({
      id: this.id,
      activity,
      done: false,
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    });

    try {
      // Get stats for progress reporting
      const totalSize = await dirSize(this.mountedIsoPath);
      const bigFileSize = await stat(
        `${this.mountedIsoPath}/sources/install.wim`,
      );
      let transferred = 0;

      // Copy over everything minus the big file
      await copy({
        source: this.mountedIsoPath,
        destination: `/Volumes/${this.diskName}`,
        options: {
          ignore: ['sources/install.wim'],
        },
        onProgress: (progress) => {
          transferred = progress.transferred;
          const remaining = totalSize - progress.transferred;

          this._sendProgress({
            id: this.id,
            activity,
            done: false,
            transferred: progress.transferred,
            speed: progress.speed,
            percentage: (progress.transferred / totalSize) * 100,
            eta: remaining / progress.speed,
          });
        },
      });

      // Next copy the big file
      const wimlibImagex = getPath({ name: 'wimlib', bin: 'wimlib-imagex' });
      const dir = dirname(wimlibImagex);
      const name = basename(wimlibImagex);
      const start = Date.now();
      let secondFileTransferred = 0;

      let progress = {
        id: this.id,
        activity,
        done: false,
        transferred: 0,
        speed: 0,
        percentage: 0,
        eta: 0,
      };

      const id = setInterval(() => {
        this._sendProgress(progress);
      }, 1000);

      await this._executeCommand(
        name,
        [
          'split',
          `${this.mountedIsoPath}/sources/install.wim`,
          `/Volumes/${this.diskName}/sources/install.swm`,
          '3800',
        ],
        {
          onOut: (data) => {
            const match = data.match(
              /(?<part>\d{1,7}\s\w{2,4}) of (?<whole>\d{1,7}\s\w{2,4})/,
            );

            if (match && match.groups) {
              const { part } = match.groups;
              const { bytes } = humanReadableToBytes(part);
              const delta = bytes - secondFileTransferred;
              secondFileTransferred = bytes;
              transferred += delta;
              const remaining = bigFileSize.size - bytes;
              const speed = bytes / ((Date.now() - start) / 1000);

              progress = {
                id: this.id,
                activity,
                done: false,
                transferred,
                speed,
                percentage: (transferred / totalSize) * 100,
                eta: remaining / speed,
              };
            }
          },
        },
        {
          cwd: dir,
        },
      );

      clearInterval(id);
    } catch (e) {
      console.log(e);
      throw Error('Failed to copy files');
    }
  }

  async cleanUp() {
    this._sendProgress({
      id: this.id,
      activity: 'Cleaning up',
      done: false,
      transferred: -1,
      speed: -1,
      percentage: -1,
      eta: -1,
    });

    try {
      await this._executeCommand('diskutil', ['eject', this.targetVolume]);
      await this._executeCommand('umount', [this.mountedIsoPath]);
    } catch (e) {
      throw Error('Failed to clean up');
    }
  }
}
