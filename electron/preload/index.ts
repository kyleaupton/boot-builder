import { ipcRenderer, contextBridge } from 'electron';
import { createIpcClient } from 'typed-electron-ipc';
import { Router } from '@main/ipc';
import { SerializedFlash } from '@shared/flash';

export const ipcInvoke = createIpcClient<Router>();

export const onUsbAttached = (callback: () => void): void => {
  ipcRenderer.on('/usb/attached', () => callback());
};

export const onUsbDetached = (callback: () => void): void => {
  ipcRenderer.on('/usb/detached', () => callback());
};

export const onFlashUpdate = (
  callback: (
    /* eslint-disable-next-line no-unused-vars */
    flash: SerializedFlash,
  ) => void,
): void => {
  ipcRenderer.on('/flash/update', (_event, flash) => callback(flash));
};

contextBridge.exposeInMainWorld('ipcInvoke', ipcInvoke);
contextBridge.exposeInMainWorld('onUsbAttached', onUsbAttached);
contextBridge.exposeInMainWorld('onUsbDetached', onUsbDetached);
contextBridge.exposeInMainWorld('onFlashUpdate', onFlashUpdate);
