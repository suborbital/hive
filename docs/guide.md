# Get started with Reactr 🚀

Reactr is a function scheduler that allows you to spawn a huge amount of 'work' and ensure it gets executed in an efficient manner without overwhelming your CPU. There are two types of 'work' that Reactr can handle; `Tasks` and `Jobs`. Tasks are lightweight functions that Reactr executes asynchronously, and Jobs contain a payload describing work to be done, and are given to Reactr to be handled by a pre-registered worker.

### The Basics

First, install Reactr's core package `rt`:
```bash
go get github.com/suborbital/reactr/rt
```

And then get started by defining something `Runnable`:
```golang
type generic struct{}

// Run runs a generic job
func (g generic) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	fmt.Println("doing job:", job.String()) // get the string value of the job's data

	// do your work here

	return fmt.Sprintf("finished %s", job.String()), nil
}

// OnChange is called when Reactr starts or stops a worker to handle jobs,
// and allows the Runnable to set up before receiving jobs or tear down if needed.
func (g generic) OnChange(change rt.ChangeEvent) error {
	return nil
}
```
A `Runnable` is something that can take care of a job, all it needs to do is conform to the `Runnable` interface as you see above.

Once you have a Runnable, create a Reactr instance, register it, and `Do` some work:
```golang
package main

import (
	"fmt"
	"log"

	"github.com/suborbital/reactr/rt"
)

func main() {
	r := rt.New()

	r.Handle("generic", generic{})

	res := r.Do(r.Job("generic", "hard work"))

	res, err := res.Then()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("done!", res.(string))
}
```
When you `Do` some work, you get a `Result`. A result is like a Rust future or a JavaScript promise, it is something you can get the job's result from once it is finished.

Calling `Then()` will block until the job is complete, and then give you the return value from the Runnable's `Run`. Cool, right?

## Runnables

There are some more complicated things you can do with Runnables:
```golang
type recursive struct{}

// Run runs a recursive job
func (r recursive) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	fmt.Println("doing job:", job.String())

	if job.String() == "first" {
		return ctx.Do(rt.NewJob("recursive", "second")), nil
	} else if job.String() == "second" {
		return ctx.Do(rt.NewJob("recursive", "last")), nil
	}

	return fmt.Sprintf("finished %s", job.String()), nil
}

func (r recursive) OnChange(change rt.ChangeEvent) error {
	return nil
}
```
The `rt.Ctx` you see there is the job context, and one of the things it can do is run more things!

Calling `ctx.Do` will schedule another job to be executed and give you a `Result`. If you return a `Result` from `Run`, then the caller will recursively recieve that `Result` when they call `Then()`!

For example:
```golang
r := r.Do(r.Job("recursive", "first"))

res, err := r.Then()
if err != nil {
	log.Fatal(err)
}

fmt.Println("done!", res.(string))
```
Will cause this output:
```
doing job: first
doing job: second
doing job: last
done! finished last
```
The ability to chain jobs is quite powerful!

You won't always need or care about a job's output, and in those cases, make sure to call `Discard()` on the result to allow the underlying resources to be deallocated!
```golang
r.Do(r.Job("recursive", "first")).Discard()
```

To to something asynchronously with the `Result` once it completes, call `ThenDo` on the result:
```golang
r.Do(r.Job("generic", "first")).ThenDo(func(res interface{}, err error) {
	if err != nil {
		// do something with the error
	}

	//do something with the result
})
```
`ThenDo` will return immediately, and provided callback will be run on a background goroutine. This is useful for handling results that don't need to be consumed by your main program execution.

### Groups

A reactr `Group` is a set of `Result`s that belong together. If you're familiar with Go's `errgroup.Group{}`, it is similar. Adding results to a group will allow you to evaluate them all together at a later time.
```golang
grp := rt.NewGroup()

grp.Add(ctx.Do(rt.NewJob("recursive", "first")))
grp.Add(ctx.Do(rt.NewJob("generic", "group work")))
grp.Add(ctx.Do(rt.NewJob("generic", "group work")))

if err := grp.Wait(); err != nil {
	log.Fatal(err)
}
```
Will print: 
```
doing job: first
doing job: group work
doing job: group work
doing job: second
doing job: last
```
As you can see, the "recursive" jobs from the `generic` runner get queued up after the two jobs that don't recurse.

Note that you cannot get result values from result groups, the error returned from `Wait()` will be the first error from any of the results in the group, if any. To get result values from a group of jobs, put them in an array and call `Then` on them individually.

**TIP** If you return a group from a Runnable's `Run`, calling `Then()` on the result will recursively call `Wait()` on the group and return the error to the original caller! You can easily chain jobs and job groups in various orders.

