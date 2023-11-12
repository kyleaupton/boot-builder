import { ipcRenderer } from 'electron';

/**
 * REMARKS
 *
 * I export this file with one default, that way I can call
 * from the renderer with `window.api.ipc.send...`
 */

const validateChannel = (channel: string) => {
  // Normal channels
  const validChannels = [
    '/usb/attached',
    '/usb/detached',
    '/flash',
    '/utils/fs/readdir',
    '/utils/macApp/getIcon',
    '/utils/app/getPath',
    '/utils/app/getFileIcon',
  ];
  if (validChannels.includes(channel)) {
    return;
  }

  // Starts with check
  const validStartsWith = ['flash-'];
  if (validStartsWith.find((x) => channel.startsWith(x))) {
    return;
  }

  throw Error('Invalid IPC channel');
};

const invoke = (channel: string, data?: unknown) => {
  //
  // Validate
  //
  validateChannel(channel);

  return ipcRenderer.invoke(channel, data);
};

const recieve = (channel: string, func: (...args: any[]) => void) => { // eslint-disable-line
  //
  // Validate
  //
  validateChannel(channel);

  ipcRenderer.on(channel, (event, ...args) => func(...args));
};

const removeListener = (channel: string, func: (...args: any[]) => void) => { // eslint-disable-line
  //
  // Validate
  //
  validateChannel(channel);

  ipcRenderer.removeListener(channel, func);
};

export default {
  invoke,
  recieve,
  removeListener,
};
