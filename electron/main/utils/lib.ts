import path from 'path';
import os from 'os';

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

export const getPath = (
  program: t_program,
  { isPackaged = true }: { isPackaged: boolean },
) => {
  // In prod: OS Install Maker.app/Contents/Resources/app.asar/dist-electron/main
  // In dev: /Users/kyleupton/Documents/GitHub/os-install-maker/dist-electron/main
  let base: string[];
  if (isPackaged) {
    base = [__dirname, '..', '..', '..', '..'];
  } else {
    base = [__dirname, '..', '..'];
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
