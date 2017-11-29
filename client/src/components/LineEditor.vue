<template>
<div class="column">
  <form class="transcription-widget" @submit.prevent="onNext">
    <a class="toggle-ctx previous" v-if="line.previous && !showPrevious"
        title="Mehr Kontext" @click="togglePrevious">
        <b-icon icon="dots-horizontal" />
      </a>
    <line-image v-if="showPrevious" @click="togglePrevious"
                type="previous" :image-src="line.previous" />
    <line-image type="focus" :image-src="line.line" />
    <line-image v-if="showNext" type="next" :image-src="line.next"
                @click="toggleNext" />
    <a class="toggle-ctx next" v-if="line.next && !showNext"
        title="Mehr Kontext" @click="toggleNext">
        <b-icon icon="dots-horizontal" />
    </a>
    <b-field>
      <div class="control is-clearfix is-expanded">
        <input ref="transcription" class="input mousetrap"
               :value="line.transcription" @input="updateTranscription($event.target.value)" />
      </div>
      <a class="button undo-btn" @click="previousLine" title="Vorherige Zeile korrigieren"
         :disabled="!hasPrevious">
        <b-icon icon="undo" />
      </a>
      <button class="button is-success done-btn" title="Nächste Zeile" type="submit">
        <b-icon icon="check" />
      </button>
      <a class="button is-danger delete-btn" @click="discardLine(currentLineIdx)"
          title="Zeile verwerfen">
        <b-icon icon="delete" />
      </a>
    </b-field>
  </form>
</div>
</template>

<script>
import { mapState } from 'vuex'

import bus from '../eventBus'

import LineImage from './LineImage'

export default {
  name: 'LineEditor',
  components: { LineImage },
  data () {
    return {
      showPrevious: false,
      showNext: false
    }
  },
  computed: {
    prevClasses () {
      return {
        'button': true,
        'undo-btn': true,
        'disabled': !this.hasPrevious
      }
    },
    ...mapState({
      currentLineIdx: 'currentLineIdx',
      hasNext: state => state.currentLineIdx < (state.lines.length - 1),
      hasPrevious: state => state.currentLineIdx > 0,
      line: state => state.lines[state.currentLineIdx]
    })
  },
  methods: {
    resetContext () {
      this.showPrevious = false
      this.showNext = false
    },
    onNext (e) {
      this.nextLine()
      this.resetContext()
    },
    togglePrevious () {
      this.showPrevious = !this.showPrevious
    },
    toggleNext () {
      this.showNext = !this.showNext
    },
    updateTranscription (val) {
      this.$store.commit('updateTranscription', {
        lineIdx: this.$store.state.currentLineIdx,
        transcription: val
      })
    },
    nextLine () {
      this.$store.commit('nextLine')
      this.resetContext()
    },
    discardLine () {
      this.$store.commit('discardLine')
      this.resetContext()
    },
    previousLine () {
      this.$store.commit('previousLine')
      this.resetContext()
    }
  },
  created () {
    const vm = this
    bus.$on('insert-grapheme', (grapheme) => {
      if (!vm.$refs.transcription) {
        return
      }
      vm.$store.commit('insertGrapheme', {
        grapheme,
        start: vm.$refs.transcription.selectionStart,
        end: vm.$refs.transcription.selectionEnd })
      vm.$refs.transcription.focus()
    })
  },
  mounted () {
    this.$refs.transcription.focus()
  }
}
</script>

<style scoped>
a.toggle-ctx {
  display: block;
  width: 100%;
  text-align: center;
  color: lightgray;
}

.toggle-ctx.previous {
  margin-bottom: 0.25em;
}

.toggle-ctx.next {
  margin-top: 0.25emm;
}

button.done-btn {
  width: 8vh;
}

button.skip-btn {
  width: 6vh;
}

button.delete-btn {
  width: 4vh;
}
</style>
