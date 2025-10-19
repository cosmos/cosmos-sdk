// Copyright 2021 Chia Network Inc
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

#ifndef __SRC_CPP_PIETRZAK_HPP_
#define __SRC_CPP_PIETRZAK_HPP_

#include "class_group.hpp"

namespace pietrzak {
/**
 * Generate a random hash_prime in the range [0, 2^l_prime_bits).
 */
integer hash_prime(const integer& x, int l_prime_bits);

/**
 * Creates a Pietrzak VDF proof.
 *
 * `d` must be a negative discriminant.
 * `x` must be a reduced `form` with discriminant `d`.
 * `num_iterations` is `T` from the paper.
 * `l_prime_bits` is `l` from the paper.
 */
std::vector<uint8_t> prove(
    const integer& d,
    const form& x,
    uint64_t num_iterations,
    int l_prime_bits);

/**
 * Verifies a Pietrzak VDF proof.
 *
 * `d` must be a negative discriminant.
 * `x_s` must be a byte representation of the input form `x`.
 * `y_s` must be a byte representation of the output form `y = x^(2^T)`.
 * `proof_s` must be a byte representation of the proof.
 * `num_iterations` is `T` from the paper.
 * `l_prime_bits` is `l` from the paper.
 */
bool verify(
    const integer& d,
    const std::vector<uint8_t>& x_s,
    const std::vector<uint8_t>& y_s,
    const std::vector<uint8_t>& proof_s,
    uint64_t num_iterations,
    int l_prime_bits);
}  // namespace pietrzak

#endif  // __SRC_CPP_PIETRZAK_HPP_