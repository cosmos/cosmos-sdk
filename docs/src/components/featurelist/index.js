import React from "react";

export default function FeatureList(url) {
  return [
    {
      title: `Learn Basics`,
      Svg: require("@site/static/img/innovation.svg").default,
      to: `${url}/learn/intro/overview`,
      description: (
        <>
          Get a quick introduction to the Cosmos SDK and its key features,
          including its modular architecture and developer-friendly tools.
        </>
      ),
    },
    {
      title: `Build a Chain`,
      to: `${url}/build/building-apps/app-go`,
      Svg: require("@site/static/img/link.svg").default,
      description: (
        <>
          Learn how to build a customized blockchain application using the Cosmos
          SDK, with support for various programming languages and consensus
          algorithms.
        </>
      ),
    },
    {
      title: `Build a Module`,
      to: `${url}/build/building-modules/intro`,
      Svg: require("@site/static/img/cube.svg").default,
      description: (
        <>
          Dive deeper into the Cosmos SDK and learn how to create custom modules
          to extend the functionality of your blockchain application.
        </>
      ),
    },
    {
      title: `Node Operation`,
      to: `${url}/user/run-node/run-node`,
      Svg: require("@site/static/img/node.svg").default,
      description: (
        <>
          Learn how to set up and operate a full node on the Cosmos network, and
          become an active participant in the governance and decision-making
          processes of the ecosystem.
        </>
      ),
    },
    {
      title: `Join the Community`,
      to: "https://discord.gg/interchain",
      Svg: require("@site/static/img/public-service.svg").default,
      description: (
        <>
          Connect with other developers, validators, and enthusiasts in the Cosmos
          ecosystem, and collaborate on building the future of decentralized
          applications.
        </>
      ),
    },
    {
      title: `Discuss`,
      to: "https://github.com/orgs/cosmos/discussions",
      Svg: require("@site/static/img/ecosystem.svg").default,
      description: (
        <>
          Collaborative forum for the community to ask/answer questions, share
          information, discuss items and give feedbacks on the teams roadmaps.
        </>
      ),
    },
  ];
}
