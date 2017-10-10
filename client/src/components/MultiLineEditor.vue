<template>
  <div class="column review">
    <div ref="line" key="line.line" class="card line-review"
          v-for="(line, idx) in lines">
      <div class="card-image">
        <figure class="image">
          <line-image :image-src="line.line" type="focus" />
        </figure>
      </div>
      <div class="card-content">
        <b-field>
          <div class="control is-clearfix is-expanded">
            <input @focus="focusLine(idx)" ref="transcription"
                  class="input mousetrap" v-model="line.transcription" />
          </div>
          <a class="button focus-btn" @click="zoomLine(idx)"
              title="Detailansicht">
            <b-icon icon="search" />
          </a>
          <a class="button is-danger delete-btn" @click="deleteLine(idx)"
              title="Zeile verwerfen">
            <b-icon icon="delete" />
          </a>
        </b-field>
      </div>
    </div>
  </div>
</template>

<script>
import bus from '../eventBus'

import LineImage from './LineImage'

export default {
  name: 'MultiLineEditor',
  components: { LineImage },
  props: ['lines', 'currentIdx'],
  data () {
    return {
      'activeInput': -1
    };
  },
  methods: {
    focusLine (idx) {
      this.activeInput = idx;
      this.$emit('change-line', idx);
    },
    zoomLine (idx) {
      this.$emit('change-line', idx);
      bus.$emit('change-screen', 'single');
    },
    deleteLine (idx) {
      this.lines.splice(idx, 1);
    }
  },
  created () {
    let vm = this;
    bus.$on('insert-grapheme', (grapheme) => {
      if (this.activeInput >= 0) {
        let ref = this.$refs.transcription[this.activeInput];
        let transcription = vm.lines[this.activeInput].transcription;
        transcription = ref.value.substring(0, ref.selectionStart) +
                        grapheme + ref.value.substring(ref.selectionEnd);
        vm.lines[this.activeInput].transcription = transcription;
        ref.focus();
      }
    });
  },
  mounted () {
    if (this.currentIdx >= 0) {
      this.$refs.line[this.currentIdx].scrollIntoView();
    }
  }
}
</script>

<style scoped>
.line-review {
  margin: 1rem;
  padding: 1.5rem 1.5rem 0 1.5rem;
}
</style>
