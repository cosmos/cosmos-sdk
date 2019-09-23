<template lang="pug">
  div
    .container
      div
        router-link(to="/")
          tm-logo
        tm-search
        .title Reference
        .section(v-for="(contents, section) in value")
          .section__title(:class="[`section__${sectionActive(contents) ? 'active' : 'inactive'}`]")
            router-link(:to="sectionUrl(contents, section) || '.'") {{section}}
          div(v-if="sectionActive(contents)")
            router-link(:to="child.path" tag="div" v-for="child in contents" v-if="child.title" :class="{'section__child__active': $page.key == child.key}").link.link__child.section__child {{child.title}}
      .footer
        a(href="https://cosmos.network").footer__item
          svg(width="8" height="14" viewBox="0 0 8 14" fill="none" xmlns="http://www.w3.org/2000/svg").footer__item__icon
            path(d="M7 1.5L1.5 7L7 12.5" stroke="#161931" stroke-width="1.5" stroke-linecap="round")
          .footer__item__text Back to Cosmos
</template>

<style lang="stylus" scoped>

.container
  padding 2rem
  height 100%
  overflow-y scroll
  position relative
  display flex
  flex-direction column
  justify-content space-between
  align-items flex-start

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

  &__child
    margin-top 0.5rem
    margin-bottom 0.5rem
    cursor pointer
    position relative
    padding-left 1.5rem

    &__active
      font-weight 500

      &:before
        content ''
        position absolute
        top 0.25rem
        left 0
        height 1rem
        width 1rem
        background url('/bullet-hex-blue.svg') no-repeat top left

  &__title
    font-weight 500
    position relative
    padding-left 1.5rem

  &__active
    &:before
      content ''
      position absolute
      top 0.55rem
      left 0
      height 1rem
      width 1rem
      background url('/bullet-dash.svg') no-repeat top left

  &__inactive
    &:before
      content ''
      position absolute
      top 0.25rem
      left 0
      height 1rem
      width 1rem
      background url('/bullet-hex-full.svg') no-repeat top left

.footer
  margin-top 1.5rem
  background-color var(--sidebar-bg)

  &__item
    color #161931
    text-transform uppercase
    font-size 0.875rem
    display flex
    align-items center
    box-shadow inset 0 0 0 2px rgba(140, 145, 177, 0.32)
    padding 0.75rem 1rem
    border-radius 0.25rem
    font-weight 500

    &__icon
      margin-right 1rem
</style>

<script>
import { includes } from "lodash";

export default {
  props: ["value"],
  methods: {
    sectionActive(section) {
      return includes(Object.values(section).map(e => e.key), this.$page.key);
    },
    sectionUrl(section, name) {
      const search = name => {
        return Object.keys(section).find(
          key => key.toLowerCase() === name.toLowerCase()
        );
      };
      const res =
        section[
          search("readme.md") || search("index.md") || Object.keys(section)[0]
        ];
      console.log(res);
      return res;
    }
  }
};
</script>