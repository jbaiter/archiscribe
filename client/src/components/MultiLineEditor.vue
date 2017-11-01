<template>
  <div class="column review">
    <div ref="line" :key="line.line" class="card line-review"
          v-for="(line, idx) in lines">
      <div class="card-image">
        <figure class="image">
          <line-image :image-src="line.line" type="focus" />
        </figure>
      </div>
      <div class="card-content">
        <b-field>
          <div class="control is-clearfix is-expanded">
            <input @focus="changeLine(idx)" ref="transcription"
                   class="input mousetrap" :value="line.transcription"
                   @input="onInput(idx, $event.target.value)"
                   @keyup.enter="focusLine(idx+1)" />
          </div>
          <a class="button focus-btn" @click="zoomLine(idx)"
              title="Detailansicht">
            <b-icon icon="search" />
          </a>
          <a class="button is-danger delete-btn" @click="discardLine(idx)"
              title="Zeile verwerfen">
            <b-icon icon="delete" />
          </a>
        </b-field>
      </div>
    </div>
  </div>
</template>

<script>
import { mapMutations, mapState } from 'vuex'

import bus from '../eventBus'

import LineImage from './LineImage'

export default {
  name: 'MultiLineEditor',
  components: { LineImage },
  methods: {
    focusLine (idx) {
      if (idx >= this.$refs.transcription.length) {
        idx = 0
      }
      this.$refs.line[idx].scrollIntoView()
      this.$refs.transcription[idx].focus()
    },
    zoomLine (idx) {
      this.changeLine(idx)
      this.changeScreen('single')
    },
    onInput (idx, val) {
      this.updateTranscription({
        lineIdx: idx,
        transcription: val })
    },
    onInsertGrapheme (grapheme) {
      if (this.$refs.transcription.length === 0 || this.currentLineIdx < 0) {
        return
      }
      let ref = this.$refs.transcription[this.currentLineIdx]
      this.$store.commit('insertGrapheme', {
        grapheme,
        start: ref.selectionStart,
        end: ref.selectionEnd
      })
      ref.focus()
    },
    ...mapMutations([
      'changeLine', 'changeScreen', 'discardLine', 'updateTranscription'])
  },
  computed: mapState(['currentLineIdx', 'lines']),
  created () {
    bus.$on('insert-grapheme', this.onInsertGrapheme)
  },
  mounted () {
    if (this.currentLineIdx >= 0) {
      this.$refs.line[this.currentLineIdx].scrollIntoView()
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
