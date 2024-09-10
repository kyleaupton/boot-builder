import { createIpcHandlers } from 'typed-electron-ipc';
import drives from 'drivelist';

export const disksIpc = () =>
  createIpcHandlers({
    '/disks/get': async () => {
      return drives.list();
    },
  });
