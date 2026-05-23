# Refactor Plan: `refactor/idiomatic-go`

Based on code review feedback. Goal: fix critical bugs and rewrite toward idiomatic Go.
Reference: [Issue #2](https://github.com/Ruebenritter/slideshow-app/issues/2)

---

> 1. Amendment: **Cursory review of current flow before changing anything.**
>    See: ..\FLOW_REVIEW.md
> 2. Amendment: **Correcting the existing flow to follow the intended Go flow should have priority over fixing single bugs to avoid new systematic hacks/issues.**

---

## Priority 1 — Critical Bugs

### 1. Fix goroutine leak in `NextSlide`

`NextSlide` calls `Start`, which spawns a new goroutine on every invocation. Old goroutines are never stopped.

- Remove `Start()` call from `NextSlide`
- `NextSlide` should only update state and send to the image channel
- The single goroutine loop started by `Start` handles all transitions

> **Go idiom:** A goroutine must always have a defined exit path. The owner of a goroutine is responsible for stopping it. Read: [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

### 2. Fix race condition on `CurrentIndex`

`CurrentIndex` is read and written from multiple goroutines without synchronization.

- Protect with `sync.Mutex` or make it accessible only through the goroutine that owns the loop
- Alternatively: remove `CurrentIndex` from the public API entirely and derive it inside the loop

> **Go idiom:** "Do not communicate by sharing memory; share memory by communicating." If state must be shared, use `sync.Mutex`. Run tests with `go test -race ./...`.

---

### 3. Fix double `StopChan` receiver

Both the goroutine in `Start` and the one in `showSlideshow` listen on `StopChan`. `Stop()` only sends once — one goroutine always leaks.

- Replace `StopChan` with `context.Context` and `cancel()` — cancellation fan-out is handled automatically
- Or: close the channel instead of sending (closing broadcasts to all receivers)

> **Go idiom:** Closing a channel is a broadcast. `context.Context` is the standard way to propagate cancellation across goroutines. Read: [context package docs](https://pkg.go.dev/context)

---

### 4. Fix blocking channel sends (deadlock risk)

`imageChan <- ...` and `progressChan <- ...` block if no receiver is ready — e.g. during shutdown.

- Use `select` with a `ctx.Done()` fallback on every send:
  ```go
  select {
  case s.imageChan <- img:
  case <-ctx.Done():
      return
  }
  ```

> **Go idiom:** Never send to a channel without a cancellation escape hatch in long-running goroutines.

---

## Priority 2 — Design / Architecture

### 5. Replace `StopChan chan bool` with `context.Context`

`chan bool` as a stop signal is non-idiomatic. Migrating to `context` also resolves issue #3.

- `NewSlideshow` accepts a `context.Context`
- `Stop` becomes `cancel()` called by the owner
- Remove `StopChan` from the struct entirely

> **Go idiom:** `chan struct{}` for signal-only channels (zero allocation). Prefer `context.Context` for cancellation that crosses API boundaries. Read: [Go Blog: Context](https://go.dev/blog/context)

---

### 6. Establish consistent visibility strategy

Current mix of public fields, private fields with getters, and public channels is arbitrary.

- Adopt Strategy B: all struct fields lowercase, public API via constructor and methods only
- `imageChan` and `progressChan` exposed as `<-chan T` (read-only) to callers
- Remove `SetImages` — configuration belongs in the constructor or a dedicated reset method
- Remove `IsPaused` if pause state is no longer relevant to callers after TogglePause is introduced

> **Go idiom:** Unexported fields + exported methods = encapsulation in Go. Exposing a channel as `<-chan T` signals read-only intent to the caller and is enforced by the compiler.

---

### 7. Rename `Pause` to `TogglePause`

The function toggles state — the name should reflect that.

> **Go idiom:** Function names should describe what they do, not what the caller intends.

---

### 8. Fix `NextSlide` abstraction

The modulo index calculation lives in the caller (`main.go`), not in `NextSlide`. The abstraction is leaky.

- `Next()` and `Prev()` methods with no index parameter
- Index wrapping handled internally

> **Go idiom:** Methods should encapsulate behavior. If the caller has to know internals to call a method correctly, the abstraction is wrong.

---

### 9. Fix timer drift

`elapsedTime += time.Second` accumulates error. Ticker fires are not exactly 1 second.

- Store `startTime time.Time` when a slide begins
- Calculate progress as `time.Since(s.startTime)`

> **Go idiom:** Use `time.Since` / `time.Until` for elapsed/remaining time. Never accumulate duration from ticker ticks.

---

## Priority 3 — Code Style

### 10. Rename `slideshowObj` in `main.go`

`slideshowObj` is Java-style. In Go, short, clear names are preferred.

- Rename to `ss` or `show`

> **Go idiom:** Go favors short variable names in narrow scopes. `Obj` suffix has no meaning in a language without mandatory OOP.

---

### 11. Sequential error handling in `main.go` start button

Current code collects two errors before checking. Idiomatic Go checks immediately after each fallible call.

```go
// before
timePerImage, err1 := ...
_, err2 := ...
if err1 != nil { ... }
if err2 != nil { ... }

// after
timePerImage, err := ...
if err != nil { ... }
amount, err := ...
if err != nil { ... }
```

---

## Validation

After all steps:

```bash
go test -race ./...        # catches race conditions
go vet ./...               # catches common mistakes
goleak in slideshow_test.go # verifies no goroutine leaks
```
