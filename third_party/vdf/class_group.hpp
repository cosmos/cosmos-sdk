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

#ifndef __SRC_CPP_CLASS_GROUP_HPP_
#define __SRC_CPP_CLASS_GROUP_HPP_

#include <gmpxx.h>
#include <stdint.h>
#include <vector>

typedef mpz_class integer;

namespace class_group {
/**
 * A binary quadratic form, consisting of three integers (a, b, c).
 *
 * `y = ax^2 + bxy + cy^2`
 *
 * `b^2 - 4ac` must be a negative discriminant `d`.
 */
struct form {
    integer a;
    integer b;
    integer c;
};

/**
 * Returns `x * y`, where `*` is the group operation.
 * `x` and `y` must have the same discriminant.
 */
form operator*(const form& x, const form& y);

/**
 * Returns `x^k`, where `^` is the group exponentiation.
 */
form operator^(const form& x, const integer& k);

/**
 * Returns the identity element in the group with discriminant `d`.
 * `d` must be a negative integer, `d < 0` and `d = 0, 1 (mod 4)`.
 */
form identity(const integer& d);

/**
 * Returns the inverse of `x`.
 */
form inverse(const form& x);

/**
 * Returns `true` iff `x` is a valid reduced form, `false` otherwise.
 */
bool is_valid(const form& x, const integer& d);

/**
 * Creates a form given `a, b` and a negative discriminant `d`.
 *
 * `d < 0`
 * `d = 0, 1 (mod 4)`
 * `b^2 - 4ac = d`
 */
form from_a_b(const integer& a, const integer& b, const integer& d);

/**
 * Creates a form from a byte representation.
 */
form from_bytes(const std::vector<uint8_t>& buffer, const integer& d);

/**
 * Returns a byte representation of the form.
 */
std.vector<uint8_t> to_bytes(const form& x);
}  // namespace class_group

#endif  // __SRC_CPP_CLASS_GROUP_HPP_