<template lang="pug">
  div
    .container
      router-link(to="/")
        tm-logo
      tm-search
      .title Reference
      .section(v-for="section in sidebar")
        router-link(:to="section.regularPath || ''" tag="div").section__title.link {{section.title}}
        .children
          router-link(:to="link.regularPath" tag="div" v-for="link in section.children").link {{link.title}}
</template>

<style lang="stylus" scoped>
.router-link-exact-active
  font-weight 500

.section__title + .children
  display none

.section__title.router-link-active + .children
  display block

.container
  padding 2rem

.title
  font-size 0.75rem
  text-transform uppercase
  letter-spacing 0.2em
  color #666
  margin-top 1rem
  margin-bottom 1rem

.section
  font-size 0.875rem
  letter-spacing 0.01em
  line-height 20px
  margin-top 0.75rem
  margin-bottom 0.75rem

  &__title
    font-weight 500
    position relative

    &:before
      content ''
      position absolute
      top 0.15rem
      left 0
      height 1rem
      width 1rem
      background url('/circle.svg') no-repeat top left

.link
  margin-top 0.5rem
  margin-bottom 0.5rem
  cursor pointer
  padding-left 1.5rem
</style>

<script>
import { find } from "lodash";

export default {
  computed: {
    sidebar() {
      return this.$themeLocaleConfig.sidebar.map(section => {
        const children = section.children.map(child => {
          return find(this.$site.pages, {
            relativePath: child.replace("./", "")
          });
        });
        return {
          title: section.title,
          regularPath: children[0] && children[0].regularPath,
          children
        };
      });
    }
  }
};
</script>