/**
 * It's not ideal to expose access to an entire Node API
 * via preload, however with path an attacker probably can't
 * do much and it's super handy...
 */

import path from 'path';
export { path };
