import Vue from 'vue'
import Vuex from 'vuex'

import axios from 'axios'

import { getRandomInt, preloadLineImages } from './util'

Vue.use(Vuex)
let storedSession = JSON.parse(localStorage.getItem('session'))
if (storedSession && !storedSession.document) {
  if (!storedSession.metadata) {
    localStorage.removeItem('session')
    storedSession = null
  } else {
    // Migrate from old format
    storedSession.document = {
      id: storedSession.metadata.id,
      title: storedSession.metadata.title,
      year: storedSession.year,
      manifest: `https://iiif.archivelab.org/iiif/${storedSession.id}/manifest.json`
    }
    storedSession.metadata = undefined
  }
}

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
    document: undefined,
    currentLineIdx: -1,
    isSubmitting: false,
    author: localStorage.getItem('identity.author'),
    email: localStorage.getItem('identity.email'),
    comment: null,
    commit: null
  },
  mutations: {
    discardSession (state, restart) {
      state.previousSession = null
      localStorage.removeItem('session')
      if (restart) {
        this.commit('changeScreen', 'config')
      }
    },
    restoreSession (state) {
      let previous = state.previousSession
      state.previousSession = null
      this.replaceState({...state, ...previous, ...{currentScreen: 'single'}})
      preloadLineImages(previous.lines.map(l => l.line))
    },
    startLoading (state) {
      state.isLoadingLines = true
    },
    stopLoading (state) {
      state.isLoadingLines = false
    },
    setDocument (state, document) {
      state.document = document
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
      state.document = null
      state.currentScreen = 'config'
      state.commit = null
      state.comment = null
    },
    updateEmail (state, email) {
      state.email = email
      localStorage.setItem('identity.email', email)
    },
    updateAuthor (state, author) {
      state.author = author
      localStorage.setItem('identity.author', author)
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
        'document', (evt) => commit('setDocument', JSON.parse(evt.data)))
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
      axios.post('/api/documents', {
        document: {
          lines: state.lines.filter(l => l.transcription),
          ...this.state.document
        },
        author: state.author ? `${state.author} <${state.email}>` : null,
        comment: state.comment
      }).then(({ data }) => {
        commit('finishSubmit')
        commit('setCommitHash', data.commit)
        commit('discardSession', false)
      })
      // TODO: Handle error
    }
  }
})

store.subscribe((mutation, state) => {
  if (!['updateTranscription', 'setLines'].includes(mutation.type)) {
    return
  }
  let { year, taskSize, lines, document, currentLineIdx } = state
  localStorage.setItem('session', JSON.stringify({
    year, taskSize, lines, document, currentLineIdx }))
})

export default store
