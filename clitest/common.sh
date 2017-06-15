# this is not executable, but helper functions for the other scripts

prepareClient() {
  echo "Preparing client keys..."
  basecli reset_all
  assertTrue $?

  for i in "${!ACCOUNTS[@]}"; do
      newKey ${ACCOUNTS[$i]}
  done
}

# initServer takes two args - (root dir, chain_id)
# and optionally port prefix as a third arg (default is 4665{6,7,8})
# it grabs the first account to give it the money
initServer() {
  echo "Setting up genesis..."
  SERVE_DIR=$1/server
  assertNotNull "no chain" $2
  CHAIN=$2
  SERVER_LOG=$1/basecoin.log
  basecoin init --home=$SERVE_DIR >>$SERVER_LOG

  #change the genesis to the first account
  GENKEY=$(basecli keys get ${ACCOUNTS[0]} -o json | jq .pubkey.data)
  GENJSON=$(cat $SERVE_DIR/genesis.json)
  echo $GENJSON | jq '.app_options.accounts[0].pub_key.data='$GENKEY \
    | jq ".chain_id=\"$2\"" > $SERVE_DIR/genesis.json

  # optionally set the port
  if [ -n "$3" ]; then
    echo "setting port $3"
    sed -ie "s/4665/$3/" $SERVE_DIR/config.toml
  fi

  echo "Starting server..."
  basecoin start --home=$SERVE_DIR >>$SERVER_LOG 2>&1 &
  sleep 5
  PID_SERVER=$!
}

# initClient requires chain_id arg, port is optional (default 4665{5,6,7})
initClient() {
  echo "Attaching client..."
  PORT=${2:-4665}
  # hard-code the expected validator hash
  basecli init --chain-id=$1 --node=tcp://localhost:${PORT}7 --valhash=EB168E17E45BAEB194D4C79067FFECF345C64DE6
  assertTrue "initialized light-client" $?
}

# newKeys makes a key for a given username, second arg optional password
newKey(){
  assertNotNull "keyname required" "$1"
  KEYPASS=${2:-qwertyuiop}
  (echo $KEYPASS; echo $KEYPASS) | basecli keys new $1 >/dev/null 2>/dev/null
  assertTrue "created $1" $?
  assertTrue "$1 doesn't exist" "basecli keys get $1"
}

# getAddr gets the address for a key name
getAddr() {
  assertNotNull "keyname required" "$1"
  RAW=$(basecli keys get $1)
  assertTrue "no key for $1" $?
  # print the addr
  echo $RAW | cut -d' ' -f2
}
