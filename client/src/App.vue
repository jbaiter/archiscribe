<template>
  <div id="app" class="container is-widescreen">
    <section class="header">
      <volume-info v-if="currentScreen !== 'config'"
        :metadata="metadata" :isLoading="isLoading"
        @pick-another="fetchLines" />
      <progress-bar v-if="currentScreen === 'single'"
        :max="lines.length" :current="currentLineIdx" />
    </section>
    <section class="main columns is-desktop"
              :style="{'margin-bottom': marginBottom}">
      <multi-line-editor v-if="currentScreen === 'multi'" :lines="lines"
              :currentIdx="currentLineIdx"
              @change-line="onChangeLine" />
      <div v-if="currentScreen === 'config'"
            class="column is-one-third is-offset-one-third">
        <setup
          @configured="onConfigured"
          :presetYear="selectedYear"
          :defaultTaskSize="taskSize"
          :isLoading="isLoading" />
      </div>
      <div v-if="currentScreen === 'single'" class="column">
        <line-editor
          :key="lines[currentLineIdx].line"
          :line="lines[currentLineIdx]"
          :hasPrevious="currentLineIdx > 0"
          @transcription-done="onTranscriptionDone"
          @previous-line="onPreviousLine"
          @delete-line="onDeleteLine" />
      </div>
      <div v-if="currentScreen === 'submit'"
            class="column is-half is-offset-one-quarter">
        <submission :lines="lines"
          @submit="onSubmit" @continue="onContinue"/>
      </div>
    </section>
    <footer ref="footer" class="footer"
      v-if="currentScreen !== 'config' && currentScreen !== 'submit'">
      <keyboard />
    </footer>
  </div>
</template>

<script>
import axios from 'axios'

import bus from './eventBus'
import Setup from './components/Setup'
import VolumeInfo from './components/VolumeInfo'
import Keyboard from './components/Keyboard'
import ProgressBar from './components/ProgressBar'
import LineEditor from './components/LineEditor'
import MultiLineEditor from './components/MultiLineEditor'
import Submission from './components/Submission'

function preloadLineImages(lines) {
  let preloadImage = (url, retryCount=0) => {
    let img = new Image();
    img.src = url;
    img.onerror = () => {
      if (retryCount < 5) {
        console.log(`Retrying ${url} for the ${retryCount + 1}. time`);
        retryCount += 1;
        setTimeout(preloadImage, retryCount*1000, url, retryCount);
      } else {
        console.log(`Failed loading ${url}.`);
        bus.$emit('image-preload-failed', url);
      }
    }
    img.onload = () => bus.$emit('image-preloaded', url);
  };

  lines.forEach((url) => {
    preloadImage(url);
  });
}

export default {
  name: 'app',
  components: {
    Setup, VolumeInfo, Keyboard, ProgressBar, LineEditor, MultiLineEditor,
    Submission
  },
  data () {
    let state = {
      currentScreen: 'config',  // config, single, multi or submit
      isLoading: false,
      eventSource: undefined,
      progress: undefined,
      selectedYear: undefined,
      taskSize: window.DEFAULT_TASK_SIZE,
      lines: [],
      metadata: null,
      currentLineIdx: -1,
      marginBottom: null
    };
    let storedState = localStorage.getItem('state');
    if (state) {
      Object.assign(state, JSON.parse(storedState));
    }
    return state;
  },
  updated () {
    localStorage.setItem('state', JSON.stringify(this.$data));
  },
  methods: {
    onConfigured (year, taskSize) {
      this.selectedYear = year;
      this.taskSize = taskSize;
      this.fetchLines();
    },
    fetchLines () {
      let vm = this;
      vm.isLoading = true;
      vm.eventSource = new EventSource(
        `/api/lines/${vm.selectedYear}?taskSize=${vm.taskSize}`);
      vm.eventSource.addEventListener("metadata", (evt) => {
        let data = JSON.parse(evt.data);
        vm.metadata = data;
      });
      vm.eventSource.addEventListener("progress", (evt) => {
        let data = JSON.parse(evt.data);
        vm.progress = Math.floor(data.progress * 100);
        console.log(data);
      });
      vm.eventSource.addEventListener("lines", (evt) => {
        let lines = JSON.parse(evt.data);
        vm.lines = lines.map(
          (line) => ({...line, 'transcription': ''}));;
        preloadLineImages(vm.lines.map((l) => l.line));
        vm.isLoading = false;
        vm.currentLineIdx = 0;
        vm.currentScreen = 'single';
        vm.eventSource.close()
        vm.eventSource = undefined;
      });
    },
    onTranscriptionDone (transcription) {
      this.lines[this.currentLineIdx].transcription = transcription;
      this.onChangeLine(this.currentLineIdx+1);
    },
    onChangeLine (idx) {
      this.currentLineIdx = idx;
      if (this.currentLineIdx == this.lines.length) {
        this.currentLineIdx = 0;
        this.currentScreen = 'multi';
      }
    },
    onPreviousLine () {
      this.onChangeLine(this.currentLineIdx-1);
    },
    onDeleteLine () {
      this.lines.splice(this.currentLineIdx, 1);
      this.onChangeLine(this.currentLineIdx+1);
    },
    onSubmit (email, name, comment) {
      localStorage.setItem(
        'identity',
        JSON.stringify({email, name, anonymous: (!name && !email)}));
      let vm = this;
      let data = {
        metadata: this.metadata,
        lines: this.lines.filter((l) => l.transcription != '')
      };
      if (email && name) {
        data['author'] = {email, name};
      }
      if (comment) {
        data['commitMessage'] = comment;
      }
      axios.post(
        '/api/transcriptions', data
      ).then((resp) => {
        bus.$emit('submission-success', resp.data.github);
        localStorage.clear('state');
      });
      bus.$emit('submission-pending');
    },
    onContinue: function() {
      this.isLoading = false;
      this.lines = [];
      this.metadata = null;
      this.currentLineIdx = -1;
      this.selectedYear = null;
      this.currentScreen = 'config';
    },
    toggleReview: function() {
      this.showReview = !this.showReview;
    },
    handleResize: function() {
      if (this.$refs.footer) {
        this.marginBottom = this.$refs.footer.offsetHeight + 'px';
      }
    }
  },
  created () {
    let vm = this;
    bus.$on('change-screen', (screen) => {
      vm.currentScreen = screen;
    });
  },
  mounted () {
    this.handleResize();
    window.addEventListener('resize', this.handleResize);
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
  padding: 1rem 1.5rem 2.5rem;
  max-width: 1344px;
}
</style>
