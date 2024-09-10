import { parentPort } from 'node:process';

export default class WorkerClass {
  constructor() {}

  async flash() {}

  async start() {
    if (parentPort) {
      const result = await this.flash();

      parentPort.postMessage({
        type: 'result',
        data: result,
      });
    }
  }
}
