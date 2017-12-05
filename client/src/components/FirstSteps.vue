<template>
  <div class="content column is-three-fifths is-offset-one-fifth">
    <Guidelines v-if="showGuidelines" />
    <section v-else>
      <h1 class="title">Willkommen!</h1>
      <p>
      Ziel diesser Anwendung soll es sein einen Bestand an Trainingsdaten aus
      heterogenen Fraktur-Schriften zu erstellen, die <em>diplomatisch</em>
      transkribiert wurden. Diese Daten werden dann verwendet um ein möglichst
      robustes OCR-Modell für deutsche Frakturschriften zu trainieren.
      <p>
      <p>
      Bevor du loslegen kannst wäre es gut wenn du dir unsere <strong>Richtlinien</strong>
      durchliest, damit klar ist wie transkribiert wird.
      </p>
      <a class="button is-success" @click="toggleGuidelines">Zu den Richtlinien</a>
    </section>
  </div>
</template>

<script>
import Guidelines from './Guidelines'

export default {
  name: 'FirstSteps',
  components: { Guidelines },
  data () {
    return {
      showGuidelines: false
    }
  },
  methods: {
    toggleGuidelines () {
      this.showGuidelines = !this.showGuidelines
    }
  }
}
</script>

<style scoped>
</style>
