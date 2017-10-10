<template>
  <div>
    <div class="box">
      <p>
      Sie haben <strong>{{ numTranscriptions }}</strong> von
      <strong>{{ lines.length }}</strong> Zeilen transkribiert.
      </p>
    </div>
    <form v-if="githubUrl" @submit="onContinue">
      <p>
      Ihre Änderungen wurden erfolgreich eingecheckt:
      <a :href="githubUrl">{{githubUrl}}</a>
      </p>
      <button class="button" type="submit">
        Weiter transkribieren
      </button>
    </form>
    <form v-if="numTranscriptions > 0 && !githubUrl"
          class='submission' @submit="onSubmit">
      <b-field>
        <b-switch v-model="anonymous">Anonym</b-switch>
      </b-field>
      <b-field grouped v-if="!anonymous">
        <b-field label="Name" expanded>
          <b-input v-model='name' />
        </b-field>
        <b-field label="Email" expanded>
          <b-input type="email" v-model="email" />
        </b-field>
      </b-field>
      <b-field label="Kommentar">
        <b-input v-model="comment" maxlength="8192" type="textarea" />
      </b-field>
      <button type="submit"
              :class="{ 'button': true, 'is-success': true, 'is-loading': isPending }">
        Abschicken
      </button>
    </form>
  </div>
</template>

<script>
import bus from '../eventBus'

export default {
  name: 'Submission',
  props: ['lines'],
  data () {
    let state = {
      comment: null,
      isPending: false,
      githubUrl: null,
    };
    let identity = JSON.parse(localStorage.getItem('identity'));
    if (identity) {
      state = {...state, ...identity};
    } else {
      state['name'] = null;
      state['email'] = null;
      state['anonymous'] = false;
    }
    return state;
  },
  methods: {
    onSubmit (e) {
      e.preventDefault();
      this.$emit('submit', this.email, this.name, this.comment);
    },
    onContinue (e) {
      e.preventDefault();
      this.$emit('continue');
    }
  },
  computed: {
    numTranscriptions () {
      let num = 0;
      this.lines.forEach((l) => l.transcription ? num += 1 : null);
      return num;
    }
  },
  created () {
    let vm = this;
    bus.$on('submission-pending', () => vm.isPending = true);
    bus.$on('submission-success', (githubUrl) => {
      vm.githubUrl = githubUrl;
      vm.isPending = false;
    });
  }
}
</script>

<style scoped>
</style>
