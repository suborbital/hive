package rt

import (
	"testing"

	"github.com/pkg/errors"
)

// ExpensiveTask adds two numbers
func ExpensiveTask(first, second int) (string, Task) {
	return "expensive", func(ctx *Ctx) (interface{}, error) {
		if first < 0 {
			return nil, errors.New("bad input")
		}

		return first + second, nil
	}
}

func TestBasicTask(t *testing.T) {
	r := New()

	val, err := r.Run(
		ExpensiveTask(3, 4)).Then()
	if err != nil {
		t.Error(err)
		return
	}

	if val.(int) != 7 {
		t.Error(errors.New("incorrect value"))
	}
}

func TestTaskError(t *testing.T) {
	r := New()

	_, err := r.Run(
		ExpensiveTask(-1, 4)).Then()
	if err == nil {
		t.Error(errors.New("should have errored..."))
		return
	}
}
