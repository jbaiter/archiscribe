<template>
  <div class="fraktur-letter">
    <img :src="letterImage" :title="letter" :alt="letter" />
  </div>
</template>

<script>
export default {
  name: 'FrakturLetter',
  props: ['letter'],
  computed: {
    letterImage () {
      return require(`../../img/letters/${this.letter.toLowerCase()}.svg`)
    }
  }
}
</script>

<style scoped>
div {
  display: inline-block;
  background: none;
  text-align: center;
}
</style>
