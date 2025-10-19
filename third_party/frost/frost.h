// Copyright 2022 The FROST authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef __FROST_H__
#define __FROST_H__

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque structs
typedef struct frost_secp256k1_key_share frost_secp256k1_key_share;
typedef struct frost_secp256k1_public_key frost_secp256k1_public_key;
typedef struct frost_secp256k1_signature frost_secp256k1_signature;

/**
 * Generates a new key share.
 *
 * `threshold` is the number of signers required to produce a valid signature.
 * `num_shares` is the total number of key shares.
 * `share_index` is the index of the key share to generate.
 */
frost_secp256k1_key_share* frost_secp256k1_key_share_create(
    uint32_t threshold,
    uint32_t num_shares,
    uint32_t share_index);

/**
 * Destroys a key share.
 */
void frost_secp256k1_key_share_destroy(frost_secp256k1_key_share* key_share);

/**
 * Derives the public key from a key share.
 */
frost_secp256k1_public_key* frost_secp256k1_public_key_from_key_share(
    const frost_secp256k1_key_share* key_share);

/**
 * Destroys a public key.
 */
void frost_secp256k1_public_key_destroy(frost_secp256k1_public_key* public_key);

/**
 * Signs a message using a key share.
 *
 * `msg` is the message to sign.
 * `msg_len` is the length of the message.
 */
frost_secp256k1_signature* frost_secp256k1_sign(
    const frost_secp256k1_key_share* key_share,
    const uint8_t* msg,
    size_t msg_len);

/**
 * Destroys a signature.
 */
void frost_secp256k1_signature_destroy(frost_secp256k1_signature* signature);

/**
 * Verifies a signature.
 *
 * `msg` is the message to verify.
 * `msg_len` is the length of the message.
 */
int frost_secp256k1_verify(
    const frost_secp256k1_public_key* public_key,
    const uint8_t* msg,
    size_t msg_len,
    const frost_secp256k1_signature* signature);

#ifdef __cplusplus
}
#endif

#endif  // __FROST_H__