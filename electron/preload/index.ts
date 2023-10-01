import { contextBridge } from 'electron';
import * as api from '../main/api';

contextBridge.exposeInMainWorld('api', api);
