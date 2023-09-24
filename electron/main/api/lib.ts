// import url from 'node:url';
import path from 'node:path';

export const getLibPath = () => {
  const isDev = !!process.env.VITE_DEV_SERVER_URL;

  if (isDev) {
    return path.resolve(__dirname, '..', '..', 'electron', 'lib');
  } else {
    return '';
  }
};
