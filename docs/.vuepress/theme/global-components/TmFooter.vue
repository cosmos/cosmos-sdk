<template lang="pug">
  div
    .wrapper
      tm-help-support
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
          img(src="/logo-cosmos.svg" alt="Cosmos Logo").logo__item.logo__image
          .logo__item.logo__link
            a(href="https://cosmos.network" target="_blank" rel="noreferrer noopener") cosmos.network
        .smallprint
          .smallprint__item.smallprint__item__links
            a(href="https://blog.cosmos.network/" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M24 24h-24v-24h24v24zm-4.03-5.649v-.269l-1.247-1.224c-.11-.084-.165-.222-.142-.359v-8.998c-.023-.137.032-.275.142-.359l1.277-1.224v-.269h-4.422l-3.152 7.863-3.586-7.863h-4.638v.269l1.494 1.799c.146.133.221.327.201.523v7.072c.044.255-.037.516-.216.702l-1.681 2.038v.269h4.766v-.269l-1.681-2.038c-.181-.186-.266-.445-.232-.702v-6.116l4.183 9.125h.486l3.593-9.125v7.273c0 .194 0 .232-.127.359l-1.292 1.254v.269h6.274z")
            a(href="https://twitter.com/cosmos" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z")
            a(href="https://www.linkedin.com/company/tendermint/" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M0 0v24h24v-24h-24zm8 19h-3v-11h3v11zm-1.5-12.268c-.966 0-1.75-.79-1.75-1.764s.784-1.764 1.75-1.764 1.75.79 1.75 1.764-.783 1.764-1.75 1.764zm13.5 12.268h-3v-5.604c0-3.368-4-3.113-4 0v5.604h-3v-11h3v1.765c1.397-2.586 7-2.777 7 2.476v6.759z")
            a(href="https://reddit.com/r/cosmosnetwork" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M14.238 15.348c.085.084.085.221 0 .306-.465.462-1.194.687-2.231.687l-.008-.002-.008.002c-1.036 0-1.766-.225-2.231-.688-.085-.084-.085-.221 0-.305.084-.084.222-.084.307 0 .379.377 1.008.561 1.924.561l.008.002.008-.002c.915 0 1.544-.184 1.924-.561.085-.084.223-.084.307 0zm-3.44-2.418c0-.507-.414-.919-.922-.919-.509 0-.923.412-.923.919 0 .506.414.918.923.918.508.001.922-.411.922-.918zm13.202-.93c0 6.627-5.373 12-12 12s-12-5.373-12-12 5.373-12 12-12 12 5.373 12 12zm-5-.129c0-.851-.695-1.543-1.55-1.543-.417 0-.795.167-1.074.435-1.056-.695-2.485-1.137-4.066-1.194l.865-2.724 2.343.549-.003.034c0 .696.569 1.262 1.268 1.262.699 0 1.267-.566 1.267-1.262s-.568-1.262-1.267-1.262c-.537 0-.994.335-1.179.804l-2.525-.592c-.11-.027-.223.037-.257.145l-.965 3.038c-1.656.02-3.155.466-4.258 1.181-.277-.255-.644-.415-1.05-.415-.854.001-1.549.693-1.549 1.544 0 .566.311 1.056.768 1.325-.03.164-.05.331-.05.5 0 2.281 2.805 4.137 6.253 4.137s6.253-1.856 6.253-4.137c0-.16-.017-.317-.044-.472.486-.261.82-.766.82-1.353zm-4.872.141c-.509 0-.922.412-.922.919 0 .506.414.918.922.918s.922-.412.922-.918c0-.507-.413-.919-.922-.919z")
            a(href="https://t.me/cosmosproject" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M12,0c-6.626,0 -12,5.372 -12,12c0,6.627 5.374,12 12,12c6.627,0 12,-5.373 12,-12c0,-6.628 -5.373,-12 -12,-12Zm3.224,17.871c0.188,0.133 0.43,0.166 0.646,0.085c0.215,-0.082 0.374,-0.267 0.422,-0.491c0.507,-2.382 1.737,-8.412 2.198,-10.578c0.035,-0.164 -0.023,-0.334 -0.151,-0.443c-0.129,-0.109 -0.307,-0.14 -0.465,-0.082c-2.446,0.906 -9.979,3.732 -13.058,4.871c-0.195,0.073 -0.322,0.26 -0.316,0.467c0.007,0.206 0.146,0.385 0.346,0.445c1.381,0.413 3.193,0.988 3.193,0.988c0,0 0.847,2.558 1.288,3.858c0.056,0.164 0.184,0.292 0.352,0.336c0.169,0.044 0.348,-0.002 0.474,-0.121c0.709,-0.669 1.805,-1.704 1.805,-1.704c0,0 2.084,1.527 3.266,2.369Zm-6.423,-5.062l0.98,3.231l0.218,-2.046c0,0 3.783,-3.413 5.941,-5.358c0.063,-0.057 0.071,-0.153 0.019,-0.22c-0.052,-0.067 -0.148,-0.083 -0.219,-0.037c-2.5,1.596 -6.939,4.43 -6.939,4.43Z")
            a(href="https://www.youtube.com/c/CosmosProject" target="_blank" rel="noreferrer noopener").smallprint__item__links__item
              svg(width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" fill="white")
                path(d="M19.615 3.184c-3.604-.246-11.631-.245-15.23 0-3.897.266-4.356 2.62-4.385 8.816.029 6.185.484 8.549 4.385 8.816 3.6.245 11.626.246 15.23 0 3.897-.266 4.356-2.62 4.385-8.816-.029-6.185-.484-8.549-4.385-8.816zm-10.615 12.816v-8l8 3.993-8 4.007z")
          .smallprint__item__desc.smallprint__item The development of the Cosmos project is led primarily by Tendermint Inc., the for-profit entity which also maintains this website. Funding for this development comes primarily from the Interchain Foundation, a Swiss non-profit.
</template>

<style lang="stylus" scoped>
.links
  display grid
  grid-template-columns repeat(auto-fit, minmax(200px, 1fr))

  &__item
    display flex
    flex-direction column
    margin 1.5rem

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
  padding-top 3.5rem
  padding-bottom 3.5rem
  padding-left 1.5rem
  padding-right 1.5rem

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

.logo
  display grid
  grid-template-columns repeat(auto-fit, minmax(200px, 1fr))
  align-items flex-end

  &__item
    padding 1.5rem

  &__link
    grid-column span 2
    font-weight 500

.smallprint
  display grid
  grid-template-columns repeat(auto-fit, minmax(200px, 1fr))
  align-items flex-end

  &__item
    padding 1.5rem
    fill white

    &__links
      &__item
        margin-right 1rem

        &__icon
          fill white

    &__desc
      grid-column span 2
      font-size 0.8125rem
      color rgba(255, 255, 255, 0.51)
      line-height 1rem
</style>