### Pools
Each `Runnable` that you register is given a worker to process their jobs. By default, each worker has one work thread processing jobs in sequence. If you want a particular worker to process more than one job concurrently, you can increase its `PoolSize`:
```golang
doGeneric := r.Handle("generic", generic{}, rt.PoolSize(3))

grp := rt.NewGroup()
grp.Add(doGeneric("first"))
grp.Add(doGeneric("second"))
grp.Add(doGeneric("random"))

if err := grp.Wait(); err != nil {
	log.Fatal(err)
}
```
Passing `PoolSize(3)` will spawn three work threads to process `generic` jobs.

### Timeouts
By default, if a job becomes stuck and is blocking execution, it will block forever. If you want to have a worker time out after a certain amount of seconds on a stuck job, pass `rt.TimeoutSeconds` to Handle:
``` golang
h := rt.New()

doTimeout := r.Handle("timeout", timeoutRunner{}, rt.TimeoutSeconds(3))
```
When `TimeoutSeconds` is set and a job executes for longer than the provided number of seconds, the worker will move on to the next job and `ErrJobTimeout` will be returned to the Result. The failed job will continue to execute in the background, but its result will be discarded.

### Schedules
The `r.Do` method will run your job immediately, but if you need to run a job at a later time, at a regular interval, or on some other schedule, then the `Schedule` interface will help. The `Schedule` interface allows for an object to choose when to execute a job. Any object that conforms to the interface can be used as a Schedule:
```golang
// Schedule is a type that returns an *optional* job if there is something that should be scheduled.
// Reactr will poll the Check() method at regular intervals to see if work is available.
type Schedule interface {
	Check() *Job
	Done() bool
}
```
The `r.Schedule` method will allow you to register a Schedule, and there are two built-in schedules(`Every` and `After`) to help:
```golang
r := rt.New()

r.Handle("worker", &workerRunner{})

// runs every hour
r.Schedule(rt.Every(60*60, func() Job {
	return NewJob("worker", nil)
}))
```
Reactr will poll all registered Schedules at a 1 second interval to `Check` for new jobs. Schedules can end their own execution by returning `false` from the `Done` method. You can use the Schedules provided with Reactr or develop your own.

Scheduled jobs' results are discarded automatically using `Discard()`

### Advanced Runnables

The `Runnable` interface defines an `OnChange` function which gives the Runnable a chance to prepare itself for changes to the worker running it. For example, when a Runnable is registered with a pool size greater than 1, the Runnable may need to provision resources for itself to enable handling jobs concurrently, and `OnChange` will be called once each time a new worker starts up. Our [Wasm implementation](https://github.com/suborbital/reactr/blob/master/rwasm/wasmrunnable.go) is a good example of this. 

Most Runnables can return `nil` from this function, however returning an error will cause the worker start to be paused and retried until the required pool size has been acheived. The number of seconds between retries (default 3) and the maximum number of retries (default 5) can be configured when registering a Runnable:
```golang
doBad := r.Handle("badRunner", badRunner{}, rt.RetrySeconds(1), rt.MaxRetries(10))
```
Any error from a failed worker will be returned to the first job that is attempted for that Runnable.

### Pre-warming
When a Runnable is mounted, it is simply registered as available to receive work. The Runnable is not actually invoked until the first job of the given type is received. For basic Runnables, this is normally fine, but for Runnables who use the `OnChange` method to provision resources, this can cause the first job to be slow. The `PreWarm` option is available to allow Runnables to be started as soon as they are mounted, rather than waiting for the first job. This mitigates cold-starts when anything expensive is needed at startup.
```
doExpensive := r.Handle("expensive", expensiveRunnable{}, rt.PreWarm())
```

### Shortcuts

There are also some shortcuts to make working with Reactr a bit easier:
```golang
type input struct {
	First, Second int
}

type math struct{}

// Run runs a math job
func (g math) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	in := job.Data().(input)

	return in.First + in.Second, nil
}
```
```golang
doMath := r.Handle("math", math{})

for i := 1; i < 10; i++ {
	equals, _ := doMath(input{i, i * 3}).ThenInt()
	fmt.Println("result", equals)
}
```
The `Handle` function returns an optional helper function. Instead of passing a job name and full `Job` into `r.Do`, you can use the helper function to instead just pass the input data for the job, and you receive a `Result` as normal. `doMath`!

## Additional features

Reactr can integrate with [Grav](https://github.com/suborbital/grav), which is the decentralized message bus developed as part of the Suborbital Development Platform. Read about the integration on [the grav documentation page.](./grav.md)

Reactr provides the building blocks for scalable asynchronous systems. This should be everything you need to help you improve the performance of your application. When you are looking to take advantage of Reactr's other features, check out its [FaaS](./faas.md) and [Wasm](./wasm.md) capabilities!