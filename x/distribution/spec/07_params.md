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

* [0] The value of `communitytax` must be positive and cannot exceed 1%.
* [1] The sum of `baseproposerreward` and `bonusproposerreward` cannot exceed 1%.
Both values must be positive.
