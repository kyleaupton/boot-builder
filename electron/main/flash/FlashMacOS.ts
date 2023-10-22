import Flash from './Flash';
import { exists } from '../utils/fs';
import { randomBytes } from 'crypto';

export default class FlashMacOS extends Flash {
  sourcePath: string;
  targetVolume: string;

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
  }

  async run() {
    await this.validate();
    await this.eraseDrive();
    await this.executeInstallerScript();

    this._sendDone();
  }

  async validate() {
    // USB must be at least 14GB in size
    // sourcePath must have the relative `Contents/Resources/createinstallmedia`
    const scriptExists = await exists(
      `${this.sourcePath}/Contents/Resources/createinstallmedia`,
    );

    if (!scriptExists) {
      throw Error('createinstallmedia script not present');
    }
  }

  async eraseDrive() {
    // Erase USB as Mac OS Extended
    this._sendActivity(`Erasing ${this.targetVolume}`);

    try {
      // https://support.apple.com/en-us/HT201372
      // Acording to the Apple Docs, the drive needs to be pre-formatted
      // as JHFS+...? Would love to verify that sometime.
      // The createinstallmedia script will format it anyways...

      const temporaryDiskName = randomBytes(4).toString('hex');

      // Erase disk
      await this._executeCommand('diskutil', [
        'eraseDisk',
        'JHFS+',
        temporaryDiskName,
        'GPT',
        this.targetVolume,
      ]);

      // TODO: verify the USB mounted back at the same path
    } catch (e) {
      throw Error(`Failed to erase ${this.targetVolume}`);
    }
  }

  async executeInstallerScript() {
    this._sendActivity('Running createinstallmedia script');

    await this._executeCommand(
      `${this.sourcePath}/Contents/Resources/createinstallmedia`,
      ['--volume', this.targetVolume],
      {
        onOut: (data) => {
          // Parse output of script to get progress and generate eta
          console.log(data);
        },
        onErr: (error) => {
          console.log(error);
        },
      },
    );
  }
}
