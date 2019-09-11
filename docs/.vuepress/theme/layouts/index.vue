<template lang="pug">
  div
    tm-layout(:sidebar="false")
      template(v-slot:content)
        .intro
          .h1 {{$frontmatter.title}}
          .p {{$frontmatter.description}}
        .h2 Getting Started
</template>

<style lang="stylus" scoped>
.intro
  width 100%
  max-width 800px

.h1
  font-size 2.25rem
  font-weight 600
  margin-bottom 1.5rem

.h2
  font-size 1.5rem
  font-weight 600
  margin-top 2.5rem

.p
  font-size 1.25rem
  line-height 28px
</style>