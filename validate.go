package validate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

type errMultiple []error

func (e errMultiple) Error() string {
	msg := make([]string, 0, len(e))
	for _, q := range e {
		msg = append(msg, q.Error())
	}
	return "validate: " + strconv.Itoa(len(e)) + " errors: [" + strings.Join(msg, "; ") + "]"
}

func All(vs ...error) error {
	var errs []error
	for _, err := range vs {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errMultiple(errs)
	}
	return nil
}

type errSingle[T any] struct {
	Name string
	Val  T
	To   T
	Op   string
}

func (e errSingle[T]) Error() string {
	return fmt.Sprintf("%s(%v) %s (%v)", e.Name, e.Val, e.Op, e.To)
}

func Min[T constraints.Ordered](name string, v, min T) error {
	if v > min {
		return nil
	}
	return errSingle[T]{Name: name, Val: v, To: min, Op: "smaller than min"}
}

func Max[T constraints.Ordered](name string, v, max T) error {
	if v < max {
		return nil
	}
	return errSingle[T]{Name: name, Val: v, To: max, Op: "higher than max"}
}

func Before(name string, t, before time.Time) error {
	if t.Before(before) {
		return nil
	}
	return errSingle[time.Time]{Name: name, Val: t, To: before, Op: "is not before"}
}

func After(name string, t, after time.Time) error {
	if t.After(after) {
		return nil
	}
	return errSingle[time.Time]{Name: name, Val: t, To: after, Op: "is not after"}
}

type errOneOf[T any] struct {
	Name    string
	Val     T
	Options []T
}

func (e errOneOf[T]) Error() string {
	return fmt.Sprintf("%s(%v) not in %v", e.Name, e.Val, e.Options)
}

func OneOf[T comparable](name string, v T, options ...T) error {
	for _, q := range options {
		if v == q {
			return nil
		}
	}
	return errOneOf[T]{Name: name, Val: v, Options: options}
}
