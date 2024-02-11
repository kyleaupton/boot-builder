import {
  ipcRenderer,
  OpenDialogOptions,
  OpenDialogReturnValue,
  MessageBoxOptions,
  MessageBoxReturnValue,
} from 'electron';

export const showOpenIsoDialog = () => {
  return ipcRenderer.invoke('/dialog/showOpenDialog', {
    properties: ['openFile'],
    filters: [{ name: '.iso', extensions: ['.iso'] }],
  } as OpenDialogOptions) as Promise<OpenDialogReturnValue>;
};

export const showOpenAppDialog = () => {
  return ipcRenderer.invoke('/dialog/showOpenDialog', {
    properties: ['openFile'],
    defaultPath: '/Applications',
    filters: [{ name: '.app', extensions: ['.app'] }],
  } as OpenDialogOptions) as Promise<OpenDialogReturnValue>;
};

export const showConfirmDialog = (message: string) => {
  return ipcRenderer.invoke('/dialog/showMessageBox', {
    message,
    type: 'question',
    buttons: ['Yes', 'Cancel'],
  } as MessageBoxOptions) as Promise<MessageBoxReturnValue>;
};
