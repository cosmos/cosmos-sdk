<template lang="pug">
  div
    search-box.search
</template>

<style lang="stylus" scoped>
.search
  z-index 10000

/deep/
  .search-box
    margin-top 1.5rem
    margin-bottom 1rem
    margin-right 0
    position relative

    & input
      width 100%
      background-color rgba(0, 0, 0, 0.02)
      margin 0

      &:focus
        border-color rgba(0, 0, 0, 0.2)
        background-color white

  .suggestions
    position absolute
    left 0
    margin-top 1rem
    font-size 0.75rem
    width 230px
    margin-bottom 2rem
    box-shadow 0 5px 10px 0 rgba(0, 0, 0, 0.05)

  input
    left 0
</style>

<script>
import SearchBox from "@SearchBox";

export default {
  components: { SearchBox },
  mounted() {
    document
      .querySelector(".search-box input")
      .setAttribute("placeholder", "Search");
  }
};
</script>