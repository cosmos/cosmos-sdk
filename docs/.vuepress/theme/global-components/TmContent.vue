<template lang="pug">
  div
    .container
      .content(:class="{noAside}")
        .content__container(:class="{noAside}")
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

  &.noAside
    padding-right 0

  &__container
    width 100%
    max-width 800px
    padding-left 4rem
    padding-right 2rem
    overflow-x hidden

    &.noAside
      max-width initial

/deep/
  .term
    text-decoration underline

  .tooltip
    h1
      font-size 0.875rem
      font-weight 500
      margin-bottom 0

    p
      margin-top 0
      margin-bottom 0

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

  @media screen and (max-width: 550px)
    .tooltip
      position fixed
      z-index 1000
      border-radius 0
      left 0
      right 0
      bottom 0

@media screen and (max-width: 1024px)
  .content
    padding-right 0

    &__container
      padding-left 2rem
</style>

<script>
export default {
  props: {
    aside: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    noAside() {
      return !this.aside;
    }
  }
};
</script>