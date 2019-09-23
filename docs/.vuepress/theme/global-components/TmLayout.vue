<template lang="pug">
  div
    .container
      .sidebar__container(:class="{sidebarVisible}" @click.self="sidebarVisible = false")
        tm-sidebar(:class="{sidebarVisible}" :value="tree").sidebar.sidebar__hidden
      .content(:class="{sidebarVisible}")
        .topbar
          svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" @click="sidebarVisible = !sidebarVisible").topbar__menu__button
            path(d="M24 18v1h-24v-1h24zm0-6v1h-24v-1h24zm0-6v1h-24v-1h24z")
              path(d="M24 19h-24v-1h24v1zm0-6h-24v-1h24v1zm0-6h-24v-1h24v1z")
          tm-breadcrumbs
          tm-select-language.topbar__language
        tm-content(:aside="aside")
          template(v-slot:content)
            slot(name="content")
        tm-footer.footer
      .aside__container(v-if="aside")
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
  background-color var(--sidebar-bg)
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
  margin-left 4rem
  margin-right 1rem
  display flex
  height var(--topbar-height)
  justify-content space-between
  align-items center

  &__breadcrumbs
    font-size 0.75rem
    text-transform uppercase
    letter-spacing 0.2em

  &__language
    margin-left auto

  &__menu
    &__button
      margin-right 1rem
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
  .topbar
    margin-left 2rem
    margin-right 2rem

  .aside
    &__container
      display none

@media screen and (max-width: 768px)
  .topbar
    margin-left 2rem

    &__breadcrumbs
      &__current
        display none

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
import { find, filter, forEach, remove, last, omit, omitBy } from "lodash";

export default {
  data: function() {
    return {
      sidebarVisible: null
    };
  },
  computed: {
    tree() {
      const langDirs = Object.keys(this.$site.locales).map(e =>
        e.replace(/\//g, "")
      );
      const files = this.$site.pages;
      const langCurrent = this.$localeConfig.path.replace(/\//g, "");
      const langOther = langCurrent.length > 0;
      let tree = {};
      files.forEach(file => {
        let location = file.relativePath.split("/");
        if (location.length === 1) {
          return (tree[location[0]] = file);
        }
        location.reduce((prevDir, currDir, i, filePath) => {
          if (i === filePath.length - 1) {
            prevDir[currDir] = file;
          }
          if (!prevDir.hasOwnProperty(currDir)) {
            prevDir[currDir] = {};
          }
          return prevDir[currDir];
        }, tree);
      });
      tree = langOther ? tree[langCurrent] : omit(tree, langDirs);
      tree = omitBy(tree, e => typeof e.key === "string");
      return tree;
    }
  },
  props: {
    aside: {
      type: Boolean,
      default: true
    }
  }
};
</script>