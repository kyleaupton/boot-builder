import { parentPort, workerData } from 'worker_threads';
import { workers } from './workers';
import { SerializedFlash } from '@shared/flash';
import { signal } from '@main/utils/signal';

if (parentPort && workerData) {
  try {
    if (
      typeof workerData !== 'object' ||
      !['type', 'options', 'state'].every((key) => key in workerData)
    ) {
      throw new Error('Invalid worker data');
    }

    // Flash job type
    const type = workerData.type as keyof typeof workers;
    // Flash job options
    const options = workerData.options as Parameters<
      (typeof workers)[typeof type]
    >[1];

    // Function to run the flash job
    const func = workers[type];
    if (!func) {
      throw new Error(`Worker type ${type} not found`);
    }

    // State signal for the flash job,
    // this is used for UI updates.
    // There are two different things
    // we can do here: we can either send
    // updates to the frontend every 0.5s
    // or we can send updates immediately
    // when the state changes. The former
    // is more efficient and looks better,
    // but the latter is cooler.
    const state = signal({
      value: workerData.state as SerializedFlash,
    });

    const id = setInterval(() => {
      parentPort!.postMessage({
        type: 'state',
        data: state.get(),
      });
    }, 500);

    func(state, options)
      .then((res) => {
        parentPort!.postMessage({
          type: 'result',
          data: res,
        });
      })
      .catch((error) => {
        throw error;
      })
      .finally(() => {
        clearInterval(id);
      });
  } catch (error) {
    console.error(error);
    parentPort.postMessage({
      type: 'error',
      data: error,
    });
  }
}
