<template>
  <div class="character-picker">
    <button class="keycap character" v-for="charDef in allChars"
            :title="charDef.key ? 'alt+' + charDef.key : ''"
            @click="onCharClick(charDef.char)">
      {{ charDef.char }}
    </button>
  </div>
</template>

<script>
import Mousetrap from 'mousetrap'

import bus from '../eventBus'

let SPECIAL_CHARACTERS = {
    'ſ': 's',
    'ꝛ': 'r',
    'Æ': null,
    'Œ': null,
    'æ': 'a',
    'œ': 'o',
    'aͤ': 'ä',
    'oͤ': 'ö',
    'uͤ': 'ü',
    'Aͤ': 'Ä',
    'Oͤ': 'Ö',
    'Uͤ': 'Ü',
    '': null,
    '⁰': '0',
    '¹': '1',
    '²': '2',
    '³': '3',
    '⁴': '4',
    '⁵': '5',
    '⁶': '6',
    '⁷': '7',
    '⁸': '8',
    '⁹': '9',
    '⸗': null,
    '—': '-',
    '‹': null,
    '›': null,
    '»': '>',
    '«': '<',
    '„': '"',
    '”': "'",
    '£': null,
    '§': null,
    "†": "+"
};
Object.assign(
  SPECIAL_CHARACTERS,
  ['½', '¼', '¾', '⅓', '⅔', '⅕', '⅖', '⅗', '⅘', '⅙', '⅐', '⅚', '⅛', '⅜', '⅝',
   '⅞', '⅑', '⅒',
   'Α', 'Δ', 'Κ', 'Π', 'Σ', 'ά', 'έ', 'ή', 'ί', 'α', 'β', 'γ', 'δ', 'ε', 'ζ',
   'η', 'θ', 'ι', 'κ', 'λ', 'μ', 'ν', 'ξ', 'ο', 'π', 'ρ', 'ς', 'σ', 'τ', 'υ',
   'φ', 'χ', 'ψ', 'ω', 'ό', 'ύ', 'ώ', 'ϑ', 'ϰ', 'ϱ'].reduce(
     ((o, c) => {o[c] = null; return o;}), {}));

let KEY_COMBINATIONS = {}
Object.entries(SPECIAL_CHARACTERS)
      .map(([grapheme, key]) => {
        if (key != null) {
          KEY_COMBINATIONS[grapheme] = 'alt+' + key;
          Mousetrap.bind('alt+' + key, (() => {
            bus.$emit('insert-grapheme', grapheme);
          }))}}); // Look Ma', we're in Lisp land!

export default {
  name: 'Keyboard',
  computed: {
    allChars: () => {
      return Object.entries(SPECIAL_CHARACTERS)
                   .map(([char, key]) => ({char, key}));
    }
  },
  methods: {
    onCharClick: function(grapheme) {
      bus.$emit('insert-grapheme', grapheme);
    }
  }
}
</script>

<style scoped>
.character-picker {
  text-align: center
}

button.keycap {
  display:inline-block;
  width:48px;
  height:48px;
  margin: 4px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0px 1px 3px 1px rgba(0, 0, 0, 0.5);
  font: 18px/48px sans-serif ;
  text-align: center;
  color: #666;
}
</style>
