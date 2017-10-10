<template>
  <h1 class="title">
    <a :href="metadata['identifier-access']">{{ metadata.title }}</a> ({{metadata.year}})
    <b-field grouped class="is-pulled-right">
      <a class="button" @click="showSingle" title="Detailansicht">
        <b-icon icon="search" />
      </a>
      <a class="button" @click="showMulti" title="Zeilenübersicht">
        <b-icon icon="format_align_left" />
      </a>
      <a :class="btnClasses" @click="pickAnother" title="Anderes Buch auswählen">
        <b-icon icon="autorenew" />
      </a>
      <a class="button is-success" @click="submit" title="Transkriptionen abschicken">
        <b-icon icon="send" />
      </a>
    </b-field>
  </h1>
</template>

<script>
import bus from '../eventBus'

export default {
  name: 'VolumeInfo',
  props: ['metadata', 'isLoading'],
  computed: {
    btnClasses () {
      return {
        'button': true,
        'is-loading': this.isLoading,
      };
    }
  },
  methods: {
    pickAnother () {
      this.$emit('pick-another');
    },
    showSingle () {
      bus.$emit('change-screen', 'single');
    },
    showMulti () {
      bus.$emit('change-screen', 'multi');
    },
    submit () {
      bus.$emit('change-screen', 'submit');
    }
  }
}
</script>

<style scope>
</style>
