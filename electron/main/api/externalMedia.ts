import { exec } from '../utils/child_process';

export const getExternalDevices = async () => {
  const { stdout } = await exec('diskutil list');
  console.log(stdout);

  // const blocks = stdout.split();
};
