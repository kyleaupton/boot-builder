import { ipcRenderer } from 'electron';

export const send = (channel: string, data?: unknown) => {
  const validChannels = [];

  if (validChannels.includes(channel)) {
    ipcRenderer.send(channel, data);
  }
};

export const recieve = (channel: string, func: (...args: any[]) => void) => { // eslint-disable-line
  const validChannels = ['/usb/attached', '/usb/detached'];

  if (validChannels.includes(channel)) {
    ipcRenderer.on(channel, (event, ...args) => func(...args));
  }
};

export default { send, recieve };
