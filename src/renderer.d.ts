import * as api from '../electron/main/api';

declare global {
  interface Window {
    api: typeof api;
  }
}
