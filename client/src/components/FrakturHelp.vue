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

<style scoped>
.fraktur-helper {
  text-align: center;
}
</style>
