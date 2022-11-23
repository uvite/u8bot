package goptuna

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

//go:generate stringer -trimprefix TrialState -output stringer_trial_state.go -type=TrialState

// TrialState is a state of Trial
type TrialState int

const (
	// TrialStateRunning means Trial is running.
	TrialStateRunning TrialState = iota
	// TrialStateComplete means Trial has been finished without any error.
	TrialStateComplete
	// TrialStatePruned means Trial has been pruned.
	TrialStatePruned
	// TrialStateFail means Trial has failed due to an uncaught error.
	TrialStateFail
	// TrialStateWaiting means Trial has been stopped, but may be resuming.
	TrialStateWaiting
)

// IsFinished returns true if trial is not running.
func (i TrialState) IsFinished() bool {
	return i != TrialStateRunning && i != TrialStateWaiting
}

// Trial is a process of evaluating an objective function.
//
// This object is passed to an objective function and provides interfaces to get parameter
// suggestion, manage the trial's state of the trial.
// Note that this object is seamlessly instantiated and passed to the objective function behind;
// hence, in typical use cases, library users do not care about instantiation of this object.
type Trial struct {
	Study               *Study
	ID                  int
	state               TrialState
	value               float64
	relativeParams      map[string]float64
	relativeSearchSpace map[string]interface{}
}

func (t *Trial) isFixedParam(name string, distribution interface{}) (float64, bool, error) {
	systemAttrs, err := t.GetSystemAttrs()
	if err != nil {
		return 0, false, err
	}
	fixedParamsJSON, ok := systemAttrs["fixed_params"]
	if !ok {
		return 0, false, nil
	}

	var fixedParams map[string]float64
	err = json.Unmarshal([]byte(fixedParamsJSON), &fixedParams)
	if err != nil {
		return 0, false, err
	}

	internalParam, ok := fixedParams[name]
	if !ok {
		return 0, false, nil
	}

	switch typedDistribution := distribution.(type) {
	case UniformDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	case LogUniformDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	case DiscreteUniformDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	case IntUniformDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	case StepIntUniformDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	case CategoricalDistribution:
		if !typedDistribution.Contains(internalParam) {
			return 0, false, nil
		}
	default:
		return 0, false, errors.New("unsupported distribution")
	}
	return internalParam, true, nil
}

// CallRelativeSampler should be called before evaluate an objective function only 1 time.
// Please note that this method is public for third party libraries like "Kubeflow/Katib".
// Goptuna users SHOULD NOT call this method.
func (t *Trial) CallRelativeSampler() error {
	if t.Study.RelativeSampler == nil {
		return nil
	}

	var err error
	var searchSpace map[string]interface{}
	if t.Study.definedSearchSpace != nil {
		searchSpace = t.Study.definedSearchSpace
	} else {
		searchSpace, err = IntersectionSearchSpace(t.Study)
		if err != nil {
			return err
		}
	}
	if searchSpace == nil {
		return nil
	}

	relativeSearchSpace := make(map[string]interface{}, len(searchSpace))
	for paramName := range searchSpace {
		distribution := searchSpace[paramName]
		if yes, _ := DistributionIsSingle(distribution); yes {
			continue
		}
		relativeSearchSpace[paramName] = distribution
	}

	frozen, err := t.Study.Storage.GetTrial(t.ID)
	if err != nil {
		return err
	}

	relativeParams, err := t.Study.RelativeSampler.SampleRelative(t.Study, frozen, searchSpace)
	if err == ErrUnsupportedSearchSpace {
		t.Study.logger.Warn("Your objective function contains unsupported search space for RelativeSampler.",
			fmt.Sprintf("trialID=%d", t.ID),
			fmt.Sprintf("searchSpace=%#v", searchSpace))
		return nil
	} else if err != nil {
		return err
	}

	t.relativeSearchSpace = searchSpace
	t.relativeParams = relativeParams
	return nil
}

func (t *Trial) isRelativeParam(name string, distribution interface{}) bool {
	expected, ok := t.relativeSearchSpace[name]
	if !ok {
		return false
	}
	return reflect.DeepEqual(expected, distribution)
}

func (t *Trial) suggest(name string, distribution interface{}) (float64, error) {
	trial, err := t.Study.Storage.GetTrial(t.ID)
	if err != nil {
		return 0.0, err
	}

	if value, ok, err := t.isFixedParam(name, distribution); err != nil {
		return 0.0, err
	} else if ok {
		err = t.Study.Storage.SetTrialParam(t.ID, name, value, distribution)
		return value, err
	}

	if t.isRelativeParam(name, distribution) {
		// isRelativeParam ensure that 'distribution' is same
		// with the one's in relativeSearchSpace.
		value, ok := t.relativeParams[name]
		if ok {
			err = t.Study.Storage.SetTrialParam(trial.ID, name, value, distribution)
			return value, err
		}
	}

	v, err := t.Study.Sampler.Sample(t.Study, trial, name, distribution)
	if err != nil {
		return 0.0, err
	}

	err = t.Study.Storage.SetTrialParam(trial.ID, name, v, distribution)
	return v, err
}

