<template>
  <div class="column">
    <b-loading :active="isLoading" />
    <div v-if="!isLoading" class="search">
      <b-input placeholder="Filter..." type="search" icon="search"
              v-model="searchQuery" />
      <p v-if="filters">
        <span class="tag is-success" v-for="(val, filter) in filters" key="filter">
          {{ filter }}: {{ val }}
          <button class="delete is-small" @click="clearFilter(filter)"></button>
        </span>
      </p>
    </div>
    <b-table
      v-if="!isLoading"
      :data="displayedDocuments"
      default-sort-direction="desc"
      default-sort="authoredDate">
      <template slot-scope="props">
        <b-table-column label="Transkribiert" field="authoredDate" sortable>
          <span :title="props.row.authoredDate">
            {{ relativeTime(props.row.authoredDate) }}
          </span><br/>
          von <a title="Nach diesem Autor filtern" @click="addFilter('author', props.row.author)">
            {{ props.row.author }}
          </a>
        </b-table-column>
        <b-table-column label="Letztes Review" field="lastReviewDate" width="130" sortable>
          <span v-if="props.row.lastReviewDate">
            <span :title="props.row.lastReviewDate">
              {{ relativeTime(props.row.lastReviewDate) }}
            </span><br/>
            von <a title="Nach diesem Reviewer filtern" @click="addFilter('reviewer', props.row.reviewer)">
            {{ props.row.reviewer }}
            </a>
          </span>
        </b-table-column>
        <b-table-column>
          <a class="button is-info" @click="startReview(props.row.id)" title="Review starten">
            <b-icon icon="message-draw"/>
          </a>
        </b-table-column>
        <b-table-column label="Titel" field="title" sortable>
          {{ props.row.title }}
        </b-table-column>
        <b-table-column label="Jahr" field="year" numeric sortable>
          {{ props.row.year }}
        </b-table-column>
        <b-table-column label="Zeilenzahl" field="numLines" numeric sortable>
          {{ props.row.numLines }}
        </b-table-column>
      </template>
    </b-table>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import { distanceInWordsToNow } from 'date-fns'
import de from 'date-fns/locale/de'

export default {
  name: 'DocumentList',
  data () {
    return {
      filters: {},
      searchQuery: undefined
    }
  },
  computed: {
    isLoading () {
      return this.allDocuments.length === 0
    },
    displayedDocuments () {
      let out = this.allDocuments.map(doc => {
        let viewDoc = {
          lastModification: new Date(doc.history[0].date),
          // TODO: Use regex instead of startsWith
          author: doc.history.filter(h => h.subject.startsWith('Transcribed '))[0].author.name,
          authoredDate: doc.history.filter(h => h.subject.startsWith('Transcribed '))[0].date,
          ...doc
        }
        let reviews = doc.history.filter(h => h.subject.startsWith('Reviewed '))
        if (reviews.length > 0) {
          viewDoc.lastReviewDate = reviews[0].date
          viewDoc.reviewer = reviews[0].author.name
        }
        return viewDoc
      })
      if (this.searchQuery || this.filters) {
        return out.filter(this.docMatchesQueryAndFilters)
      } else {
        return out
      }
    },
    ...mapState(['allDocuments'])
  },
  methods: {
    addFilter (field, val) {
      this.$set(this.filters, field, val)
    },
    clearFilter (field) {
      this.$delete(this.filters, field)
    },
    startReview (ident) {
      let doc = this.allDocuments.filter(doc => doc.id === ident)[0]
      this.$store.dispatch('fetchDocumentLines', ident)
      this.$store.commit('setActiveDocument', doc)
      this.$store.commit('changeScreen', 'multi')
    },
    matchesQuery (val) {
      return val.toLowerCase().indexOf(this.searchQuery.toLowerCase()) >= 0
    },
    docMatchesQueryAndFilters (doc) {
      let match = !this.searchQuery || (
        this.matchesQuery(doc.title) ||
        this.matchesQuery(doc.author) ||
        this.matchesQuery(doc.id))
      if (!match) {
        return false
      }
      for (let field in this.filters) {
        if (doc[field] !== this.filters[field]) {
          return false
        }
      }
      return true
    },
    relativeTime (timestamp) {
      let date = Date.parse(timestamp)
      return distanceInWordsToNow(date, { locale: de, addSuffix: true })
    },
    ...mapActions(['fetchDocuments'])
  },
  mounted () {
    this.fetchDocuments()
  }
}
</script>

<style scoped>
.input[type="search"] {
  width: 16rem;
}
</style>
