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
  display:inline-block;
  margin: 0.5em 1.25em;
  background: none;
  border-radius: 4px;
  text-align: center;
}

img {
  height:40px;
  color: #000;
  filter: drop-shadow(0px 1px 3px rgba(0, 0, 0, 0.5));
}
</style>
