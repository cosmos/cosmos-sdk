# Additive Increase Multiplicative Decrease (AIMD) EIP-1559

## Overview

> **Definitions:**
>
> * **`Target Block Size`**: This is the target block gas consumption.
> * **`Max Block Size`**: This is the maximum block gas consumption.

This plugin implements the AIMD (Additive Increase Multiplicative Decrease) EIP-1559 fee market as described in this [AIMD EIP-1559](https://arxiv.org/abs/2110.04753) research publication.

The AIMD EIP-1559 fee market is a slight modification to Ethereum's EIP-1559 fee market. Specifically it introduces the notion of a adaptive learning rate which scales the base fee (reserve price to be included in a block) more aggressively when the network is congested and less aggressively when the network is not congested. This is primarily done to address the often cited criticism of EIP-1559 that it's base fee often lags behind the current demand for block space. The learning rate on Ethereum is effectively hard-coded to be 12.5%, which means that between any two blocks the base fee can maximally increase by 12.5% or decrease by 12.5%. Additionally, AIMD EIP-1559 differs from Ethereum's EIP-1559 by considering a configured time window (number of blocks) to consider when calculating and comparing target block utilization and current block utilization.

## Parameters

### Ethereum EIP-1559

Base EIP-1559 currently utilizes the following parameters to compute the base fee:

* **`PreviousBaseFee`**: This is the base fee from the previous block. This must be a value that is greater than `0`.
* **`TargetBlockSize`**: This is the target block size in bytes. This must be a value that is greater than `0`.
* **`PreviousBlockSize`**: This is the block size from the previous block.

The calculation for the updated base fee for the next block is as follows:

```golang
currentBaseFee := previousBaseFee * (1 + 0.125 * (currentBlockSize - targetBlockSize) / targetBlockSize)
```

### AIMD EIP-1559

AIMD EIP-1559 introduces a few new parameters to the EIP-1559 fee market:

* **`Alpha`**: This is the amount we additively increase the learning rate when the target utilization is less than the current utilization i.e. the block was more full than the target size. This must be a value that is greater than `0.0`.
* **`Beta`**: This is the amount we multiplicatively decrease the learning rate when the target utilization is greater than the current utilization i.e. the block was less full than the target size. This must be a value that is greater than `0.0`.
* **`Window`**: This is the number of blocks we look back to compute the current utilization. This must be a value that is greater than `0`. Instead of only utilizing the previous block's utilization, we now consider the utilization of the previous `Window` blocks.
* **`Gamma`**: This determines whether you are additively increase or multiplicatively decreasing the learning rate based on the target and current block utilization. This must be a value that is between `[0, 1]`. For example, if `Gamma = 0.25`, then we multiplicatively decrease the learning rate if the average ratio of current block size to max block size over some window of blocks is within `(0.25, 0.75)` and additively increase it if outside that range.
* **`MaxLearningRate`**: This is the maximum learning rate that can be applied to the base fee. This must be a value that is between `[0, 1]`.
* **`MinLearningRate`**: This is the minimum learning rate that can be applied to the base fee. This must be a value that is between `[0, 1]`.
* **`Delta`**: This is a trailing constant that is used to smooth the learning rate. In order to further converge the long term net gas usage and net gas goal, we introduce another integral term which tracks how much gas off from 0 gas we’re at. We add a constant c which basically forces the fee to slowly trend in some direction until this has gone to 0.

The calculation for the updated base fee for the next block is as follows:

```golang

// sumBlockSizesInWindow returns the sum of the block sizes in the window.
blockConsumption := sumBlockSizesInWindow(window) / (window * maxBlockSize)

if blockConsumption < gamma || blockConsumption > 1 - gamma {
    // MAX_LEARNING_RATE is a constant that is configured by the chain developer
    newLearningRate := min(MaxLearningRate, alpha + currentLearningRate)
} else {
    // MIN_LEARNING_RATE is a constant that is configured by the chain developer
    newLearningRate := max(MinLearningRate, beta * currentLearningRate)
}

// netGasDelta returns the net gas difference between every block in the window and the target block size.
newBaseFee := currentBaseFee * (1 + newLearningRate * (currentBlockSize - targetBlockSize) / targetBlockSize) + delta * netGasDelta(window)
```

## Examples

> **Assume the following parameters:**
>
> * `TargetBlockSize = 50`
> * `MaxBlockSize = 100`
> * `Window = 1`
> * `Alpha = 0.025`
> * `Beta = 0.95`
> * `Gamma = 0.25`
> * `MAX_LEARNING_RATE = 1.0`
> * `MIN_LEARNING_RATE = 0.0125`
> * `Current Learning Rate = 0.125`
> * `Previous Base Fee = 10.0`
> * `Delta = 0`

### Block is Completely Empty

In this example, we expect the learning rate to additively increase and the base fee to decrease.

```golang
blockConsumption := sumBlockSizesInWindow(1) / (1 * 100) == 0
newLearningRate := min(1.0, 0.025 + 0.125) == 0.15
newBaseFee := 10 * (1 + 0.15 * (0 - 50) / 50) == 8.5
```

As we can see, the base fee decreased by 1.5 and the learning rate increases.

### Block is Completely Full

In this example, we expect the learning rate to multiplicatively increase and the base fee to increase.

```golang
blockConsumption := sumBlockSizesInWindow(1) / (1 * 100) == 1
newLearningRate := min(1.0, 0.025 + 0.125) == 0.15
newBaseFee := 10 * (1 + 0.95 * 0.125) == 11.875
```

As we can see, the base fee increased by 1.875 and the learning rate increases.

### Block is at Target Utilization

In this example, we expect the learning rate to decrease and the base fee to remain the same.

```golang
blockConsumption := sumBlockSizesInWindow(1) / (1 * 100) == 0.5
newLearningRate := max(0.0125, 0.95 * 0.125) == 0.11875
newBaseFee := 10 * (1 + 0.11875 * (0 - 50) / 50) == 10
```

As we can see, the base fee remained the same and the learning rate decreased.

## Default EIP-1559 With AIMD EIP-1559

It is entirely possible to implement the default EIP-1559 fee market with the AIMD EIP-1559 fee market. This can be done by setting the following parameters:

* `Alpha = 0.0`
* `Beta = 1.0`
* `Gamma = 1.0`
* `MAX_LEARNING_RATE = 0.125`
* `MIN_LEARNING_RATE = 0.125`
* `Delta = 0`
* `Window = 1`
