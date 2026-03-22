## Highly Opinionated Generator Example

This is an example of a highly opinionated generator from my [code generation toolbox](https://nhatp.com).
It has no configuration - all decisions are baked into the binary.

These are the opinions:

- Business logic should live in a struct named with suffix `Op`. Standard steps are
  `[validate] -> [authorize] -> [process] -> [handle]`.
- All `*Op` structs are composed together into a public interface `Service`.
- All enum `String()` methods are generated using line comments.
- The generated file is `codegen.go` and has the same package name as the source code.

### Usage

Clone the repository and run

~~~sh
# runs the generator against the included example project
go run ./main.go -w ./example

# runs the generator in dry-run mode against the included example project
go run ./main.go --dry-run -w ./example
~~~

### Contributing & License

PRs are welcome! Distributed under the Apache License 2.0.

---

If you like the project, feel free to [buy me a coffee](https://buymeacoffee.com/toniphan21). Thank you!
