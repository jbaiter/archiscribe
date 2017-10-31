<template>
  <!-- TODO: Information about title, year, number of transcribed lines -->
  <div class="column is-half is-offset-one-quarter">
    <h1 class="title">Sitzung wiederherstellen?</h1>
    <p>
    In einer vergangenen Sitzung wurden <strong>{{ numTranscribedLines }}</strong> Zeilen aus
      <strong><a :href="previousSession.metadata['identifier-access']">
        {{ previousSession.metadata.title }} ({{ previousSession.metadata.year }})
      </a></strong> transkribiert und nicht abgeschickt.
    <p>
    <p>
    Wollen Sie diese Sitzung <strong>wiederherstellen und fortsetzen?</strong> Wenn Sie die
    Sitzung <strong>verwerfen</strong> werden ihre bisherigen Transkriptionen <strong>unwiderruflich gelöscht</strong>.
    </p>
    <a class="button is-success" @click="restoreSession">
      <b-icon icon="restore" /><span>Wiederherstellen</span>
    </a>
    <a class="button is-danger" @click="discardSession(true)">
      <b-icon icon="delete_forever" /><span>Verwerfen</span>
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
