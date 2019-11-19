import axios from 'axios'
import Vue from 'vue'

Vue.use({
    install (Vue) {
    Vue.prototype.$axios = axios.create()
  }
})