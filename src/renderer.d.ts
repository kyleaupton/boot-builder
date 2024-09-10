import {
  ipcInvoke,
  onUsbAttached,
  onUsbDetached,
  onFlashUpdate,
} from '@preload/index';

declare global {
  interface Window {
    ipcInvoke: typeof ipcInvoke;
    onUsbAttached: typeof onUsbAttached;
    onUsbDetached: typeof onUsbDetached;
    onFlashUpdate: typeof onFlashUpdate;
  }
}
