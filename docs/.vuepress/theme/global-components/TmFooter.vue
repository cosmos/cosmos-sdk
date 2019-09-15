<template lang="pug">
  div
    .wrapper
      .container
        .links
          .links__item
            .links__item__title Documentation
            a(href="https://tendermint.com/docs/" target="_blank" rel="noopenner noreferrer").links__item__link Tendermint Core Docs
            a(href="https://cosmos.network/docs/spec/ibc/" target="_blank" rel="noopenner noreferrer").links__item__link IBC Docs
            a(href="https://hub.cosmos.network" target="_blank" rel="noopenner noreferrer").links__item__link Cosmos Hub Docs
          .links__item
            .links__item__title Community
            .links__item__link Cosmos Community
            .links__item__link Cosmos Blog
            .links__item__link Cosmos Forum
          .links__item
            .links__item__title Contributing
            .links__item__link Contributing to the docs
            .links__item__link Source code on GitHub
            .links__item__link Careers at Tendermint
        .logo
          img(src="/logo-cosmos.svg" alt="Cosmos Logo")
</template>

<style lang="stylus" scoped>
.links
  display grid
  grid-template-columns repeat(auto-fit, minmax(200px, 1fr))

  &__item
    margin-bottom 2rem
    display flex
    flex-direction column

    &__title
      font-size 0.75rem
      letter-spacing 0.2em
      text-transform uppercase
      font-weight 600
      margin-bottom 1rem

    &__link
      font-size 0.875rem
      letter-spacing 0.01em
      line-height 20px
      margin-top 0.5rem
      margin-bottom 0.5rem

.container
  background-color #161931
  color white
  padding-top 5rem
  padding-bottom 5rem
  padding-left 3rem
  padding-right 3rem

.wrapper
  --height 50px
  margin-top var(--height)

  &:before
    content ''
    height var(--height)
    background linear-gradient(transparent, white)
    position absolute
    top calc(-1 * var(--height))
    width 100%
</style>