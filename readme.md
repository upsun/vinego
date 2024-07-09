# Vinego

Vinego is a new set of linters built on the Go `analysis.Analyzer` framework. Many of the linters focus on increasing strictness around variable initialization and type safety. They fall somewhere in-between "drop these into your codebase with no changes" and "rewrite all your code to conform to weird new language-extra conventions" in terms of invasiveness.

The linters are built as a single `.so` with control over individual checks via a `.vinego.yaml` file. Methods for using this with `golangci-lint` are provided, but you can use them with other frameworks as well.

# Analyzers

- `allfields`

  Confirms that all required fields are explicitly initialized in struct literals.

  Add a comment before a struct type like:

  ```
  // check:allfields
  type MyStruct {
    X int
    Y string
    Z bool `optional:""`
  }
  ```

  Then, if you do (for instance):

  ```
  m := MyStruct{X: 4}
  ```

  you'll get an error that `Y` is missing.

- `varinit`

  Enabled with `enable_varinit: true` in `.vinego.yaml`.

  Checks that all variables have been explicitly initialized (with a value) before usage.

  For example:

  ```go
  var i int
  if something {
     i = 4
  } else {
     doOtherThings()
  }
  myFunc(i)
  ```

  would produce an error saying that `i` hasn't been initialized in the `else` branch.

  Taking the address of a variable is considered initialization (ex: in doing `json.Unmarshal(bytes, &config)` config would be marked as initialized).

- `explicitcast`

  Enabled with `enable_explicitcast: true` in `.vinego.yaml`.

  Checks that primitive literals are never implicitly casted (during assignments, function calls, and returns).

  For example:

  ```go
  var x time.Duration
  x = 24
  ```

  would produce an error saying that `24` is being implicitly cast to `time.Duration`.

- `loopvariableref`

  Enabled with `enable_loopvariableref: true` in `.vinego.yaml`.

  Checks that loop variables are not used outside of a single iteration, either via reference/pointer or function capture.

  For example:

  ```go
  for _, x := range myList {
     go func() {
        print(x)
     }
  }
  ```

  would produce an error saying that the capture of `x` here is risky.

  The simple work around is to explicitly declare a local variable like:

  ```go
  for _, x := range myList {
     x := x
     go func() {
        print(x)
     }
  }
  ```

  to make sure the captured data is unique.

  This will be obsolete after [Go 1.22](https://go.dev/blog/loopvar-preview).

- `capturederr`

  Enabled with `enable_capturederr: true` in `.vinego.yaml`.

  The `staticcheck` linter `SA4006` check which makes sure we properly consume error variables [ignores anything that happens with captured variables](https://github.com/dominikh/go-tools/issues/287). Therefore if you accidentally capture `err` from an outer function, assign it a value, then never check it, `SA4006` won't help you. Go's behavior using variable reuse/reinitialization with `=` and `:=` makes it easy to transplant code and accidentally reuse an existing variable, which makes it easy to accidentally capture external variables in closures.

  `capturederror` will give you an error when an error variable is captured by a closure (_specifically_ error variables). As with the other linters here, capturing an error variable isn't necessarily incorrect, but it's generally unintended and when done unintentionally can lead to hard to track incorrect failure behavior.

  Example usage:

  ```go
  err := something()
  if err != nil {
     return err
  }
  wrapper(func() error {
     err = otherthing()
  })
  ...
  ```

  would produce an error saying that we're assgning to the captured variable `err`. (The workaround is to do `err := otherthing()`, which would make this error disappear and the `SA4006` one appear in its place, as intended.)

# Usage

1. We provide a pre-made Docker batteries-included image for CI and development environments: `https://ghcr.io/platformsh/vinego:latest`

   It includes

   - Go
   - Vinego
   - `golangci-lint`
   - `gci`
   - `goimports`
   - `dlv`
   - `staticcheck`

   (Distributed this way because `golangci-lint` needs linters to be built with the same dependency versions and the easiest way to guarantee that is to build them together)

   You can build the Docker container yourself with `docker build --tag vinego src` at the root of this repo.

   Alternatively, you can build just the plugin `.so` - see the `Dockerfile` for details (it's a straightforward Go `.so` build).

1. Add custom linter plugin to your project's `.golangci.yaml` file:

   ```
   linters:
      enable:
         - vinego

   linters-settings:
      custom:
         vinego:
            path: "/custom_linters/vinego.so"
            description: "Vinego linters"
   ```

1. For optional linters, enable them in a `.vinego.yaml` in the same directory as `.golangci.json`. For details see the per-linter explanations above.

1. Run the linters with `docker run --rm --volume $PWD:/mnt --workdir /mnt vinego /bin/golangci-lint run --verbose`. You should see `vinego` listed in the output.
