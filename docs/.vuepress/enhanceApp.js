export default ({ router }) => {
  router.addRoutes([
    { path: '/main/spec/*', redirect: '/main/modules/' },
    { path: '/main/spec/governance/', redirect: '/main/modules/gov/' },
    { path: '/v0.41/', redirect: '/v0.42/' },
    { path: '/v0.43/', redirect: '/v0.44/' },
  ])
}
