package cmaes

import (
	"math/rand"
)

const (
	restartStrategyIPOP  = "ipop"
	restartStrategyBIPOP = "bipop"
)

// SamplerOption is a type of the function to customizing CMA-ES sampler.
type SamplerOption func(sampler *Sampler)

// SamplerOptionSeed sets seed number.
func SamplerOptionSeed(seed int64) SamplerOption {
	return func(sampler *Sampler) {
		sampler.rng = rand.New(rand.NewSource(seed))
	}
}

// SamplerOptionInitialMean sets the initial mean vectors.
func SamplerOptionInitialMean(mean map[string]float64) SamplerOption {
	return func(sampler *Sampler) {
		sampler.x0 = mean
	}
}

// SamplerOptionInitialSigma sets the initial sigma.
func SamplerOptionInitialSigma(sigma float64) SamplerOption {
	return func(sampler *Sampler) {
		sampler.sigma0 = sigma
	}
}

// SamplerOptionOptimizerOptions sets the options for Optimizer.
func SamplerOptionOptimizerOptions(opts ...OptimizerOption) SamplerOption {
	return func(sampler *Sampler) {
		sampler.optimizerOptions = opts
	}
}

// SamplerOptionNStartupTrials sets the number of startup trials.
func SamplerOptionNStartupTrials(nStartupTrials int) SamplerOption {
	return func(sampler *Sampler) {
		sampler.nStartUpTrials = nStartupTrials
	}
}

// SamplerOptionIPop enables restart CMA-ES with increasing population size.
// The argument is multiplier of population size before each restart and basically you should choose 2.
// From the experiments in the IPOP-CMA-ES, it reveal similar performance for factors between 2 and 3.
func SamplerOptionIPop(incPopSize int) SamplerOption {
	return func(sampler *Sampler) {
		sampler.restartStrategy = restartStrategyIPOP
		sampler.incPopSize = incPopSize
	}
}

// SamplerOptionBIPop enables restart CMA-ES with two interlaced restart strategies,
// one with an increasing population size and one with varying small population size.
// The argument is multiplier of population size before each restart and basically you should choose 2.
func SamplerOptionBIPop(incPopSize int) SamplerOption {
	return func(sampler *Sampler) {
		sampler.restartStrategy = restartStrategyBIPOP
		sampler.incPopSize = incPopSize
	}
}
