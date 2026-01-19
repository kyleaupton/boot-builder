import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { Events, WailsEvent } from '@wailsio/runtime'
import { StartJob, ListJobs } from '@bindings/boot-builder/internal/service/jobsservice'
import { Status } from '@bindings/boot-builder/internal/jobs/models'
import type { Job, JobEvent, StartJobRequest } from '@/types'
import { WailsEventNames } from '@/composables'

export const useJobStore = defineStore('job', () => {
  // State
  const currentJobId = ref<string | null>(null)
  const jobs = ref<Map<string, Job>>(new Map())
  const progress = ref(0)
  const currentStep = ref<string | null>(null)
  const currentMessage = ref<string | null>(null)
  const error = ref<string | null>(null)
  const isStarting = ref(false)

  // Event subscription handle
  let eventUnsubscribe: (() => void) | null = null

  // Getters
  const currentJob = computed(() =>
    currentJobId.value ? jobs.value.get(currentJobId.value) ?? null : null
  )

  const status = computed((): Status | null => currentJob.value?.Status ?? null)

  const isRunning = computed(() => status.value === Status.StatusRunning)
  const isComplete = computed(() => status.value === Status.StatusSucceeded)
  const isFailed = computed(() => status.value === Status.StatusFailed)
  const isPending = computed(() => status.value === Status.StatusPending)
  const isIdle = computed(() => !currentJobId.value || (!isRunning.value && !isPending.value))

  // Actions
  function handleJobEvent(event: JobEvent): void {
    // Only process events for our current job
    if (event.JobID !== currentJobId.value) return

    switch (event.Type) {
      case 'state': {
        const job = jobs.value.get(event.JobID)
        if (job) {
          // Map state message to Status enum
          const stateMap: Record<string, Status> = {
            pending: Status.StatusPending,
            running: Status.StatusRunning,
            succeeded: Status.StatusSucceeded,
            failed: Status.StatusFailed,
          }
          job.Status = stateMap[event.Message] ?? job.Status
        }
        break
      }

      case 'step-start':
        currentStep.value = event.Step
        currentMessage.value = event.Message
        break

      case 'step-end':
        // Could track completed steps if needed
        break

      case 'progress':
        progress.value = event.Percent
        if (event.Message) {
          currentMessage.value = event.Message
        }
        break

      case 'log':
        currentMessage.value = event.Message
        break

      case 'error':
        error.value = event.Error || event.Message
        break
    }
  }

  function subscribeToEvents(): void {
    if (eventUnsubscribe) return

    eventUnsubscribe = Events.On(WailsEventNames.JOB_EVENT, (ev: WailsEvent) => {
      handleJobEvent(ev.data as JobEvent)
    })
  }

  function unsubscribeFromEvents(): void {
    eventUnsubscribe?.()
    eventUnsubscribe = null
  }

  async function startJob(request: StartJobRequest): Promise<string> {
    isStarting.value = true
    error.value = null
    progress.value = 0
    currentStep.value = null
    currentMessage.value = null

    try {
      // Ensure we're subscribed to events before starting
      subscribeToEvents()

      const jobId = await StartJob(request)
      currentJobId.value = jobId

      // Create a placeholder job entry
      jobs.value.set(jobId, {
        ID: jobId,
        Plan: null,
        Status: Status.StatusPending,
        Progress: 0,
        CreatedAt: null as any,
        UpdatedAt: null as any,
      })

      return jobId
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to start job'
      throw e
    } finally {
      isStarting.value = false
    }
  }

  async function refreshJobs(): Promise<void> {
    try {
      const jobList = await ListJobs()
      jobList.forEach((job) => {
        jobs.value.set(job.ID, job)
      })
    } catch (e) {
      console.error('Failed to refresh jobs:', e)
    }
  }

  function clearCurrentJob(): void {
    currentJobId.value = null
    progress.value = 0
    currentStep.value = null
    currentMessage.value = null
    error.value = null
  }

  function $reset(): void {
    unsubscribeFromEvents()
    currentJobId.value = null
    jobs.value.clear()
    progress.value = 0
    currentStep.value = null
    currentMessage.value = null
    error.value = null
    isStarting.value = false
  }

  return {
    // State
    currentJobId,
    jobs,
    progress,
    currentStep,
    currentMessage,
    error,
    isStarting,
    // Getters
    currentJob,
    status,
    isRunning,
    isComplete,
    isFailed,
    isPending,
    isIdle,
    // Actions
    subscribeToEvents,
    unsubscribeFromEvents,
    startJob,
    refreshJobs,
    clearCurrentJob,
    $reset,
  }
})
