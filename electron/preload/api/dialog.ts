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

export const showConfirmDialog = async (message: string): Promise<boolean> => {
  const { response } = (await ipcRenderer.invoke('/dialog/showMessageBox', {
    message,
    type: 'question',
    buttons: ['Yes', 'Cancel'],
  } as MessageBoxOptions)) as MessageBoxReturnValue;

  // response == 0 means "Yes"
  // response == 1 means "No"

  return response === 0;
};
