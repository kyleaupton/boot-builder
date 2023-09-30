import { app } from 'electron';
import path from 'path';
import os from 'os';

type t_program = 'wimlib';

const { isPackaged } = app;
const arch = os.arch();

export const getPath = (program: t_program) => {
  // In prod: /Users/kyleupton/Documents/GitHub/windows-install-maker/release/0.0.1/mac-arm64/OS Install Maker.app/Contents/Resources/app.asar/dist-electron/main
  // In dev: /Users/kyleupton/Documents/GitHub/windows-install-maker/dist-electron/main
  let base: string[];
  if (isPackaged) {
    base = [__dirname, '..', '..', '..', '..'];
  } else {
    base = [__dirname, '..', '..'];
  }

  switch (program) {
    case 'wimlib':
      return path.join(
        ...base,
        'electron',
        'lib',
        program,
        arch,
        'bin',
        'wimlib-imagex',
      );
    default:
      throw Error('Invalid program');
  }
};
