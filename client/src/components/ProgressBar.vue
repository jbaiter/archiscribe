<template>
  <progress class="progress is-info" :value="current" :max="max">
    {{ current }}/{{ max }}
  </progress>
</template>

<script>
export default {
  name: 'ProgressBar',
  props: ['max', 'current']
}
</script>

<style scoped>
.progress {
  border-radius: 0;
}
</style>
