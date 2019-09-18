<template lang="pug">
  div
    .container
      div
        router-link(to="/")
          tm-logo
        tm-search
        .title Reference
        .section(v-for="section in value")
          router-link(:to="section.regularPath || ''" tag="div").section__title.link {{section.title}}
          .children
            router-link(:to="link.regularPath" tag="div" v-for="link in section.children").link.link__child {{link.title}}
      .footer
        a(href="https://cosmos.network").footer__item
          svg(width="8" height="14" viewBox="0 0 8 14" fill="none" xmlns="http://www.w3.org/2000/svg").footer__item__icon
            path(d="M7 1.5L1.5 7L7 12.5" stroke="#161931" stroke-width="1.5" stroke-linecap="round")
          .footer__item__text Back to Cosmos
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

  &__title
    font-weight 500
    position relative

    &:before
      content ''
      position absolute
      top 0.25rem
      left 0
      height 1rem
      width 1rem
      background url('/bullet-hex-full.svg') no-repeat top left

    &.router-link-active
      &:before
        content ''
        position absolute
        top 0.55rem
        left 0
        height 1rem
        width 1rem
        background url('/bullet-dash.svg') no-repeat top left

.link
  margin-top 0.5rem
  margin-bottom 0.5rem
  cursor pointer
  padding-left 1.5rem

  &__child
    position relative

    &.router-link-exact-active
      &:before
        content ''
        position absolute
        top 0.25rem
        left 0
        height 1rem
        width 1rem
        background url('/bullet-hex-blue.svg') no-repeat top left

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
export default {
  props: {
    value: {
      type: Array,
      default: () => []
    }
  }
};
</script>