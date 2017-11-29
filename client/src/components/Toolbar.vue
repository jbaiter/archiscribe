<template>
  <b-field grouped class="is-pulled-right">
    <a class="button" @click="showSingle" title="Detailansicht">
      <b-icon icon="magnify" />
    </a>
    <a class="button" @click="showMulti" title="Zeilenübersicht">
      <b-icon icon="format-align-left" />
    </a>
    <a v-if="isReview" class="button is-danger" @click="backToList" title="Zurück zur Übersicht">
      <b-icon icon="replay" />
    </a>
    <a v-else class="button is-danger" @click="pickAnother" title="Anderes Buch auswählen">
      <b-icon icon="autorenew" />
    </a>
    <a class="button is-success" @click="submit" title="Transkriptionen abschicken">
      <b-icon icon="send" />
    </a>
  </b-field>
</template>

<script>
import { mapActions, mapMutations, mapState, mapGetters } from 'vuex'

export default {
  name: 'Toolbar',
  computed: {
    ...mapState(['activeDocument']),
    ...mapGetters(['isReview'])
  },
  methods: {
    pickAnother () {
      this.fetchLines()
      this.changeScreen('config')
    },
    showSingle () {
      this.changeScreen('single')
    },
    showMulti () {
      this.changeScreen('multi')
    },
    submit () {
      this.changeScreen('submit')
    },
    backToList () {
      this.changeScreen('list')
    },
    ...mapMutations(['changeScreen']),
    ...mapActions(['fetchLines'])
  }
}
</script>

<style scope>
</style>
