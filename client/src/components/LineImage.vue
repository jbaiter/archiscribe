<template>
  <div class="line-image" @click="onClick">
    <div v-if="loading" class="control is-loading">
      <input disabled placeholder="Bild wird geladen..." class="input" />
    </div>
    <img :src="imageSrc" :class="classes" :title="label"
         @load="onLoaded" @error="onError"
         :style="{'display': loaded ? 'block': 'none'}" />
    <div v-if="error" class="box is-error">
      Fehler beim Laden des Zeilenbildes
    </div>
  </div>
</template>

<script>
export default {
  name: 'LineImage',
  props: ['image-src', 'type', 'label'],
  data () {
    return {
      'loading': false,
      'loaded': false,
      'error': false
    }
  },
  watch: {
    imageSrc (oldVal, newVal) {
      if (oldVal !== newVal && newVal) {
        this.onLoadStart()
      }
    }
  },
  created () {
    if (this.imageSrc) {
      this.onLoadStart()
    }
  },
  methods: {
    onClick () {
      this.$emit('click')
    },
    onLoadStart () {
      this.error = false
      this.loaded = false
      this.loading = true
    },
    onLoaded () {
      this.error = false
      this.loading = false
      this.loaded = true
    },
    onError () {
      this.loading = false
      this.error = true
    }
  },
  computed: {
    classes () {
      if (this.type === 'focus') {
        return {'focus-line': true}
      } else {
        let cls = {'context-line': true}
        cls[this.type] = true
        return cls
      }
    }
  }
}
</script>

<style scoped>
.line-image {
  text-align: center;
}

img.focus-line, img.context-line {
  width: auto;
}

img.context-line {
  filter: opacity(0.3);
}
</style>
