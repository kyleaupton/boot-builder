import { ipcRenderer, OpenDialogReturnValue } from 'electron';

export const showOpenIsoDialog = () => {
  return ipcRenderer.invoke('/dialog/showOpenDialog', {
    properties: ['openFile'],
    filters: [{ name: '.iso', extensions: ['.iso'] }],
  }) as Promise<OpenDialogReturnValue>;
};
