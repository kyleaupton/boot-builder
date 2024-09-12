import { createWorkers } from './utils';

import windows from './windows';

export const workers = createWorkers({
  windows,
});

export type _Workers = typeof workers;
