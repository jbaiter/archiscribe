<template>
  <div id="app" class="container is-widescreen">
    <section class="header" v-if="showToolbar">
      <div class="columns">
        <div class="column is-four-fifths">
          <h1>
            <b-icon v-if="isReview" icon="message-draw" type="is-info" title="Review Modus"/>
            <a :href="`https://archive.org/details/${activeDocument.id}`">{{ truncatedTitle }}</a> ({{activeDocument.year}})</h1>
        </div>
        <div class="column">
          <toolbar />
        </div>
      </div>
      <progress-bar v-if="showProgress"
                    :max="lines.length" :current="currentLineIdx"/>
    </section>
    <section class="main columns is-desktop"
             :style="{'margin-bottom': marginBottom + 'px'}">
      <component v-bind:is='currentScreen' />
    </section>
    <div class="footer-wrapper">
      <a v-if="showFooter" @click="toggleBottom" class="toggle-bottom is-pulled-right"
         :title="toggleBottomTitle">
        <b-icon :icon="toggleBottomIcon" />
      </a>
      <footer ref="footer" class="footer" v-if="showFooter && !bottomIsHidden">
        <div class="help-toggle">
          <label class="switch">
            <span class="switch-left">
              <b-icon icon="keyboard" :type="showHelp ? 'is-black' : 'is-primary'"/>
            </span>
            <input type="checkbox" v-model="showHelp" />
            <span class="check" />
            <span class="switch-right">
              <b-icon icon="help-circle" :type="showHelp ? 'is-primary' : 'is-black'"/>
            </span>
          </label>
        </div>
        <fraktur-help v-if="showHelp" />
        <keyboard v-else />
      </footer>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

import bus from './eventBus'
import SessionRestore from './components/SessionRestore'
import Setup from './components/Setup'
import Toolbar from './components/Toolbar'
import Keyboard from './components/Keyboard'
import FrakturHelp from './components/FrakturHelp'
import ProgressBar from './components/ProgressBar'
import LineEditor from './components/LineEditor'
import MultiLineEditor from './components/MultiLineEditor'
import Submission from './components/Submission'
import DocumentList from './components/DocumentList'
import FirstSteps from './components/FirstSteps'

export default {
  name: 'app',
  components: {
    Toolbar,
    Keyboard,
    FrakturHelp,
    ProgressBar,
    'intro': FirstSteps,
    'single': LineEditor,
    'multi': MultiLineEditor,
    'config': Setup,
    'submit': Submission,
    'restore': SessionRestore,
    'list': DocumentList
  },
  data () {
    return {
      showHelp: false,
      marginBottom: null,
      bottomIsHidden: false
    }
  },
  computed: {
    toggleBottomIcon () {
      if (this.bottomIsHidden) {
        return 'chevron-up'
      } else {
        return 'chevron-down'
      }
    },
    toggleBottomTitle () {
      if (this.bottomIsHidden) {
        return 'Tastatur/Lesehilfe anzeigen'
      } else {
        return 'Tastatur/Lesehilfe verstecken'
      }
    },
    truncatedTitle () {
      let cut = this.activeDocument.title.indexOf(' ', 96)
      if (cut === -1) return this.activeDocument.title
      return this.activeDocument.title.substring(0, cut) + '‥'
    },
    showFooter () {
      return ['multi', 'single'].includes(this.currentScreen)
    },
    showToolbar () {
      return ['multi', 'single', 'submit'].includes(this.currentScreen)
    },
    showProgress () {
      return this.currentScreen === 'single'
    },
    showKeyboard () {
      return !this.showHelp && this.showBottom
    },
    ...mapState(['currentScreen', 'lines', 'activeDocument', 'currentLineIdx']),
    ...mapGetters(['isReview'])
  },
  watch: {
    showFooter (val) {
      this.$nextTick(this.adjustPaddingBottom)
    },
    showHelp (val) {
      this.$nextTick(this.adjustPaddingBottom)
    }
  },
  methods: {
    toggleBottom () {
      this.bottomIsHidden = !this.bottomIsHidden
      this.$nextTick(this.adjustPaddingBottom)
    },
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
    onContinue: function () {
      this.isLoading = false
      this.lines = []
      this.activeDocument = null
      this.currentLineIdx = -1
      this.selectedYear = null
      this.currentScreen = 'config'
    },
    toggleReview: function () {
      this.showReview = !this.showReview
    },
    adjustPaddingBottom: function () {
      if (this.$refs.footer) {
        this.marginBottom = this.$refs.footer.offsetHeight
      } else {
        this.marginBottom = 0
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

<style>
@import url('https://fonts.googleapis.com/css?family=Vollkorn');

body {
  font-family: 'Vollkorn', serif;
}

.main.columns {
  margin-top: 5vh;
}

.header {
  padding: 1rem;
  background: #f5f5f5;
}

.footer {
  padding: 1.5vh 0;
}

.footer-wrapper {
  width: 100%;
  position: fixed;
  bottom: 0px;
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

.header h1 {
  display: inline-block;
  font-size: 1.3rem;
}
</style>
