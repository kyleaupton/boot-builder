import { app } from 'electron';
import path from 'path';
import os from 'os';

type t_program = 'wimlib';

const { isPackaged } = app;
const arch = os.arch();

export const getPath = (program: t_program) => {
  switch (program) {
    case 'wimlib':
      console.log(__dirname);

      if (isPackaged) {
        return '';
      }

      return path.join(
        __dirname,
        '..',
        '..',
        'electron',
        'lib',
        arch,
        program,
        'bin',
        'wimlib-imagex',
      );
    default:
      throw Error('Invalid program');
  }
};
