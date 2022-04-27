export default ({ router }) => {
  router.addRoutes([
    { path: "/main/spec/*", redirect: "/modules/" },
    { path: "/main/spec/governance/", redirect: "/modules/gov/" },
    { path: "/v0.43/", redirect: "/v0.44/" }, // TODO to fix: https://github.com/cosmos/cosmos-sdk/issues/11798
    { path: "/master/", redirect: "/" },
  ]);
};
