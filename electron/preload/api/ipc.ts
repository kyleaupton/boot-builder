import { ipcRenderer } from 'electron';

/**
 * REMARKS
 *
 * I export this file with one default, that way I can call
 * from the renderer with `window.api.ipc.send...`
 */

const send = (channel: string, data?: unknown) => {
  const validChannels = [];

  if (validChannels.includes(channel)) {
    ipcRenderer.send(channel, data);
  }
};

const recieve = (channel: string, func: (...args: any[]) => void) => { // eslint-disable-line
  const validChannels = ['/usb/attached', '/usb/detached'];

  if (validChannels.includes(channel)) {
    ipcRenderer.on(channel, (event, ...args) => func(...args));
  }
};

export default {
  send,
  recieve,
};
