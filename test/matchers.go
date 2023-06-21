// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"errors"
	"fmt"
	"time"

	"github.com/onsi/gomega/types"

	"github.com/stv0g/cunicu/pkg/daemon"
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

func (matcher *entropyMatcher) Match(actual any) (success bool, err error) {
	buff, ok := actual.([]byte)
	if !ok {
		return false, fmt.Errorf("HaveEntropy matcher expects a byte slice")
	}

	return Entropy(buff) > matcher.minEntropy, nil
}

func (matcher *entropyMatcher) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto have at least an entropy of %f", actual, matcher.minEntropy)
}

func (matcher *entropyMatcher) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto have an entropy lower than %f", actual, matcher.minEntropy)
}

func ReceiveEvent[E any](e *E) types.GomegaMatcher {
	return &eventMatcher[E]{
		event: e,
	}
}

type eventMatcher[E any] struct {
	event *E
}

func (matcher *eventMatcher[E]) Match(actual any) (success bool, err error) {
	events, ok := actual.(chan daemon.Event)
	if !ok {
		return false, errors.New("actual is not an event channel")
	}

	to := time.NewTimer(time.Second)
	select {
	case <-to.C:
		return false, fmt.Errorf("timed-out")
	case evt := <-events:
		if *matcher.event, ok = evt.(E); !ok {
			return false, fmt.Errorf("received wrong type of event: %#+v", evt)
		}

		return true, nil
	}
}

func (matcher *eventMatcher[E]) FailureMessage(_ any) (message string) {
	return "Did not receive expected event"
}

func (matcher *eventMatcher[E]) NegatedFailureMessage(_ any) (message string) {
	return "Received event unexpectedly"
}
