import { onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'

/**
 * Type-safe Wails event subscription with automatic cleanup on unmount.
 *
 * @example
 * useWailsEvents<JobEvent>('job:event', (data) => {
 *   console.log(data.Percent)
 * })
 */
export function useWailsEvents<T = unknown>(
  eventName: string,
  callback: (data: T) => void
): () => void {
  const handler = (event: Events.WailsEvent) => {
    callback(event.data as T)
  }

  const unsubscribe = Events.On(eventName, handler)

  onUnmounted(() => {
    unsubscribe()
  })

  return unsubscribe
}

/**
 * Subscribe to an event once, with automatic cleanup if component unmounts before event fires.
 */
export function useWailsEventOnce<T = unknown>(
  eventName: string,
  callback: (data: T) => void
): () => void {
  const handler = (event: Events.WailsEvent) => {
    callback(event.data as T)
  }

  const unsubscribe = Events.Once(eventName, handler)

  onUnmounted(() => {
    unsubscribe()
  })

  return unsubscribe
}

/**
 * Subscribe to events outside of Vue component lifecycle (e.g., in Pinia stores).
 * Returns unsubscribe function - caller is responsible for cleanup.
 */
export function subscribeToWailsEvent<T = unknown>(
  eventName: string,
  callback: (data: T) => void
): () => void {
  const handler = (event: Events.WailsEvent) => {
    callback(event.data as T)
  }

  return Events.On(eventName, handler)
}

// Common event name constants
export const WailsEventNames = {
  JOB_EVENT: 'job:event',
  FILES_DROPPED: 'common:WindowFilesDropped',
} as const
