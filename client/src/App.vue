<template>
  <div id="app" class="container is-widescreen">
    <section class="header" v-if="showToolbar">
      <toolbar />
      <progress-bar v-if="showProgress"
                    :max="lines.length" :current="currentLineIdx"/>
    </section>
    <section class="main columns is-desktop"
             :style="{'margin-bottom': marginBottom}">
      <component v-bind:is='currentScreen' />
    </section>
    <footer ref="footer" class="footer" v-if="showFooter">
      <div class="help-toggle">
        <label class="switch">
          <span class="switch-left">
            <b-icon icon="keyboard" :type="showHelp ? 'is-black' : 'is-primary'"/>
          </span>
          <input type="checkbox" v-model="showHelp" />
          <span class="check" />
          <span class="switch-right">
            <b-icon icon="help" :type="showHelp ? 'is-primary' : 'is-black'"/>
          </span>
        </label>
      </div>
      <fraktur-help v-if="showHelp" />
      <keyboard v-else />
    </footer>
  </div>
</template>

<script>
import axios from 'axios'
import { mapState } from 'vuex'

import bus from './eventBus'
import Setup from './components/Setup'
import Toolbar from './components/Toolbar'
import Keyboard from './components/Keyboard'
import FrakturHelp from './components/FrakturHelp'
import ProgressBar from './components/ProgressBar'
import LineEditor from './components/LineEditor'
import MultiLineEditor from './components/MultiLineEditor'
import Submission from './components/Submission'

export default {
  name: 'app',
  components: {
    Toolbar,
    Keyboard,
    FrakturHelp,
    ProgressBar,
    'single': LineEditor,
    'multi': MultiLineEditor,
    'config': Setup,
    'submit': Submission
  },
  data () {
    return {
      showHelp: false,
      marginBottom: null
    }
  },
  computed: {
    showFooter () {
      return this.currentScreen !== 'config' && this.currentScreen !== 'submit'
    },
    showToolbar () {
      return this.currentScreen !== 'config'
    },
    showProgress () {
      return this.currentScreen === 'single'
    },
    ...mapState(['currentScreen', 'lines', 'metadata', 'currentLineIdx'])
  },
  watch: {
    showFooter (val) {
      this.$nextTick(this.adjustPaddingBottom)
    }
  },
  methods: {
    onTranscriptionDone (transcription) {
      this.lines[this.currentLineIdx].transcription = transcription
      this.onChangeLine(this.currentLineIdx + 1)
    },
    onChangeLine (idx) {
      this.currentLineIdx = idx
      if (this.currentLineIdx === this.lines.length) {
        this.currentLineIdx = 0
        this.currentScreen = 'multi'
      }
    },
    onPreviousLine () {
      this.onChangeLine(this.currentLineIdx - 1)
    },
    onDeleteLine () {
      this.lines.splice(this.currentLineIdx, 1)
      this.onChangeLine(this.currentLineIdx + 1)
    },
    onSubmit (email, name, comment) {
      localStorage.setItem(
        'identity',
        JSON.stringify({email, name, anonymous: (!name && !email)}))
      let data = {
        metadata: this.metadata,
        lines: this.lines.filter((l) => l.transcription !== '')
      }
      if (email && name) {
        data['author'] = {email, name}
      }
      if (comment) {
        data['commitMessage'] = comment
      }
      axios.post(
        '/api/transcriptions', data
      ).then((resp) => {
        bus.$emit('submission-success', resp.data.github)
        localStorage.clear('state')
      })
      bus.$emit('submission-pending')
    },
    onContinue: function () {
      this.isLoading = false
      this.lines = []
      this.metadata = null
      this.currentLineIdx = -1
      this.selectedYear = null
      this.currentScreen = 'config'
    },
    toggleReview: function () {
      this.showReview = !this.showReview
    },
    adjustPaddingBottom: function () {
      if (this.$refs.footer) {
        this.marginBottom = this.$refs.footer.offsetHeight + 'px'
      }
    }
  },
  created () {
    let vm = this
    bus.$on('change-screen', (screen) => {
      vm.currentScreen = screen
    })
  },
  mounted () {
    window.addEventListener('resize', this.adjustPaddingBottom)
  }
}
</script>

<style scoped>
body {
  overflow-y: hidden;
}

.main.columns {
  margin-top: 5vh;
}

.header {
  padding: 1rem;
  background: #f5f5f5;
}

.footer {
  position: fixed;
  bottom: 0px;
  padding: 1.5vh 0;
  max-width: 1344px;
}

.help-toggle {
  text-align: center;
  margin-bottom: 1em;
}

.switch-left {
  padding-right: 0.5em;
}

.switch-right {
  padding-left: 0.5em;
}
</style>
