// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import Buefy from 'buefy'
import Raven from 'raven-js'
import RavenVue from 'raven-js/plugins/vue'
import 'buefy/lib/buefy.css'
import App from './App'
import store from './store'

if (process.env.NODE_ENV === 'production') {
  // Enable sentry.io integration for production deployment
  Raven
    .config('https://f9a9cba1ea2348ae8f5886c75581a9b9@sentry.io/250597')
    .addPlugin(RavenVue, Vue)
    .install()
}

Vue.use(Buefy)

Vue.config.productionTip = false

/* eslint-disable no-new */
new Vue({
  el: '#app',
  template: '<App/>',
  store,
  components: { App }
})
