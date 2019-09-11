import axios from "axios";
import Vue from "vue";
import pageComponents from "@internal/page-components";

Vue.use({
  install(Vue) {
    Vue.prototype.$axios = axios.create();
  }
});

export default ({ Vue }) => {
  for (const [name, component] of Object.entries(pageComponents)) {
    Vue.component(name, component);
  }
};
