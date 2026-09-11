# Changelog

All notable changes to Flux are documented in this file.

## 0.1.7

### Added

- Explicit `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`, and
  `Handle` route APIs.
- Generic `Body`, `QueryParams`, and `PathParams` parsing with validation.
- Generic `JSONHandler` and `JSONBodyHandler` endpoint adapters.
- Route metadata options for names, descriptions, requests, and responses.
- Standard `context.Context` interoperability on every request.
- Safe defaults through `DefaultConfig` and support for `flux.New(nil)`.
- Framework version metadata and `flux --version` support.
- CI coverage for the minimum supported Go version and the latest stable Go on
  Linux, macOS, and Windows.

### Changed

- Runtime route-file generation is opt-in with `Config.GenerateRouteFiles`.
- Starting an app no longer creates a plugin directory when one is not present.
- Unsupported native-plugin platforms such as Windows can build and use the
  rest of Flux; `.so` loading remains available on Linux, macOS, and FreeBSD.
- Unknown handler errors return a production-safe 500 response.
- Controller handlers now receive a context connected to their application.
- Optional database and authentication accessors return `nil` when disabled.
- Generated projects target Go 1.23 and Flux 0.1.7.
- Removed the pinned toolchain directive so any compatible Go 1.23+ toolchain
  can build the framework without an automatic toolchain download.

### Fixed

- `AppError` now supports `errors.Is` and `errors.As` through `Unwrap`.
- Error detail builders no longer mutate shared sentinel errors across requests.
- Invalid controller values no longer cause reflection panics.
- Resource-named controller actions such as `HelloController.HandleGetHello`
  now map to `/hello` instead of `/hello/hello`.
- The Redis integration suite skips cleanly when Redis is not available.
- The root import smoke test is now a proper `_test.go` file, allowing
  `go build ./...` to build every package.

### Migration

Existing controller applications continue to compile. Applications that depend
on generated route files at runtime should set `GenerateRouteFiles: true`; new
applications should use explicit route registration.
