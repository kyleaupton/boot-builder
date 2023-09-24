import * as api from '../electron/main/api';
import ipc from '../electron/main/api/ipc';

declare global {
  interface Window {
    api: typeof api;
    ipc: typeof ipc;
  }
}
