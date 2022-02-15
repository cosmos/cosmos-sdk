export default ({ router }) => {
  router.addRoutes([
    { path: '/master/spec/*', redirect: '/master/modules/' },
    { path: '/master/spec/governance/', redirect: '/master/modules/gov/' },
    { path: '/v0.41/', redirect: '/v0.42/' },
    { path: '/v0.43/', redirect: '/v0.44/' },
  ])
}
