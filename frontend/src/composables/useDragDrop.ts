import { ref, onMounted, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'
import { WailsEventNames } from './useWailsEvents'

export interface FileDropPayload {
  files: string[]
}

export interface UseDragDropOptions {
  /** Callback when files are dropped */
  onDrop?: (files: string[]) => void
  /** File extensions to accept, e.g., ['.iso', '.img'] */
  accept?: string[]
}

/**
 * Composable for handling Wails file drop events.
 *
 * In your template, add `data-file-drop-target` attribute to the drop zone element.
 *
 * @example
 * const { droppedFiles, clearDroppedFiles } = useDragDrop({
 *   onDrop: (files) => console.log('Dropped:', files),
 *   accept: ['.iso']
 * })
 */
export function useDragDrop(options?: UseDragDropOptions) {
  const droppedFiles = ref<string[]>([])
  const lastDroppedFile = ref<string | null>(null)

  let unsubscribe: (() => void) | null = null

  const filterFiles = (files: string[]): string[] => {
    if (!options?.accept || options.accept.length === 0) {
      return files
    }
    return files.filter((file) =>
      options.accept!.some((ext) => file.toLowerCase().endsWith(ext.toLowerCase()))
    )
  }

  const handleFileDrop = (event: Events.WailsEvent) => {
    const payload = event.data as FileDropPayload
    const filtered = filterFiles(payload.files)

    if (filtered.length > 0) {
      droppedFiles.value = filtered
      lastDroppedFile.value = filtered[0]
      options?.onDrop?.(filtered)
    }
  }

  const clearDroppedFiles = () => {
    droppedFiles.value = []
    lastDroppedFile.value = null
  }

  onMounted(() => {
    unsubscribe = Events.On(WailsEventNames.FILES_DROPPED, handleFileDrop)
  })

  onUnmounted(() => {
    unsubscribe?.()
  })

  return {
    /** All files from the last drop (filtered by accept if provided) */
    droppedFiles,
    /** First file from the last drop (convenience for single-file use cases) */
    lastDroppedFile,
    /** Clear the dropped files state */
    clearDroppedFiles,
  }
}
