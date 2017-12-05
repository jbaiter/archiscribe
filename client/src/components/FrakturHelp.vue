<template>
<div class="fraktur-helper">
  <FrakturLetter v-for="letter in letters" :letter="letter" :key="letter" />
</div>
</template>

<script>
import FrakturLetter from './FrakturLetter'

export default {
  name: 'FrakturHelp',
  components: { FrakturLetter },
  data () {
    return {
      letters: [
        'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'IJ', 'K', 'L', 'M', 'N', 'O',
        'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
        'ch', 'ck', 'sch', 'sp', 'ss', 'st', 'ß', 'th', 'tz']
    }
  }
}
</script>

<style>
.fraktur-helper {
  text-align: center;
}

div.fraktur-letter {
  display:inline-block;
  margin: 0.5em 1.25em;
  background: none;
  border-radius: 4px;
  text-align: center;
}

div.fraktur-letter>img {
  height:40px;
  color: #000;
  filter: drop-shadow(0px 1px 3px rgba(0, 0, 0, 0.5));
}
</style>
