<template>
  <!-- TODO: Information about title, year, number of transcribed lines -->
  <div class="column is-half is-offset-one-quarter">
    <h1 class="title">Sitzung wiederherstellen?</h1>
    <p v-if="previousSession.activeDocument.history">
    In einer vergangenen Sitzung wurde ein Review von
    <strong><a :href="`https://archive.org/details/${previousSession.activeDocument.id}`">
        {{ previousSession.activeDocument.title }} ({{ previousSession.activeDocument.year }})
      </a></strong><span> </span>angefangen und nicht abgeschlossen.
    </p>
    <p v-else>
    In einer vergangenen Sitzung wurden <strong>{{ numTranscribedLines }}</strong> Zeilen aus
    <strong><a :href="`https://archive.org/details/${previousSession.activeDocument.id}`">
        {{ previousSession.activeDocument.title }} ({{ previousSession.activeDocument.year }})
      </a></strong><span> </span>transkribiert und nicht abgeschickt.
    </p>
    <p>
    Wollen Sie diese Sitzung <strong>wiederherstellen und fortsetzen?</strong> Wenn Sie die
    Sitzung <strong>verwerfen</strong> werden ihre Änderungen <strong>unwiderruflich gelöscht</strong>.
    </p>
    <a class="button is-success" @click="restoreSession">
      <b-icon icon="restore" /><span>Wiederherstellen</span>
    </a>
    <a class="button is-danger" @click="discardSession(true)">
      <b-icon icon="delete-forever" /><span>Verwerfen</span>
    </a>
  </div>
</template>

<script>
import { mapMutations, mapState } from 'vuex'

export default {
  name: 'SessionRestore',
  computed: {
    numTranscribedLines () {
      return this.previousSession.lines.filter((l) => l.transcription !== '').length
    },
    ...mapState(['previousSession'])
  },
  methods: mapMutations(['restoreSession', 'discardSession'])
}
</script>

<style scoped>
  div {
    text-align: center;
  }

  p {
    text-align: justify;
    -webkit-hyphens: auto;
    -moz-hyphens: auto;
    hyphens: auto;
  }

  p:last-of-type {
    margin-bottom: 1em;
  }
</style>
