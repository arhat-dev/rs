# rs

[![CI](https://github.com/arhat-dev/rs/workflows/CI/badge.svg)](https://github.com/arhat-dev/rs/actions?query=workflow%3ACI)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/rs)](https://pkg.go.dev/arhat.dev/rs)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/rs)](https://goreportcard.com/report/arhat.dev/rs)
[![Coverage](https://badge.arhat.dev/sonar/coverage/arhat-dev_rs?branch=master&token=563ff8cf318c9303285b4dd4eeb0c660)](https://sonar.arhat.dev/dashboard?id=arhat-dev_rs)

`R`endering `S`uffix Support for yaml

## What?

Before we start, let's agree on the concept `renderer` is just a simple function that takes a input and generates a output.

```text
input -> [ renderer ] -> output
```

A plain yaml doc looks like this:

```yaml
foo: bar
```

Everything's static, once unmarshaled, `foo` gets value `bar`.

Now what if we would like to take an environment variable as `foo`'s value?

```yaml
foo: ${FOO}
```

you will only get `${FOO}` after unmarshaling, you have to code you own logic to map `${FOO}` to some system environment variable.

What you code, is actually a `renderer` with its usage specific to `foo`.

Rendering suffix is there to help, it offers a way to make your yaml doc dynamic on its own, you control what you get in yaml rather than your compiled code.

### How it looks?

Let's continue with the environment variable support, say we have developed a renderer `env` (which calls `os.ExpandEnv` to map env reference to its value)

```yaml
foo@env: ${FOO}
```

Before I would start, probably you have already figured out what's going on:

`@` is like a notifier, and notifies the renderer `env` with value `${FOO}`, then `env` does its own job and generates some output with `os.ExpandEnv` function call, and at last, the output is set as `foo`'s value.

Simple and straightforward, right?

### Why?

A piece of software configuration is like a piece of contract between developer and end user: set this, do that.

For very long time, both writing and resolving software configuration are tedious and error prone, especially when you want to achieve felixibility to some extent.

Let's discuss with a simple example, take target config as this

```go
type Config struct {
    A string
    B int
    C float64

    // RemoteHostScript for remote host execution
    RemoteHostScript string `yaml:"remote_host_script"`
}
```

What whould you do when you want to support environment variables for its fields `A`, `B`, `C`, while not `remote_host_script` (obviously the `remote_host_script` should be executed in some remote system).

```yaml
a: ${ENV_FOR_A}
b: ${ENV_FOR_B}
c: ${ENV_FOR_C}

remote_host_script: |-
  #!/bin/sh

  echo ${ENV_WITH_DIFFERENT_CONTEXT}
```

- Solution 1: Make `remote_host_script` a file path reference (or add a new field like `remote_host_script_file`)?
  - Simple and effective, but now you have two source of remote host script, more fixed code logic added, more document to come for preferred options when both is set. End user MUST read your documentation in detail for such subtle issues (if you are good at documenting).
- Solution 2: Expand environment variables before unmarshaling?
  - What would you do with `${ENV_WITH_DIFFERENT_CONTEXT}`? Well, you can unmarshal the config into two config, one with environment variables expanded, another not, and merge them into one.
  - Looks like a effective solution, but now you have to do the merge in fixed code.
- Solution 3: Unmarshal yaml as `map[string]interface{}` first, then do custom handling for every field?
  - Now you have to work with types manually, tedious yet error prone job starts now.
- Solution 4: Create a new DSL, add some keywords...
  - We have already seen so many DSLs created just for configuration purpose, almost none of them really simplified the configuration management, and usually only useful for development not deployment.

As developers, what we actually need is let end user decide which field is resolved by what method, and we just control when to resolve which field.

Rendering suffix is applicable to every single yaml field, doing exactly what end user need, and it can resolve fields partially with certain strategy in code, which is exactly what developers want.

Rendering suffix does one thing to make both end users and developers happy:

Make your configuration flexible while still intuitive

- Developers don't define how to resolve, but only control when to resolve
- Provide end user easy to understand renderers, rather than confine their creativity to certain resolving rules.

Now, think again, would gradle be a better tool using dynamic yaml instead of Groovy DSL? Could GitOps for kubernetes be more easier with rendering suffix yaml instead of special purpose built kustomize or helm?

## Prerequisites

- Use `gopkg.in/yaml.v3` for yaml marshaling/unmarshaling

## Features

- Field specific rendering: customizable in-place yaml data rendering
  - Add a suffix starts with `@`, followed by some renderer name, to your yaml field name as rendering suffix (e.g. `foo@<renderer-name>: bar`)
- Type hinting: keep you data as what it supposed to be
  - Add a type hint suffix `?<some-type>` to your renderer (e.g. `foo@<renderer-name>?[]obj: bar` suggests `foo` should be an array of objects using result from `<renderer-name>` generated with input `bar`)
  - [list of supported type hints](https://github.com/arhat-dev/rs/blob/v0.4.0/typehint.go#L24))
- Data merging and patching made esay: create patching spec in yaml doc
  - Add a patching suffix `!` to your renderer (after the type hint if any), feed it a [patch spec](https://pkg.go.dev/arhat.dev/rs#PatchSpec) object
- Renderer chaining: render you data with a rendering pipeline
  - join you renderers with pipes (`|`), get your data rendered through the pipeline (e.g. join three renderers `a`, `b`, `c` -> `a|b|c`)
- Supports arbitraty yaml doc without type definition in code
  - Use [`AnyObject`](https://pkg.go.dev/arhat.dev/rs#AnyObject) as `interface{}`
  - Use [`AnyObjectMap`](https://pkg.go.dev/arhat.dev/rs#AnyObjectMap) as `map[string]interface{}`
- Vanilla yaml, valid for all standard yaml parser
- Everything `gopkg.in/yaml.v3` supports are supported
  - Anchors, Alias, YAML Merges

Sample YAML Doc with all features above

```yaml
foo@a!: &foo
  value@env|http?[]obj: https://example.com/${FOO_FILE_PATH}
  merge:
  - value@file: ./value-a.yml
    select: |-
      [ .[] | sort ]
  patch:
  - { op: remove, path: /0/foo }
  - op: add
    path: /0/foo
    value: "bar"
    select: '. + "-suffix-from-jq-query"'
  select: |-
    { foo: .[0].foo, bar: .[1].bar }

bar@a!: *foo
```

__NOTE:__ This module provides no renderer implementation, and the only built-in renderer is a pseudo renderer with empty name that skips rendering (output is what input is) for data patching and type hinting purpose (e.g. `foo@?int!: { ... patch spec ... }`). You have to roll out your own renderers.

## Usage

See [example_test.go](./example_test.go)

__NOTE:__ You can find more examples in [`arhat-dev/dukkha`](https://github.com/arhat-dev/dukkha)

## Known Limitations

See [known_limitation_test.go](./known_limitation_test.go) for sample code and workaround.

- Built-in map data structure with rendering suffix applied to map key are treated as is, rendering suffix won't be recognized.
  - which means for `map[string]interface{}`, `foo@foo: bar` is just a map item with key `foo@foo`, value `bar`, no data to be resolved

## How it works?

IoC (Inversion of Control) is famous for its application in DI (Dependency Injection), but what it literally says is to control the outside world at somewhere inside the world, that's the start point.

Custom yaml unmarshaling requires custom implementation of `yaml.Unmarshaler`, so usually you just get your structs unmarshaling by go-yaml directly.

`BaseField` lives in your struct as a embedded field, all its methods are exposed to the outside world by default, and of course `BaseField` implements `yaml.Unmarshaler`.

Can you control how sibling fields of a struct? not possible unless with the help of outside world, that's why we need `Init()` function, calling `Init()` with your struct actually activates the `BaseField` in it, `Init()` function tells the inner `BaseField` what fields the parent struct have, and how to unmarshal yaml data to them, with the help of reflection.

You only have to call `Init()` once for the top level struct, `BaseField` in it knows what to do with sibling fields, it will search for all structs with `BaseField` embedded when unmarshaling, call `Init()` for them, until the yaml doc is unmarshaled in this recursive fashion.

Basically `BaseField` handles everything related to yaml unmarshaling to support rendering suffix after initial activation, so all you need to do is to embed a `BaseField` as the very first field in your struct.

## LICENSE

```txt
Copyright 2021 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
