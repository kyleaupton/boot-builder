import { ipcMain, IpcMainInvokeEvent } from 'electron';
import { getPath } from '../utils/lib';

export default function start() {
  ipcMain.handle(
    '/utils/lib/getPath',
    (event: IpcMainInvokeEvent, program: 'wimlib') => {
      return getPath(program);
    },
  );
}
