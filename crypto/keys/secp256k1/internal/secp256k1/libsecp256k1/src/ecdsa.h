/**********************************************************************
 * Copyright (c) 2013, 2014 Pieter Wuille                             *
 * Distributed under the MIT software license, see the accompanying   *
 * file COPYING or http://www.opensource.org/licenses/mit-license.php.*
 **********************************************************************/

#ifndef _SECP256K1_ECDSA_
#define _SECP256K1_ECDSA_

#include <stddef.h>

#include "scalar.h"
#include "group.h"
#include "ecmult.h"

static int secp256k1_ecdsa_sig_parse(cosmos_secp256k1_scalar *r, cosmos_secp256k1_scalar *s, const unsigned char *sig, size_t size);
static int secp256k1_ecdsa_sig_serialize(unsigned char *sig, size_t *size, const cosmos_secp256k1_scalar *r, const cosmos_secp256k1_scalar *s);
static int secp256k1_ecdsa_sig_verify(const secp256k1_ecmult_context *ctx, const cosmos_secp256k1_scalar* r, const cosmos_secp256k1_scalar* s, const secp256k1_ge *pubkey, const cosmos_secp256k1_scalar *message);
static int secp256k1_ecdsa_sig_sign(const secp256k1_ecmult_gen_context *ctx, cosmos_secp256k1_scalar* r, cosmos_secp256k1_scalar* s, const cosmos_secp256k1_scalar *seckey, const cosmos_secp256k1_scalar *message, const cosmos_secp256k1_scalar *nonce, int *recid);

#endif
