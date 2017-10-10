<template>
  <form class="configuration" @submit="onConfigured">
    <b-field label="Jahr" expanded>
      <b-field>
        <b-input v-model="year" min="1800" max="1899" type="number" expanded />
        <button class="button is-info" @click="randomizeYear"
                title="Zufälliges Jahr" type="button">
          <b-icon icon="casino" />
        </button>
      </b-field>
    </b-field>
    <b-field label="Anzahl an Zeilen">
      <b-input v-model="taskSize" min="10" max="200" type="number" />
    </b-field>
    <button type="submit"
            :class="{ 'button': true, 'is-primary': true, 'is-loading': isLoading }">
      Start
    </button>
  </form>
</template>

<script>
function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min)) + min;
}

export default {
  name: 'Setup',
  props: ['isLoading', 'defaultTaskSize', 'presetYear'],
  data () {
    return {
      'year': this.presetYear || getRandomInt(1800, 1900),
      'taskSize': this.defaultTaskSize || 50
    };
  },
  methods: {
    onConfigured (e) {
      e.preventDefault();
      this.$emit('configured', parseInt(this.year), parseInt(this.taskSize));
    },
    randomizeYear () {
      this.year = getRandomInt(1800, 1899);
    }
  }
}
</script>

<style scoped>
.configuration button.is-primary {
  width: 100%;
}
</style>
