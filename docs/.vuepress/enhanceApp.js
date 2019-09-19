import axios from "axios";
import Vue from "vue";
import pageComponents from "@internal/page-components";
import TmTooltip from "tm-tooltip";

Vue.use({
  install(Vue) {
    Vue.prototype.$axios = axios.create();
  }
});

export default ({ Vue }) => {
  Vue.component("def", TmTooltip);
  for (const [name, component] of Object.entries(pageComponents)) {
    Vue.component(name, component);
  }
};
