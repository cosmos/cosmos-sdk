printf "\n-------------------\n $ make build\n"
make build || exit 1

printf "\n-------------------\n $ ./build/simd config\n"
./build/simd config

printf "\n-------------------\n $ ./build/simd config home $HOME/.tmp-simd\n"
./build/simd config home $HOME/.tmp-simd

printf "\n-------------------\n $ ./build/simd config keyring-backend test\n"
./build/simd config keyring-backend test

printf "\n-------------------\n $ ./build/simd config node mynodeurl:26657\n"
./build/simd config node "mynodeurl:26657"

printf "\n-------------------\n $ ./build/simd config\n"
./build/simd config

printf "\n-------------------\n $ ./build/simd config home $HOME/.simapp\n"
./build/simd config home $HOME/.simapp

printf "\n-------------------\n $ ./build/simd config\n"
./build/simd config
