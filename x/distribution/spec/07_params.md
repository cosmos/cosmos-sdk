<!--
order: 7
-->

# Parameters

The distribution module contains the following parameters:

| Key                 | Type         | Example                    |
| ------------------- | ------------ | -------------------------- |
| communitytax        | string (dec) | "0.020000000000000000" [0] |
| baseproposerreward  | string (dec) | "0.010000000000000000" [1] |
| bonusproposerreward | string (dec) | "0.040000000000000000" [1] |
| withdrawaddrenabled | bool         | true                       |

* [0] The value of `communitytax` must be positive and cannot exceed 1.00.
* [1] `baseproposerreward` and `bonusproposerreward` must be positive and their sum cannot exceed 1.00.
