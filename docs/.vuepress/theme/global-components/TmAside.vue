<template lang="pug">
  div
    div
      a(href="https://hackathon.sfblockchainweek.io" target="_blank" rel="noreferrer noopener")
        img(src="/banner-hackathon.svg").aside__image
    .aside__title On this page
    .aside__link(v-for="link in $page.headers")
      a(:href="`${$page.regularPath}#${link.slug}`").aside__link__href.header-anchor {{link.title}}
</template>

<style lang="stylus" scoped>
.aside
  &__image
    width 100%

  &__title
    font-size 0.75rem
    text-transform uppercase
    letter-spacing 0.2em
    color #666
    margin-top 3rem

  &__link
    color rgba(22, 25, 49, 0.65)
    margin-top 0.75rem
    margin-bottom 0.75rem
    font-size 0.875rem

    &__href
      color rgba(22, 25, 49, 0.65)
</style>