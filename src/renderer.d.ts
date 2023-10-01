import * as api from '../electron/preload/api';

declare global {
  interface Window {
    api: typeof api;
  }
}
