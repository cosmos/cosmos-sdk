#!/bin/bash
set -euo pipefail

BRANCH=$1
CIRCLE_TOKEN=$2

# let the website CI have a chance to run through the builds
#sleep 120

recent_build=$(curl https://circleci.com/api/v1.1/project/github/cosmos/cosmos.network/tree/$BRANCH?circle-token=$CIRCLE_TOKEN)

# check the last three builds for that branch
last_three=()
for i in {1..3}
do
	last_three+=( $(echo $recent_build | jq '.['"$i"'].outcome') )
done

# zero by default, over-ridden if any of the three builds fail
exit_code=0
i=0

for build in ${last_three[@]}; do
	if [ $build == "\"success\"" ]; then
		echo "Docs build OK"
	else
		failed=$(echo $recent_build | jq '.['"$i"'].build_url')
		echo "Docs build Failed, see $failed"
		exit_code=1
	fi
	i+=1
done

exit $exit_code
