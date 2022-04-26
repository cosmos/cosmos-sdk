export default ({ router }) => {
  router.addRoutes([
    { path: '/main/spec/*', redirect: '/modules/' },
    { path: '/main/spec/governance/', redirect: '/modules/gov/' },
    { path: '/v0.41/', redirect: '/v0.42/' }, // TODO to fix
    { path: '/v0.43/', redirect: '/v0.44/' }, // TODO to fix
    { path: '/master/', redirect: '/' },
  ])
}
