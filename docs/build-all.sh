#!/bin/sh

# This builds the docs.cosmos.network docs using docusaurus.
# Old documentation, which have not been migrated to docusaurus are generated with vuepress.
COMMIT=$(git rev-parse HEAD)
mkdir -p ~/versioned_docs  ~/versioned_sidebars
for version in $(jq -r .[] versions.json); do
    echo "building docusaurus $version docs"
    git clean -fdx && git reset --hard && git checkout release/$version.x
    ./pre.sh
    npm ci && npm run docusaurus docs:version $version
    mv ./versioned_docs/* ~/versioned_docs/
    mv ./versioned_sidebars/* ~/versioned_sidebars/
done
echo "building docusaurus main docs"
(git clean -fdx && git reset --hard && git checkout $COMMIT)
mv ~/versioned_docs ~/versioned_sidebars .
npm ci && npm run build
mv build ~/output
echo "adding swagger docs"
cp -r ../client/docs/swagger-ui ~/output/swagger
while read -r branch path_prefix; do
    echo "building vuepress $branch docs"
    (git clean -fdx && git reset --hard && git checkout $branch && npm install && VUEPRESS_BASE="/$path_prefix/" npm run build)
    mkdir -p ~/output/$path_prefix
    cp -r .vuepress/dist/* ~/output/$path_prefix/
done < vuepress_versions
echo "setup domain"
echo $DOCS_DOMAIN > ~/output/CNAME