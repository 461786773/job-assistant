<template>
  <aside class="guide-note" :class="[`tone-${tone}`]">
    <p v-if="title" class="guide-note-title">{{ title }}</p>
    <div class="guide-note-body">
      <slot />
    </div>
  </aside>
</template>

<script setup>
defineProps({
  title: { type: String, default: '' },
  tone: { type: String, default: 'calm' }, // calm | soft | tip
})
</script>
