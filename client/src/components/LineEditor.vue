<template>
  <form class="transcription-widget" @submit="onNext">
    <a class="toggle-ctx previous" v-if="line.previous && !showPrevious"
        title="Mehr Kontext" @click="togglePrevious">
        <b-icon icon="more_horiz" />
      </a>
    <line-image v-if="showPrevious" @click="togglePrevious"
                type="previous" :image-src="line.previous" />
    <line-image type="focus" :image-src="line.line" />
    <line-image v-if="showNext" type="next" :image-src="line.next"
                @click="toggleNext" />
    <a class="toggle-ctx next" v-if="line.next && !showNext"
        title="Mehr Kontext" @click="toggleNext">
        <b-icon icon="more_horiz" />
    </a>
    <b-field>
      <div class="control is-clearfix is-expanded">
        <input ref="transcription" class="input mousetrap" v-model="transcription" />
      </div>
      <a :class="prevClasses" @click="onPrevious" title="Vorherige Zeile korrigieren">
        <b-icon icon="undo" />
      </a>
      <button class="button is-success done-btn" @click="onNext"
              title="Nächste Zeile" type="submit">
        <b-icon icon="done" />
      </button>
      <a class="button is-danger delete-btn" @click="onMarkGarbage"
          title="Zeile verwerfen">
        <b-icon icon="delete" />
      </a>
    </b-field>
  </form>
</template>

<script>
import bus from '../eventBus'
import LineImage from './LineImage'

export default {
  name: 'LineEditor',
  props: ['line', 'hasPrevious'],
  components: {LineImage},
  data () {
    return {
      transcription: this.line.transcription,
      showPrevious: false,
      showNext: false
    };
  },
  computed: {
    prevClasses () {
      return {
        'button': true,
        'undo-btn': true,
        'disabled': !this.hasPrevious
      };
    }
  },
  methods: {
    onNext (e) {
      e.preventDefault();
      this.$emit('transcription-done', this.transcription);
    },
    onPrevious () {
      if (this.hasPrevious) {
        this.$emit('previous-line');
      }
    },
    onMarkGarbage () {
      this.$emit('delete-line');
    },
    togglePrevious () {
      this.showPrevious = !this.showPrevious;
    },
    toggleNext () {
      this.showNext = !this.showNext;
    }
  },
  created () {
    var vm = this;
    bus.$on('insert-grapheme', function(grapheme) {
      if (vm.$refs.transcription) {
        let ref = vm.$refs.transcription;
        vm.transcription = vm.transcription.substring(0, ref.selectionStart) +
                          grapheme + vm.transcription.substring(ref.selectionEnd);
        ref.focus();
      }
    });
  },
  mounted () {
    this.$refs.transcription.focus();
  }
}
</script>

<style scoped>
a.toggle-ctx {
  display: block;
  width: 100%;
  text-align: center;
  color: #000000;
}

.toggle-ctx.previous {
  margin-bottom: 0.1em;
}

.toggle-ctx.next {
  margin-top: 0.1em;
}

button.done-btn {
  width: 8vh;
}

button.skip-btn {
  width: 6vh;
}

button.delete-btn {
  width: 4vh;
}
</style>
