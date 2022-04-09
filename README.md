# `depsize`

This module provides methods to determine the size of a Go module and its dependencies.

It does this by performing a HEAD request for a zip file the module's source
from the Go module proxy, and parsing the `Content-Length` header.

Modules' dependencies are not included in the returned size by default, including any
vendored dependencies that may exist in the repo.

The returned size is the compressed size according to the Go module proxy.

No content is actually fetched. No gophers are harmed in the determination of this information.

### Go package

```go
import "github.com/imjasonh/depsize"

...

 // Get latest version of the module.
latest, err := depsize.Latest("github.com/foo/bar")

 // Get size of module @ version.
size, err := depsize.Size("github.com/foo/bar", latest)

 // Get deps of the module, and their sizes.
deps, err := depsize.Deps("github.com/foo/bar", latest)

```

The version can be:
- `latest`, to get the latest release according to Go's semver rules
- a semver release (e.g., `v0.1.2`)
- a pseudoversion (e.g., `v0.1.3-0.20220110151055-d3adb33ffac3`)

### CLI

There's also a CLI, at `./cli`:

You can pass both the module name and version:

```sh
$ go run ./cli github.com/google/go-containerregistry v0.8.0
1933721
```

Or just the module name, and the latest version will be assumed:

```sh
$ go run ./cli github.com/google/go-containerregistry
1933721
```

You can also request human readable byte sizes, with `-h`:

```sh
$ go run ./cli -h github.com/google/go-containerregistry
1.9 MB
```

You can request a recursive lookup of a dep's transitive deps, with `-R`:

```sh
$ go run ./cli -R github.com/google/go-containerregistry/pkg/authn/k8schain
github.com/google/go-containerregistry/pkg/authn/k8schain v0.0.0-20220328141311-efc62d802606 80033
    github.com/Azure/azure-sdk-for-go v62.0.0+incompatible 67808264
    github.com/aws/aws-sdk-go-v2 v1.14.0 8954805
    golang.org/x/text v0.3.7 8610883    ...
total 118602041
```

Deps are sorted by size, largest at the top.