import Vue from 'vue'
import Vuex from 'vuex'

import axios from 'axios'

import { getRandomInt, preloadLineImages } from './util'

Vue.use(Vuex)
let storedSession = JSON.parse(localStorage.getItem('session'))

let store = new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state: {
    previousSession: storedSession,
    year: getRandomInt(1800, 1900),
    taskSize: 50,
    currentScreen: storedSession ? 'restore' : 'config',
    isLoadingLines: false,
    loadingProgress: undefined,
    lines: [],
    metadata: undefined,
    currentLineIdx: -1,
    isSubmitting: false,
    author: null,
    email: null,
    comment: null,
    commit: null
  },
  mutations: {
    discardSession (state) {
      state.previousSession = null
      localStorage.removeItem('session')
      this.commit('changeScreen', 'config')
    },
    restoreSession (state) {
      let previous = state.previousSession
      state.previousSession = null
      this.replaceState({...state, ...previous, ...{currentScreen: 'single'}})
    },
    startLoading (state) {
      state.isLoadingLines = true
    },
    stopLoading (state) {
      state.isLoadingLines = false
    },
    setMetadata (state, metadata) {
      state.metadata = metadata
    },
    updateProgress (state, progress) {
      state.loadingProgress = progress
    },
    changeLine (state, lineIdx) {
      state.currentLineIdx = lineIdx
    },
    previousLine (state) {
      if (state.currentLineIdx > 0) {
        state.currentLineIdx -= 1
      }
    },
    nextLine (state) {
      if (state.currentLineIdx < (state.lines.length - 1)) {
        state.currentLineIdx += 1
      } else {
        this.commit('changeScreen', 'multi')
        state.currentLineIdx = 0
      }
    },
    discardLine (state, lineIdx) {
      if (lineIdx === undefined) {
        lineIdx = state.currentLineIdx
      }
      state.lines.splice(lineIdx, 1)
      if (lineIdx === state.currentLineIdx && state.currentLineIdx === state.lines.length) {
        state.currentLineIdx -= 1
      }
    },
    changeScreen (state, screen) {
      state.currentScreen = screen
    },
    setLines (state, lines) {
      this.commit('stopLoading')
      this.state.loadingProgress = undefined
      state.lines = lines
      this.commit('changeLine', 0)
      this.commit('changeScreen', 'single')
      preloadLineImages(this.state.lines.map(l => l.line))
    },
    setYear (state, year) {
      state.year = year
    },
    setTaskSize (state, taskSize) {
      state.taskSize = taskSize
    },
    randomizeYear (state) {
      this.commit('setYear', getRandomInt(1800, 1899))
    },
    insertGrapheme (state, { grapheme, start, end }) {
      let line = state.lines[state.currentLineIdx]
      let transcription = line.transcription.substring(0, start) +
                          grapheme +
                          line.transcription.substring(end)
      this.commit(
        'updateTranscription',
        { lineIdx: state.currentLineIdx,
          transcription })
    },
    updateTranscription (state, {lineIdx, transcription}) {
      let line = state.lines[lineIdx]
      Vue.set(state.lines, lineIdx, {...line, transcription})
    },
    startSubmit (state) {
      state.isSubmitting = true
    },
    finishSubmit (state) {
      state.isSubmitting = false
    },
    setCommitHash (state, hash) {
      state.commit = hash
    },
    resetWorkflow (state) {
      state.year = getRandomInt(1800, 1900)
      state.lines = []
      state.currentLineIdx = -1
      state.metadata = null
      state.currentScreen = 'config'
      state.commit = null
      state.comment = false
    },
    updateEmail (state, email) {
      state.email = email
    },
    updateAuthor (state, author) {
      state.author = author
    },
    updateComment (state, comment) {
      state.comment = comment
    }
  },
  actions: {
    fetchLines ({ commit, state }) {
      commit('startLoading')
      const eventSource = new EventSource(`/api/lines/${state.year}?taskSize=${state.taskSize}`)
      eventSource.addEventListener(
        'metadata', (evt) => commit('setMetadata', JSON.parse(evt.data)))
      eventSource.addEventListener(
        'progress', (evt) => commit('updateProgress', JSON.parse(evt.data)))
      eventSource.addEventListener('lines', (evt) => {
        commit('setLines', JSON.parse(evt.data).map(
          (line) => ({...line, transcription: ''})))
        eventSource.close()
      })
    },
    submit ({ commit, state }) {
      commit('startSubmit')
      axios.post('/api/transcriptions', {
        id: state.metadata.identifier,
        lines: state.lines,
        author: state.author ? `${state.author} <${state.email}>` : null,
        comment: state.comment,
        metadata: state.metadata
      }).then(({ data }) => {
        commit('finishSubmit')
        commit('setCommitHash', data.commit)
        commit('discardSession')
      })
      // TODO: Handle error
    }
  }
})

store.subscribe((mutation, state) => {
  if (mutation.type !== 'updateTranscription') {
    return
  }
  let { year, taskSize, lines, metadata, currentLineIdx } = state
  localStorage.setItem('session', JSON.stringify({
    year, taskSize, lines, metadata, currentLineIdx }))
})

export default store
