#!/usr/bin/env python3

import argparse
import datetime
import json
import math
import sys


# default_date = datetime.datetime(year=2019, month=2, day=11, hour=12, tzinfo=datetime.timezone.utc) # '2019-02-11T12:00:00Z'

parser = argparse.ArgumentParser(description='Convert genesis.json from v0.33.2 to v0.34.0')
parser.add_argument('exported_genesis', type=argparse.FileType('r'), default=sys.stdin)
parser.add_argument('--chain-id', type=str, default='cosmoshub-n')
parser.add_argument('--start-time', type=str, default='2019-02-11T12:00:00Z')
args = parser.parse_args()

chain_id = args.chain_id.strip()

if chain_id == '':
  sys.exit('chain-id required')

raw = args.exported_genesis.read()
genesis = json.loads(raw)

# update genesis with breaking changes
genesis['consensus_params']['block'] = genesis['consensus_params']['block_size']
del genesis['consensus_params']['block_size']
genesis['consensus_params']['block']['time_iota_ms'] = '1000' # default tm value


crisis_data = {
  'constant_fee': {
    'amount': '1333000000',  # $5,000 worth of uatoms
    'denom': 'uatom'
  }
}
genesis['app_state']['crisis'] = crisis_data

# proposal #1 updates
genesis['app_state']['mint']['params']['blocks_per_year'] = '4855015'

# proposal #2 updates
genesis['consensus_params']['block']['max_gas'] = '2000000'
genesis['consensus_params']['block']['max_bytes'] = '200000'

# enable transfers
genesis['app_state']['bank']['send_enabled'] = True
genesis['app_state']['distr']['withdraw_add_enabled'] = True

# Set new chain ID and genesis start time 
genesis['chain_id'] = chain_id.strip()
genesis['genesis_time'] = args.start_time

print(json.dumps(genesis, indent=True))
