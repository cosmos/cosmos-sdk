export default ({ router }) => {
  router.addRoutes([
    { path: '/v0.41/', redirect: '/v0.42/' },
  ])
}