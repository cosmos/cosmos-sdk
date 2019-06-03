#!/usr/bin/env python3

import lib


def process_raw_genesis(genesis, parsed_args):
    # migrate governance state as the internal structure of proposals has changed
    migrate_gov_data(genesis['app_state']['gov'])

    # Set new chain ID and genesis start time
    genesis['chain_id'] = parsed_args.chain_id.strip()
    genesis['genesis_time'] = parsed_args.start_time

    return genesis


def migrate_gov_data(gov_data):
    for p in gov_data['proposals']:
        p['content'] = p['proposal_content']
        del p['proposal_content']

        p['id'] = p['proposal_id']
        del p['proposal_id']


if __name__ == '__main__':
    parser = lib.init_default_argument_parser(
        prog_desc='Convert genesis.json from v0.34.x to v0.35.0',
        default_chain_id='cosmoshub-3',
        default_start_time='2019-02-11T12:00:00Z',
    )
    lib.main(parser, process_raw_genesis)
