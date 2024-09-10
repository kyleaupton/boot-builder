import { ipcRouter } from 'typed-electron-ipc';
import { dialogIpc } from './dialog';
import { disksIpc } from './disks';
import { flashIpc } from './flash';
import { pathIpc } from './path';
import { utilsIpc } from './utils';

const router = ipcRouter({
  ...dialogIpc(),
  ...disksIpc(),
  ...flashIpc(),
  ...pathIpc(),
  ...utilsIpc(),
});

export type Router = typeof router;
