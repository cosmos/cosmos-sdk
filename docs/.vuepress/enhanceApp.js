export default ({ router }) => {
  router.addRoutes([
    { path: '/master/spec/*', redirect: '/master/modules/' },
    { path: '/master/spec/governance/', redirect: '/master/modules/gov/' },
  ])
}