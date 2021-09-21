export default ({ router }) => {
  router.addRoutes([
    { path: '/v0.41/', redirect: '/v0.42/' },
    { path: '/v0.43/', redirect: '/v0.44/' },
  ])
}
