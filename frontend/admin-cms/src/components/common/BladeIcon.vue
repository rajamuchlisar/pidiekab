<script setup lang="ts">
import { computed } from 'vue'
import * as outlineIcons from '@heroicons/vue/24/outline'
import * as solidIcons from '@heroicons/vue/24/solid'

interface Props {
  name: string
  solid?: boolean
  size?: string | number
  strokeWidth?: string | number
}

const props = withDefaults(defineProps<Props>(), {
  solid: false,
  size: '1.25rem',
  strokeWidth: 1.75,
})

const iconComponent = computed(() => {
  if (!props.name) return null
  const clean = props.name.replace(/^x-heroicon-[os]-/, '').toLowerCase()
  const pascal = clean
    .split('-')
    .map((s) => s.charAt(0).toUpperCase() + s.slice(1))
    .join('') + 'Icon'

  const iconSet: Record<string, any> = props.solid ? solidIcons : outlineIcons
  return iconSet[pascal] || null
})

const computedStyle = computed(() => {
  const s = props.size
  const sizeValue = typeof s === 'number' || /^\d+$/.test(String(s)) ? `${s}px` : s
  return {
    width: sizeValue,
    height: sizeValue,
    minWidth: sizeValue,
    minHeight: sizeValue,
  }
})
</script>

<template>
  <component
    :is="iconComponent"
    v-if="iconComponent"
    class="blade-icon"
    :style="computedStyle"
    :stroke-width="props.solid ? undefined : props.strokeWidth"
    aria-hidden="true"
  />
  <span v-else class="blade-icon-fallback" :style="computedStyle"></span>
</template>

<style scoped>
.blade-icon {
  display: inline-block;
  vertical-align: middle;
  flex-shrink: 0;
}
.blade-icon-fallback {
  display: inline-block;
}
</style>
