<template lang="pug">
  div
    tm-layout(:sidebar="false")
      template(v-slot:content)
        .intro
          .h1 {{$frontmatter.title}}
          .p {{$frontmatter.description}}
        .h2 Getting Started
        .features
          a(:href="feature.url" v-for="feature in $frontmatter.features" :class="[feature.special]" target="_blank" rel="noreferrer noopener").features__item
            .features__item__cta {{feature.cta}}
            .features__item__title {{feature.title}}
            .features__item__desc {{feature.desc}}
            .features__item__label
              .features__item__label__text {{feature.label}}
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
  margin-bottom 2rem

.p
  font-size 1.25rem
  line-height 28px

.features
  display grid
  grid-template-columns repeat(3, 1fr)
  gap 1.5rem

  &__item
    border 1px solid rgba(140, 145, 177, 0.32)
    padding-top 1.5rem
    padding-left 1.5rem
    padding-right 1.5rem
    border-radius 0.5rem
    position relative
    padding-bottom 4rem
    color inherit

    &__cta
      text-transform uppercase
      font-size 0.75rem
      letter-spacing 0.2em

    &__title
      font-weight 600
      font-size 1.25rem
      line-height 28px
      margin-bottom 0.5rem

    &__desc
      font-size 0.875rem
      line-height 20px

    &__label
      position absolute
      bottom 1.5rem

      &__text
        display inline-block
        background rgba(176, 180, 207, 0.2)
        border-radius 0.25rem
        padding 0.25rem 0.5rem
        font-size 0.8125rem

.dark
  background rgba(22, 25, 49, 0.9)
  color white
</style>