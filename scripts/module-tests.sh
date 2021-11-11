#!/usr/bin/env bash

# this script is used by Github CI to tranverse all modules an run module tests.
# the script expects a diff to be generated in order to skip some modules.

# Executes go module tests and merges the coverage profile.
# If GIT_DIFF variable is set then it's used to test if a module has any file changes - if
# it doesn't have any file changes then we will ignore the module tests.
execute_mod_tests() {
    go_mod=$1;
    mod_dir=$(dirname "$go_mod");
    mod_dir=${mod_dir:2}; # remove "./" prefix
    root_dir=$(pwd);

    # TODO: in the future we will need to disable it once we go into multi module setup, because
    # we will have cross module dependencies.
    if [ -n "$GIT_DIFF" ] && ! grep $mod_dir <<< $GIT_DIFF; then
        echo ">>> ignoring module $mod_dir - no changes in the module";
        return;
    fi;

    echo ">>> running $go_mod tests"
    cd $mod_dir;
    go test -mod=readonly -timeout 30m -coverprofile=${root_dir}/${coverage_file}.tmp -covermode=atomic -tags='norace ledger test_ledger_mock'  ./...
    local ret=$?
    echo "test return: " $ret;
    cd -;
    # strip mode statement
    tail -n +1 ${coverage_file}.tmp >> ${coverage_file}
    rm ${coverage_file}.tmp;
    return $ret;
}

# GIT_DIFF=`git status --porcelain`

echo "GIT_DIFF: " $GIT_DIFF

coverage_file=coverage-go-submod-profile.out
return_val=0;

for f in $(find -name go.mod -not -path "./go.mod"); do
    execute_mod_tests $f;
    if [[ $? -ne 0  ]] ; then
        return_val=2;
    fi;
done

exit $return_val;
