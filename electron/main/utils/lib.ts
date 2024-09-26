import path from 'node:path';
import os from 'node:os';
import { isDevelopment } from '@main/utils/env';

type Binary = {
  name: 'wimlib';
  binary: 'wimlib-imagex';
};

const arch = os.arch();
const currentOs = (() => {
  switch (os.platform()) {
    case 'darwin':
      return 'mac';
    case 'linux':
      return 'linux';
    case 'win32':
      return 'win';
    default:
      throw new Error('Unsupported OS');
  }
})();

const getPath = ({ name, binary }: Binary) => {
  // This is what __dirname looks like in different environments:
  // In prod: Boot Builder.app/Contents/Resources/app.asar/dist-electron/main
  // In dev: /Users/<username>/Documents/GitHub/boot-bulder/dist-electron/main
  const base = isDevelopment
    ? [__dirname, '..', '..']
    : [__dirname, '..', '..', '..', '..'];

  return path.join(...base, 'lib', currentOs, name, arch, 'bin', binary);
};

export const getWimlibPath = () => {
  return getPath({ name: 'wimlib', binary: 'wimlib-imagex' });
};
