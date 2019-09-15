<template lang="pug">
  div
    .container
      .content(:class="{noSidebar}")
        .content__container(:class="{noSidebar}")
          slot(name="content")
</template>

<style lang="stylus" scoped>
.container
  position relative
  min-height 100vh

.content
  padding-right var(--sidebar-width)
  width 100%
  position relative

  &.noSidebar
    padding-right 0

  &__container
    width 100%
    max-width 800px
    padding-left 4rem
    padding-right 2rem
    overflow-x hidden

    &.noSidebar
      max-width initial

/deep/
  a
    color var(--color-accent)
    text-decoration none

  h1
    font-size 2.25rem
    font-weight 600
    margin-bottom 2rem

  h2
    font-size 1.5rem
    font-weight 600
    margin-top 2rem
    margin-bottom 1.5rem

  p
    margin-top 1rem
    margin-bottom 1rem
    line-height 1.5

  ul
    line-height 1.5

  li
    padding-left 2rem
    list-style none
    margin-bottom 1rem
    position relative

    &:before
      content ''
      width 1rem
      height 1rem
      background url('/bullet.svg') no-repeat top left
      position absolute
      top 0.35rem
      left 0

@media screen and (max-width: 1024px)
  .content
    padding-right 0

    &__container
      padding-left 2rem
</style>

<script>
export default {
  props: {
    sidebar: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    noSidebar() {
      return !this.sidebar;
    }
  }
};
</script>