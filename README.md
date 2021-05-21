![logo_transparent_wide](https://user-images.githubusercontent.com/5942370/107126087-ca589080-687b-11eb-820e-c6161c355eba.png)

Reactr is a fast, performant function scheduling system. It is designed to be flexible, with the ability to run embedded in your Go applications or as a standalone FaaS server, with first-class support for Wasm/WASI bundles.

Reactr can be used to execute a wide range of workloads from tiny serverless functions up to large long-running jobs. It is designed to scale automatically and adapt to the application you're designing.

## Wasm

Reactr has support for Wasm-packaged Runnables. The `rwasm` package contains a multi-tenant Wasm scheduler, an API to grant capabilities to Wasm Runnables, and support for several languages including Rust (stable) and Swift (experimental). See [wasm](./docs/wasm.md) and the [subo CLI](https://github.com/suborbital/subo) for details.

## FaaS

Reactr has early (read: alpha) support for acting as a Functions-as-a-Service system. Reactr can be run as a server, accepting jobs from HTTP/S and making the job results available to be fetched later. See [faas](./docs/faas.md) for details.

## Usage

Reactr has three basic units of work: `Task`, `Job`, and `Schedule`.
- A `Task` is a lightweight function that needs to be executed asynchronously.
- A `Job` is a defined piece of work that is given to Reactr to be executed by a pool of `Runnables`.
- A `Schedule` is a description of how to run a `Job` on a predefined schedule.

## Reactr has some very powerful capabilities, visit the [Reactr guide](./docs/guide.md) to learn all about it.

Reactr is being actively developed and has planned improvements, including optimized memory usage, library stability, additional Wasm capabilities, and more. Cheers!

Copyright Suborbital contributors 2021
