<template>
  <div class="line-image" @click="onClick">
    <div v-if="loading" class="control is-loading">
      <input disabled placeholder="Bild wird geladen..." class="input" />
    </div>
    <img v-if="loaded" :src="imageSrc" :class="classes" :title="label" />
    <div v-if="error" class="box is-error">
      Fehler beim Laden des Zeilenbildes
      <button class="button" title="Nochmals versuchen" @click="loadImage">
        <b-icon icon="refresh" />
      </button>
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
    };
  },
  methods: {
    onClick () {
      this.$emit('click');
    },
    loadImage () {
      let img = new Image();
      img.src = this.imageSrc;
      let vm = this;
      vm.loading = true;
      img.onload = () => {
        vm.error = false;
        vm.loading = false;
        vm.loaded = true;
      };
      img.onerror = () => {
        vm.loading = false;
        vm.error = true;
      };
    }
  },
  computed: {
    classes () {
      if (this.type === 'focus') {
        return {'focus-line': true};
      } else {
        let cls = {'context-line': true}
        cls[this.type] = true;
        return cls;
      }
    },
  },
  created () {
    this.loadImage();
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