// ShouldPrune judges whether the trial should be pruned.
// This method calls prune method of the pruner, which judges whether
// the trial should be pruned at the given step.
// If it should be pruned, this method return ErrTrialPruned.
func (t *Trial) ShouldPrune(step int, value float64) error {
	if t.Study.Pruner == nil {
		t.Study.logger.Warn("Although it's not registered pruner, but you calls ShouldPrune method")
		return nil
	}

	if step < 0 {
		return errors.New("step should be larger equal than 0")
	}

	if err := t.Study.Storage.SetTrialIntermediateValue(t.ID, step, value); err != nil {
		return err
	}

	trial, err := t.Study.Storage.GetTrial(t.ID)
	if err != nil {
		return err
	}

	if shouldPrune, err := t.Study.Pruner.Prune(t.Study, trial); err != nil {
		return err
	} else if shouldPrune {
		return ErrTrialPruned
	}
	return nil
}

// Number return trial's number which is consecutive and unique in a study.
func (t *Trial) Number() (int, error) {
	return t.Study.Storage.GetTrialNumberFromID(t.ID)
}

// SuggestUniform suggests a value from a uniform distribution.
// Deprecated: This method will be removed at v1.0.0. Please use SuggestFloat method.
func (t *Trial) SuggestUniform(name string, low, high float64) (float64, error) {
	return t.SuggestFloat(name, low, high)
}

// SuggestLogUniform suggests a value from a uniform distribution in the log domain.
// Deprecated: This method will be removed at v1.0.0. Please use SuggestLogFloat method.
func (t *Trial) SuggestLogUniform(name string, low, high float64) (float64, error) {
	return t.SuggestLogFloat(name, low, high)
}

// SuggestDiscreteUniform suggests a value from a discrete uniform distribution.
// Deprecated: This method will be removed at v1.0.0. Please use SuggestDiscreteFloat method.
func (t *Trial) SuggestDiscreteUniform(name string, low, high, q float64) (float64, error) {
	return t.SuggestDiscreteFloat(name, low, high, q)
}

// SuggestFloat suggests a value for the floating point parameter.
func (t *Trial) SuggestFloat(name string, low, high float64) (float64, error) {
	if low > high {
		return 0, errors.New("'low' must be smaller than or equal to the 'high'")
	}
	return t.suggest(name, UniformDistribution{
		High: high, Low: low,
	})
}

// SuggestLogFloat suggests a value for the log-scale floating point parameter.
func (t *Trial) SuggestLogFloat(name string, low, high float64) (float64, error) {
	if low > high {
		return 0, errors.New("'low' must be smaller than or equal to the 'high'")
	}
	return t.suggest(name, LogUniformDistribution{
		High: high, Low: low,
	})
}

// SuggestDiscreteFloat suggests a value for the discrete floating point parameter.
func (t *Trial) SuggestDiscreteFloat(name string, low, high, q float64) (float64, error) {
	if low > high {
		return 0, errors.New("'low' must be smaller than or equal to the 'high'")
	}
	d := DiscreteUniformDistribution{
		High: high, Low: low, Q: q,
	}
	ir, err := t.suggest(name, d)
	if err != nil {
		return 0, err
	}
	return d.ToExternalRepr(ir).(float64), err
}

// SuggestInt suggests an integer parameter.
func (t *Trial) SuggestInt(name string, low, high int) (int, error) {
	if low > high {
		return 0, errors.New("'low' must be smaller than or equal to the 'high'")
	}
	d := IntUniformDistribution{
		High: high, Low: low,
	}
	v, err := t.suggest(name, d)
	return d.ToExternalRepr(v).(int), err
}

// SuggestStepInt suggests a step-interval integer parameter.
func (t *Trial) SuggestStepInt(name string, low, high, step int) (int, error) {
	if low > high {
		return 0, errors.New("'low' must be smaller than or equal to the 'high'")
	}
	if step <= 0 {
		return 0, errors.New("'step' must be larger than 0")
	}
	d := StepIntUniformDistribution{
		High: high, Low: low, Step: step,
	}
	v, err := t.suggest(name, d)
	return d.ToExternalRepr(v).(int), err
}

// SuggestCategorical suggests an categorical parameter.
func (t *Trial) SuggestCategorical(name string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", errors.New("'choices' must contains one or more elements")
	}
	v, err := t.suggest(name, CategoricalDistribution{
		Choices: choices,
	})
	return choices[int(v)], err
}

// SetUserAttr to store the value for the user.
func (t *Trial) SetUserAttr(key, value string) error {
	return t.Study.Storage.SetTrialUserAttr(t.ID, key, value)
}

// SetSystemAttr to store the value for the system.
func (t *Trial) SetSystemAttr(key, value string) error {
	return t.Study.Storage.SetTrialSystemAttr(t.ID, key, value)
}

// GetUserAttrs to store the value for the user.
func (t *Trial) GetUserAttrs() (map[string]string, error) {
	return t.Study.Storage.GetTrialUserAttrs(t.ID)
}

// GetSystemAttrs to store the value for the system.
func (t *Trial) GetSystemAttrs() (map[string]string, error) {
	return t.Study.Storage.GetTrialSystemAttrs(t.ID)
}

// GetContext returns a context which is registered at 'study.WithContext()'.
func (t *Trial) GetContext() context.Context {
	return t.Study.ctx
}
