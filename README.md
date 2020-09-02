# zapfilter

:smile: zapfilter

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/moul.io/zapfilter)
[![License](https://img.shields.io/badge/license-Apache--2.0%20%2F%20MIT-%2397ca00.svg)](https://github.com/moul/zapfilter/blob/master/COPYRIGHT)
[![GitHub release](https://img.shields.io/github/release/moul/zapfilter.svg)](https://github.com/moul/zapfilter/releases)
[![Made by Manfred Touron](https://img.shields.io/badge/made%20by-Manfred%20Touron-blue.svg?style=flat)](https://manfred.life/)

[![Go](https://github.com/moul/zapfilter/workflows/Go/badge.svg)](https://github.com/moul/zapfilter/actions?query=workflow%3AGo)
[![Release](https://github.com/moul/zapfilter/workflows/Release/badge.svg)](https://github.com/moul/zapfilter/actions?query=workflow%3ARelease)
[![PR](https://github.com/moul/zapfilter/workflows/PR/badge.svg)](https://github.com/moul/zapfilter/actions?query=workflow%3APR)
[![GolangCI](https://golangci.com/badges/github.com/moul/zapfilter.svg)](https://golangci.com/r/github.com/moul/zapfilter)
[![codecov](https://codecov.io/gh/moul/zapfilter/branch/master/graph/badge.svg)](https://codecov.io/gh/moul/zapfilter)
[![Go Report Card](https://goreportcard.com/badge/moul.io/zapfilter)](https://goreportcard.com/report/moul.io/zapfilter)
[![CodeFactor](https://www.codefactor.io/repository/github/moul/zapfilter/badge)](https://www.codefactor.io/repository/github/moul/zapfilter)


## Usage

```go
import "moul.io/zapfilter"

func ExampleParseRules() {
	core := zap.NewExample().Core()
	// *=myns             => any level, myns namespace
    // info,warn:myns.*   => info or warn level, any namespace matching myns.*
	// error=*            => everything with error level
	logger := zap.New(zapfilter.NewFilteringCore(core, zapfilter.MustParseRules("*:myns info,warn:myns.* error:*")))
	defer logger.Sync()

	logger.Debug("top debug")                                 // no match
	logger.Named("myns").Debug("myns debug")                  // matches *:myns
	logger.Named("bar").Debug("bar debug")                    // no match
	logger.Named("myns").Named("foo").Debug("myns.foo debug") // no match

	logger.Info("top info")                                 // no match
	logger.Named("myns").Info("myns info")                  // matches *:myns
	logger.Named("bar").Info("bar info")                    // no match
	logger.Named("myns").Named("foo").Info("myns.foo info") // matches info,warn:myns.*

	logger.Warn("top warn")                                 // no match
	logger.Named("myns").Warn("myns warn")                  // matches *:myns
	logger.Named("bar").Warn("bar warn")                    // no match
	logger.Named("myns").Named("foo").Warn("myns.foo warn") // matches info,warn:myns.*

	logger.Error("top error")                                 // matches error:*
	logger.Named("myns").Error("myns error")                  // matches *:myns and error:*
	logger.Named("bar").Error("bar error")                    // matches error:*
	logger.Named("myns").Named("foo").Error("myns.foo error") // matches error:*

	// Output:
	// {"level":"debug","logger":"myns","msg":"myns debug"}
	// {"level":"info","logger":"myns","msg":"myns info"}
	// {"level":"info","logger":"myns.foo","msg":"myns.foo info"}
	// {"level":"warn","logger":"myns","msg":"myns warn"}
	// {"level":"warn","logger":"myns.foo","msg":"myns.foo warn"}
	// {"level":"error","msg":"top error"}
	// {"level":"error","logger":"myns","msg":"myns error"}
	// {"level":"error","logger":"bar","msg":"bar error"}
	// {"level":"error","logger":"myns.foo","msg":"myns.foo error"}
}
```

More examples on https://pkg.go.dev/moul.io/zapfilter

## Install

### Using go

```console
$ go get -u moul.io/zapfilter
```

### Releases

See https://github.com/moul/zapfilter/releases

## Contribute

![Contribute <3](https://raw.githubusercontent.com/moul/moul/master/contribute.gif)

I really welcome contributions. Your input is the most precious material. I'm well aware of that and I thank you in advance. Everyone is encouraged to look at what they can do on their own scale; no effort is too small.

Everything on contribution is sum up here: [CONTRIBUTING.md](./CONTRIBUTING.md)

### Contributors ✨

<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg)](#contributors)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="http://manfred.life"><img src="https://avatars1.githubusercontent.com/u/94029?v=4" width="100px;" alt=""/><br /><sub><b>Manfred Touron</b></sub></a><br /><a href="#maintenance-moul" title="Maintenance">🚧</a> <a href="https://github.com/moul/zapfilter/commits?author=moul" title="Documentation">📖</a> <a href="https://github.com/moul/zapfilter/commits?author=moul" title="Tests">⚠️</a> <a href="https://github.com/moul/zapfilter/commits?author=moul" title="Code">💻</a></td>
    <td align="center"><a href="https://manfred.life/moul-bot"><img src="https://avatars1.githubusercontent.com/u/41326314?v=4" width="100px;" alt=""/><br /><sub><b>moul-bot</b></sub></a><br /><a href="#maintenance-moul-bot" title="Maintenance">🚧</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

### Stargazers over time

[![Stargazers over time](https://starchart.cc/moul/zapfilter.svg)](https://starchart.cc/moul/zapfilter)

## License

© 2020 [Manfred Touron](https://manfred.life)

Licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) ([`LICENSE-APACHE`](LICENSE-APACHE)) or the [MIT license](https://opensource.org/licenses/MIT) ([`LICENSE-MIT`](LICENSE-MIT)), at your option. See the [`COPYRIGHT`](COPYRIGHT) file for more details.

`SPDX-License-Identifier: (Apache-2.0 OR MIT)`
