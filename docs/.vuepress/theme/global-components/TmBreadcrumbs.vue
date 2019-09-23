<template lang="pug">
  div
    router-link(:to="item.path" v-for="item in breadcrumbs").item {{item.title}}
</template>

<style lang="stylus" scoped>
.item
  display inline-block
  &:after
    content "/"
    padding-left .5rem
    padding-right .5rem

  &:last-child
    opacity .5

    &:after
      content ""
</style>

<script>
import { find, without } from "lodash";

export default {
  computed: {
    breadcrumbs() {
      let crumbs = this.$page.path
        .split("/")
        .map((currentValue, index, array) => {
          let path = array.slice(0, index + 1).join("/");
          return path;
        });
      crumbs = without(crumbs, "");
      crumbs = crumbs.map(path => {
        const result = find(this.$site.pages, page => {
          return page.path.match(new RegExp(`^${path}(/.*)?$`));
        });
        return result;
      });
      return crumbs;
    }
  }
};
</script>