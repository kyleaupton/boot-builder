<script setup lang="ts">
import {
  Stepper,
  StepperItem,
  StepperSeparator,
  StepperTrigger,
  StepperIndicator,
  StepperTitle,
} from '@/components/ui/stepper'

defineProps<{
  currentStep: 1 | 2 | 3
}>()

const steps = [
  { step: 1, title: 'Select ISO' },
  { step: 2, title: 'Select Drive' },
  { step: 3, title: 'Flash' },
]
</script>

<template>
  <Stepper :model-value="currentStep" class="flex w-10/12 mx-auto items-start gap-2">
    <StepperItem
      v-for="(item, index) in steps"
      :key="item.step"
      :step="item.step"
      class="relative flex w-full flex-col items-center justify-center"
    >
      <StepperTrigger>
        <StepperIndicator class="bg-muted">
          <template v-if="currentStep > item.step">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </template>
          <template v-else>
            {{ item.step }}
          </template>
        </StepperIndicator>
      </StepperTrigger>
      <StepperSeparator
        v-if="index < steps.length - 1"
        class="absolute left-[calc(50%+20px)] right-[calc(-50%+10px)] top-5 block h-0.5 shrink-0 rounded-full bg-muted group-data-[state=completed]:bg-primary"
      />
      <StepperTitle class="mt-1 text-xs font-medium">
        {{ item.title }}
      </StepperTitle>
    </StepperItem>
  </Stepper>
</template>
