<template lang="pug">
  div
    .container
      .sidebar__container(:class="{sidebarVisible}" @click.self="sidebarVisible = false")
        tm-sidebar(:class="{sidebarVisible}").sidebar.sidebar__hidden
      .content(:class="{sidebarVisible}")
        .topbar
          svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" @click="sidebarVisible = !sidebarVisible").topbar__menu__button
            path(d="M24 18v1h-24v-1h24zm0-6v1h-24v-1h24zm0-6v1h-24v-1h24z")
              path(d="M24 19h-24v-1h24v1zm0-6h-24v-1h24v1zm0-6h-24v-1h24v1z")
          tm-select-language
        tm-content(:sidebar="sidebar")
          template(v-slot:content)
            slot(name="content")
        tm-footer.footer
      .aside__container(v-if="sidebar")
        .aside
          tm-aside
</template>

<style lang="stylus" scoped>
.footer
  z-index 1000
  position relative

.sidebar__container
  overflow-y scroll
  z-index 10000
  pointer-events none
  transition background-color 0.5s

.sidebar
  width 100%
  max-width var(--sidebar-width)
  background-color #f8f9fc
  position absolute
  left 0
  top 0
  bottom 0
  overflow-y scroll
  overflow-x hidden
  z-index 1000
  transition transform 0.5s
  pointer-events all

.content
  margin-left var(--sidebar-width)
  position absolute
  right 0
  left 0
  top 0
  bottom 0
  overflow-y scroll

.topbar
  margin-left 2rem
  margin-right 2rem
  margin-bottom 2rem
  display flex
  height var(--topbar-height)
  justify-content flex-end
  align-items center

  &__menu
    &__button
      margin-right auto
      display none

.aside
  margin-top var(--topbar-height)
  pointer-events initial
  background white
  padding-left 1rem
  padding-right 1rem

  &__container
    pointer-events none
    width 100%
    max-width var(--sidebar-width)
    position absolute
    right 0
    top 0
    bottom 0
    overflow-y scroll
    overflow-x hidden
    z-index 500

@media screen and (max-width: 1024px)
  .aside
    &__container
      display none

@media screen and (max-width: 768px)
  .topbar
    &__menu
      &__button
        display block

  .content
    margin-left 0

  .content.sidebarVisible
    overflow-y hidden

  .sidebar__container
    width 100%
    left 0
    right 0
    top 0
    bottom 0
    position fixed
    overflow-y scroll

    &.sidebarVisible
      background rgba(0, 0, 0, 0.2)
      pointer-events all
      cursor pointer

  .sidebar
    cursor inherit

  .sidebar__hidden
    transform translateX(-100%)

  .sidebarVisible
    transform translateX(0)
</style>

<script>
export default {
  data: function() {
    return {
      sidebarVisible: null
    };
  },
  props: {
    sidebar: {
      type: Boolean,
      default: true
    }
  }
};
</script>