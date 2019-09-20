import pageComponents from "@internal/page-components";
import TmTooltip from "tm-tooltip";

export default ({ Vue }) => {
  Vue.component("def", TmTooltip);
  for (const [name, component] of Object.entries(pageComponents)) {
    Vue.component(name, component);
  }
};
