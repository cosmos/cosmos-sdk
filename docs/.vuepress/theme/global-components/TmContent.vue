<template lang="pug">
  div
    .container
      .content(:class="{noSidebar}")
        .content__container(:class="{noSidebar}")
          slot(name="content")
      .sidebar(:class="{noSidebar}")
        .sidebar__container
          slot(name="sidebar")
</template>

<style lang="stylus" scoped>
.container
  position relative

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

.sidebar
  width 100%
  max-width var(--sidebar-width)
  position absolute
  right 0
  top 0
  bottom 0

  &.noSidebar
    display none

  &__container
    position fixed
    padding-right 2rem

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
    margin-left 2rem
    list-style none
    margin-bottom 1rem

@media screen and (max-width: 1024px)
  .content
    padding-right 0

    &__container
      padding-left 2rem

  .sidebar
    display none
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