package main

import (
	"fmt"
	"time"

	"github.com/suborbital/reactr/rt"
)

func main() {
	r := rt.New()

	val, err := r.Run(
		SomethingExpensive(1, "hello", false)).Then()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(val)

	r.Register("print", &printJob{})

	r.Do(rt.NewJob("print", "start")).Discard()

	r.Schedule(rt.Every(5, func() rt.Job {
		return rt.NewJob("print", "every 5")
	}))

	r.Schedule(rt.After(7, func() rt.Job {
		return rt.NewJob("print", "after 7")
	}))

	time.Sleep(time.Second * 25)
}

type printJob struct{}

func (p *printJob) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	fmt.Println(job.String())
	return nil, nil
}

func (p *printJob) OnChange(c rt.ChangeEvent) error { return nil }

func SomethingExpensive(first int, second string, third bool) (string, rt.Task) {
	return "something.expensive", func(ctx *rt.Ctx) (interface{}, error) {
		return fmt.Sprintf("%d %s", first, second), nil
	}
}
