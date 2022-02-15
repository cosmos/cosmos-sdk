<!--
order: 7
-->

# Parameters

The distribution module contains the following parameters:

| Key                 | Type         | Example                    |
| ------------------- | ------------ | -------------------------- |
| communitytax        | string (dec) | "0.020000000000000000" [0] |
| baseproposerreward  | string (dec) | "0.010000000000000000" [0] |
| bonusproposerreward | string (dec) | "0.040000000000000000" [0] |
| withdrawaddrenabled | bool         | true                       |

* [0] `communitytax`, `baseproposerreward` and `bonusproposerreward` must be
  positive and their sum cannot exceed 1.00.
