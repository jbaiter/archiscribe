<template>
  <div class="column is-half is-offset-one-quarter">
    <div class="box">
      <p>
      Sie haben <strong>{{ numTranscriptions }}</strong> von
      <strong>{{ lines.length }}</strong> Zeilen transkribiert.
      </p>
    </div>
    <form v-if="githubUrl" @submit.prevent="resetWorkflow">
      <p>
      Ihre Änderungen wurden erfolgreich eingecheckt:
      <a :href="githubUrl">{{ githubUrl }}</a>
      </p>
      <button class="button" type="submit">
        Weiter transkribieren
      </button>
    </form>
    <form v-if="numTranscriptions > 0 && !githubUrl"
          class='submission' @submit.prevent="submit">
      <b-field>
        <b-switch v-model="anonymous">Anonym</b-switch>
      </b-field>
      <b-field grouped v-if="!anonymous">
        <b-field label="Name" expanded>
          <b-input :value='author' @input="updateAuthor" />
        </b-field>
        <b-field label="Email" expanded>
          <b-input type="email" :value='email' @input="updateEmail" />
        </b-field>
      </b-field>
      <b-field label="Kommentar">
        <b-input :value="comment" @input="updateComment" type="textarea" />
      </b-field>
      <button type="submit"
              :class="{ 'button': true, 'is-success': true, 'is-loading': isSubmitting }">
        Abschicken
      </button>
    </form>
  </div>
</template>

<script>
import { mapActions, mapMutations, mapState } from 'vuex'

export default {
  name: 'Submission',
  data () {
    return {
      anonymous: false
    }
  },
  computed: {
    githubUrl () {
      if (this.commit) {
        return `https://github.com/jbaiter/archiscribe-corpus/commit/${this.commit}`
      } else {
        return null
      }
    },
    numTranscriptions () {
      let num = 0
      for (let l of this.lines) {
        if (l.transcription) {
          num += 1
        }
      }
      return num
    },
    ...mapState(['lines', 'author', 'email', 'comment', 'commit', 'isSubmitting'])
  },
  watch: {
    anonymous (val) {
      if (val) {
        this.updateAuthor(null)
        this.updateEmail(null)
      }
    }
  },
  methods: {
    ...mapActions(['submit', 'resetWorkflow']),
    ...mapMutations(['resetWorkflow', 'updateAuthor', 'updateEmail', 'updateComment'])
  }
}
</script>

<style scoped>
</style>
