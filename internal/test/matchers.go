package test

import (
	"github.com/onsi/gomega/types"

	"fmt"
)

func BeRandom() types.GomegaMatcher {
	return HaveEntropy(4)
}

func HaveEntropy(minEntropy float64) types.GomegaMatcher {
	return &entropyMatcher{
		minEntropy: minEntropy,
	}
}

type entropyMatcher struct {
	minEntropy float64
}

func (matcher *entropyMatcher) Match(actual interface{}) (success bool, err error) {
	buff, ok := actual.([]byte)
	if !ok {
		return false, fmt.Errorf("HaveEntropy matcher expects a byte slice")
	}

	return Entropy(buff) > matcher.minEntropy, nil
}

func (matcher *entropyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto have at least an entropy of %f", actual, matcher.minEntropy)
}

func (matcher *entropyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto have an entropy lower than %f", actual, matcher.minEntropy)
}
