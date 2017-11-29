<template>
  <div class="column is-one-third is-offset-one-third">
  <form class="configuration" @submit.prevent="fetchLines">
    <b-field label="Jahr" expanded>
      <b-field>
        <b-input v-model="year" min="1800" max="1899" type="number"
                 :disabled="isLoadingLines" expanded />
        <button class="button is-info" @click="randomizeYear"
                title="Zufälliges Jahr" type="button" :disabled="isLoadingLines">
          <b-icon icon="dice-4" />
        </button>
      </b-field>
    </b-field>
    <b-field label="Anzahl an Zeilen">
      <b-input v-model="taskSize" min="10" max="200" type="number"
               :disabled="isLoadingLines"/>
    </b-field>
    <button type="submit" :disabled="isLoadingLines"
            :class="{ 'button': true, 'is-primary': true, 'is-loading': isLoadingLines }">
      Start
    </button>
    <progress-bar v-if="percentDone" max="100" :current="percentDone" />
  </form>
  <a class="start-review" @click="changeScreen('list')">Bisherige Transkriptionen anzeigen</a>
  </div>
</template>

<script>
import { mapActions, mapMutations, mapState } from 'vuex'

import ProgressBar from './ProgressBar'

export default {
  name: 'Setup',
  components: { ProgressBar },
  computed: {
    percentDone () {
      if (!this.loadingProgress) {
        return 0
      } else {
        return Math.floor(this.loadingProgress.progress * 100)
      }
    },
    year: {
      get () { return this.$store.state.year },
      set (val) { this.$store.commit('setYear', parseInt(val)) }
    },
    taskSize: {
      get () { return this.$store.state.taskSize },
      set (val) { this.$store.commit('setTaskSize', parseInt(val)) }
    },
    ...mapState([
      'isLoadingLines',
      'loadingProgress'
    ])
  },
  methods: {
    ...mapMutations(['randomizeYear', 'changeScreen']),
    ...mapActions(['fetchLines'])
  }
}
</script>

<style scoped>
.configuration button.is-primary {
  width: 100%;
}
</style>
