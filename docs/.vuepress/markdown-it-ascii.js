var child_process = require("child_process");

module.exports = function asciiDiagram(md, options) {
  let defaultFenceRenderer = md.renderer.rules.fence;
  function customFenceRenderer(tokens, idx, options, env, slf) {
    let token = tokens[idx];
    let info = token.info.trim();
    let langName = info ? info.split(/\s+/g)[0] : "";
    switch (langName) {
      case "ascii": {
        try {
          var child = child_process.spawnSync(
            process.env.SVGBOB_PATH,
            ["-s", "--", token.content],
            {
              encoding: "utf8"
            }
          );
          return (
            "<img src='data:image/svg+xml;base64," +
            Buffer.from(child.stdout).toString("base64") +
            "'>"
          );
        } catch (e) {
          console.log(e);
        }
        break;
      }
      default: {
        return defaultFenceRenderer(tokens, idx, options, env, slf);
      }
    }
  }
  md.renderer.rules.fence = customFenceRenderer;
};
