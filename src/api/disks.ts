export const getDisks = () => {
  return window.ipcInvoke('/disks/get');
};
