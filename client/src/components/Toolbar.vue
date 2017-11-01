<template>
  <b-field grouped class="is-pulled-right">
    <a class="button" @click="showSingle" title="Detailansicht">
      <b-icon icon="search" />
    </a>
    <a class="button" @click="showMulti" title="Zeilenübersicht">
      <b-icon icon="format_align_left" />
    </a>
    <a class="button" @click="pickAnother" title="Anderes Buch auswählen">
      <b-icon icon="autorenew" />
    </a>
    <a class="button is-success" @click="submit" title="Transkriptionen abschicken">
      <b-icon icon="send" />
    </a>
  </b-field>
</template>

<script>
import { mapActions, mapMutations, mapState } from 'vuex'

export default {
  name: 'Toolbar',
  computed: {
    ...mapState(['metadata'])
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
    ...mapMutations(['changeScreen']),
    ...mapActions(['fetchLines'])
  }
}
</script>

<style scope>
</style>
