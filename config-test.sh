BIN=./build/simd

printf "\n-------------------\nBuilding binary\n"
printf " $ make build\n"
make build || exit 1

echo "Running version:"
$BIN version

printf "\n-----------------\nShowing the current home directory\n"
printf " $ ./build/simd config home\n"
echo "Showing home directory"
$BIN config home

printf "\n-----------------\nShow node in original home\n"
printf " $ ./build/simd config client get node\n"
$BIN config get client node

printf "\n-----------------\nSetting a new home directory (1st option)\n"
printf " $ ./build/simd config home $HOME/.tmp-simd\n"
$BIN config home $HOME/.tmp-simd

printf "\n-----------------\nSetting a new home directory (2nd option)\n"
printf " $ ./build/simd config set home $HOME/.tmp-simd\n"
$BIN config home $HOME/.tmp-simd

printf "\n-----------------\nAdjusting chain-id in new home\n"
printf " $ ./build/simd config set client chain-id test\n"
$BIN config set client chain-id test
#
printf "\n-----------------\nAdjusting node in new home\n"
printf " $ ./build/simd config set client node mynodeurl:26657\n"
$BIN config set client node "mynodeurl:26657"

printf "\n-----------------\nShow node in new home\n"
printf " $ ./build/simd config get client node\n"
$BIN config get client node

printf "\n-----------------\nSet home back to original path\n"
printf " $ ./build/simd config home $HOME/.simapp\n"
$BIN config home $HOME/.simapp

printf "\n-----------------\nGet current home\n"
printf " $ ./build/simd config home\n"
$BIN config home

printf "\n-----------------\nShow node in original home\n"
printf " $ ./build/simd config client get node\n"
$BIN config get client node
