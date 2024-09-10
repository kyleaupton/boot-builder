export const showConfirmDialog = async (message: string): Promise<boolean> => {
  const { response } = await window.ipcInvoke('/dialog/showMessageBox', {
    message,
    type: 'question',
    buttons: ['Yes', 'Cancel'],
  });

  // response == 0 means "Yes"
  // response == 1 means "No"

  return response === 0;
};

export const showOpenIsoDialog = () => {
  return window.ipcInvoke('/dialog/showOpenDialog', {
    properties: ['openFile'],
    filters: [{ name: '.iso', extensions: ['.iso'] }],
  });
};

export const showOpenAppDialog = () => {
  return window.ipcInvoke('/dialog/showOpenDialog', {
    properties: ['openFile'],
    defaultPath: '/Applications',
    filters: [{ name: '.app', extensions: ['.app'] }],
  });
};
