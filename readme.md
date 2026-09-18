# Vinego

Vinego is a new set of linters built on the Go `analysis.Analyzer` framework. Many of the linters focus on increasing strictness around variable initialization and type safety. They fall somewhere in-between "drop these into your codebase with no changes" and "rewrite all your code to conform to weird new language-extra conventions" in terms of invasiveness.

Each linter is an `analysis.Analyzer` and they're aggregated as a `golangci-lint` plugin. Individual non-opt-in linters can be enabled via the `golangci-lint` settings.  The analyzers should work with any `analysis.Analyzer` framework though.

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

## All-in-one development container

1. We provide a pre-made Docker batteries-included image for CI and development environments: `ghcr.io/upsun/vinego:latest`

   It includes

   - Go
   - `golangci-lint` built with `vinego` included as a `module`-type plugin
   - `gci`
   - `goimports`
   - `dlv`
   - `staticcheck`

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
            type: "module"
            settings:
               enable_varinit: true
               enable_explicitcast: true
               enable_capturederr: true
   ```

1. Run the linters with `docker run --rm --volume $PWD:/mnt --workdir /mnt vinego /bin/golangci-lint run --verbose`. You should see `vinego` listed in the output.

## Building your own golangci-lint

If you need an image with different tools or want to use the linter in some other situation, we recommend using golangci-lint's [module](https://golangci-lint.run/plugins/module-plugins/) system to bootstrap a new golangci-lint with the vinego linters included.

1. Install any recent version of golangci-lint

1. Create `.custom-gcl.yml` with:

   ```yaml
   version: v1.64.6
   plugins:
     - module: 'github.com/upsun/vinego/src'
       import: 'github.com/upsun/vinego/src'
       version: latest
   ```

   The top-level version is the version of golangci-lint that the process will bootstrap - it doesn't need to be the same version as the golangci-lint you installed in (1.).

1. Run `golangci-lint custom -v`

   This will produce a _new_ `golangci-lint`

1. Use the new `golangci-lint` with this `.golangci.yml`:

   ```yaml
   linters-settings:
     custom:
       vinego:
         type: "module"
         settings:
           enable_varinit: true
           enable_explicitcast: true
           enable_capturederr: true
   linters:
     enable:
       - vinego
   ```
# Releases

Dependency updates and releases are automated:

1. Dependabot opens weekly grouped pull requests for the Go modules
   (`src/go.mod`), the `FROM golang:...` base images (`src/Dockerfile`) and the
   pinned GitHub Actions.
1. `sync-pins` covers what Dependabot has no manager for: the `golangci-lint`
   version and the `go` directive. It runs weekly, and again whenever a new base
   image lands on `main`, raising its pull request the same way.
1. The `ci` workflow builds `src/Dockerfile` on the branch - building the custom
   `golangci-lint`, linting this repo with it and running the tests - and its
   `pins` job fails if the versions below have drifted apart.
1. `automerge` merges the pull request only once ci has succeeded. A failing
   build leaves it open instead.
1. Anything landing on `main` that changes `src/**` cuts the next patch tag, and
   `ci` publishes `ghcr.io/upsun/vinego` for it.

To cut a larger version by hand, run the `release` workflow from the Actions tab
and choose `minor` or `major`.

## The version pins

| Pin | Where | Maintained by |
| --- | --- | --- |
| Go modules | `src/go.mod`, `src/go.sum` | Dependabot |
| Base images | `src/Dockerfile` (`FROM golang:X.Y`) | Dependabot |
| Action versions | `.github/workflows/*.yml` | Dependabot |
| `golangci-lint` | `src/.custom-gcl.yml` **and** `src/Dockerfile` | `sync-pins` |
| Go language version | `src/go.mod` (the `go` directive) | `sync-pins` |

The two `golangci-lint` values must always match each other, or
`golangci-lint custom` bootstraps a binary that disagrees with the plugin, and
the `go` directive must never ask for more than the base image provides. The
`pins` job enforces both on every push, so a half-updated set cannot be merged.
