import { Disks, SystemProfiler } from '../types/disks';
import { exec } from '../utils/child_process';

export const getDisks = async (): Promise<Disks> => {
  return JSON.parse(
    (
      await exec(
        'diskutil list -plist external physical | plutil -convert json -o - -',
      )
    ).stdout,
  );
};

export const getUsbData = async (): Promise<SystemProfiler> => {
  return JSON.parse((await exec('system_profiler SPUSBDataType -json')).stdout);
};
