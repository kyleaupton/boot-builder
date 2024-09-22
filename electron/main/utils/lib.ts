import path from 'path';
import os from 'os';
import { isDevelopment } from '@main/utils/env';

type t_program =
  | {
      name: 'wimlib';
      bin: 'wimlib-imagex';
    }
  | {
      name: 'rsync';
      bin: 'rsync';
    };

const arch = os.arch();

export const getPath = (program: t_program) => {
  // In prod: Boot Builder.app/Contents/Resources/app.asar/dist-electron/main
  // In dev: /Users/<username>/Documents/GitHub/boot-bulder/dist-electron/main
  let base: string[];
  if (isDevelopment) {
    base = [__dirname, '..', '..'];
  } else {
    base = [__dirname, '..', '..', '..', '..'];
  }

  return path.join(
    ...base,
    'electron',
    'lib',
    program.name,
    arch,
    'bin',
    program.bin,
  );
};
