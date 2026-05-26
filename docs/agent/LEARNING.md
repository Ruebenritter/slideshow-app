# Learning Resources for Idiomatic Go Refactor

Topics are ordered by dependency — each section unlocks understanding for the next.
Each entry maps to one or more items in `REFACTOR_PLAN.md`.

---

## 1. Go Philosophy vs OOP

> Relevant to: all Priority 2 items, general structure of `slideshow.go`

Go is not object-oriented in the C#/Java sense. There are no classes, no inheritance, and no `this`.
The closest mental model is closer to Python modules + structs, but with explicit method receivers.

| C# / Java concept      | Go equivalent                                   |
| ---------------------- | ----------------------------------------------- |
| `class Slideshow { }`  | `type Slideshow struct { }`                     |
| `this.field`           | `s.field` (named receiver)                      |
| `new Slideshow(...)`   | `NewSlideshow(...)` constructor function        |
| `interface ISlideable` | `interface { Method() }` — implicit, structural |
| inheritance            | composition (embed structs)                     |
| `private` / `public`   | lowercase / Uppercase first letter              |

- [Effective Go](https://go.dev/doc/effective_go) — the authoritative style and philosophy guide. Read the _Names_, _Methods_, and _Interfaces_ sections first.
- [Go FAQ: Is Go object-oriented?](https://go.dev/doc/faq#Is_Go_an_object-oriented_language)
- [A Tour of Go — Methods and Interfaces](https://go.dev/tour/methods/1)

---

## 2. Pointers in Go

> Relevant to: receiver types on `Slideshow` methods, passing `*Slideshow` around

Go pointers work like C/C++ in syntax (`*T`, `&x`) but with critical differences:

- No pointer arithmetic
- The garbage collector manages memory — no `malloc`/`free`
- The compiler decides stack vs heap allocation (`new` exists but is rarely needed)
- Method receivers: `func (s *Slideshow)` mutates the original; `func (s Slideshow)` works on a copy

The reason `*Slideshow` is passed around the app is exactly the same as passing a `ref` parameter or a pointer in C++ — you want one shared instance, not copies.

- [A Tour of Go — Pointers](https://go.dev/tour/moretypes/1)
- [Effective Go — Pointers vs Values](https://go.dev/doc/effective_go#pointers_vs_values)

---

## 3. Goroutines and Goroutine Ownership

> Relevant to: Plan item #1 (goroutine leak in `NextSlide`)

A goroutine is _not_ a thread and is _not_ a coroutine in the Python/C# sense. It is a lightweight function scheduled by the Go runtime. `go f()` starts one and returns immediately.

The critical rule that the current code violates:

> **The goroutine that starts a goroutine is responsible for stopping it.**

Every goroutine must have a defined exit path. If `Start()` spawns a goroutine, `Start()` (or its owner) must guarantee that goroutine eventually returns.

- [Go Blog: Concurrency is not Parallelism](https://go.dev/blog/waza-talk)
- [Go Blog: Go Concurrency Patterns (Rob Pike)](https://go.dev/blog/io2012-videos) — look for the "Go Concurrency Patterns" talk
- [Effective Go — Goroutines](https://go.dev/doc/effective_go#goroutines)
- [Go Blog: Pipelines and Cancellation](https://go.dev/blog/pipelines) _(referenced in Plan item #1)_

---

## 4. Channels

> Relevant to: Plan items #3, #4, #6 — `StopChan`, `imageChan`, `progressChan`

Channels are typed conduits. Key rules relevant to this project:

| Concept           | Meaning                                                           |
| ----------------- | ----------------------------------------------------------------- |
| `make(chan T)`    | unbuffered — send blocks until receiver is ready                  |
| `make(chan T, n)` | buffered — send blocks only when buffer is full                   |
| `close(ch)`       | **broadcast** — all receivers unblock immediately with zero value |
| `<-chan T`        | read-only channel type — compiler enforces caller cannot send     |
| `chan<- T`        | write-only channel type                                           |

The current `StopChan chan bool` uses a _send_ for cancellation. This only unblocks **one** receiver. Closing a channel unblocks _all_ receivers — that is the correct pattern for shutdown signals.

- [A Tour of Go — Channels](https://go.dev/tour/concurrency/2)
- [A Tour of Go — Range and Close](https://go.dev/tour/concurrency/4)
- [A Tour of Go — Select](https://go.dev/tour/concurrency/5)
- [Effective Go — Channels](https://go.dev/doc/effective_go#channels)

---

## 5. The `select` Statement

> Relevant to: Plan items #3, #4 — escape hatches on channel sends

`select` is like a `switch` for channel operations. It blocks until one case is ready, then executes it. If multiple cases are ready simultaneously, one is chosen at random.

The pattern used in the refactor plan for safe sends:

```go
select {
case s.imageChan <- img:
case <-ctx.Done():
    return
}
```

Without the `ctx.Done()` case, a send to an unbuffered channel blocks forever if the receiver has already exited — a deadlock.

- [A Tour of Go — Select](https://go.dev/tour/concurrency/5)
- [Go Spec — Select statements](https://go.dev/ref/spec#Select_statements)

---

## 6. `context.Context` — Cancellation and Timeout

> Relevant to: Plan items #3, #5 — replacing `StopChan`

`context.Context` is the idiomatic Go way to propagate cancellation, deadlines, and request-scoped values across goroutine boundaries. It solves exactly the split-receiver problem in the current code.

```go
ctx, cancel := context.WithCancel(context.Background())
// pass ctx to anything that should stop when cancel() is called
// calling cancel() closes ctx.Done() — a broadcast to all goroutines holding ctx
cancel()
```

The key advantage over a `chan bool`: `cancel()` can be called multiple times safely and broadcasts to every goroutine that holds the same context.

- [Go Blog: Context](https://go.dev/blog/context) _(referenced in Plan item #5)_
- [pkg.go.dev — context package](https://pkg.go.dev/context) _(referenced in Plan item #3)_
- [A Tour of Go — Context (exercise)](https://go.dev/tour/concurrency/9)

---

## 7. Shared State and `sync.Mutex`

> Relevant to: Plan item #2 — `CurrentIndex` race condition

When multiple goroutines read and write the same variable, accesses must be synchronized. Go's `sync.Mutex` is the standard tool.

```go
type Slideshow struct {
    mu           sync.Mutex
    currentIndex int
}

func (s *Slideshow) index() int {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.currentIndex
}
```

The Go race detector (`go test -race`) will flag unprotected concurrent accesses automatically.

- [A Tour of Go — sync.Mutex](https://go.dev/tour/concurrency/9)
- [pkg.go.dev — sync package](https://pkg.go.dev/sync)
- [Go Blog: Race Detector](https://go.dev/blog/race-detector)
- [Go Blog: Share Memory by Communicating](https://go.dev/blog/codelab-share) — explains _when_ to use channels vs mutex

---

## 8. Go Naming and Visibility

> Relevant to: Plan items #6, #7, #10

Go visibility is purely by case: `Exported` (uppercase) is public; `unexported` (lowercase) is package-private. There are no `private`, `protected`, or `internal` keywords.

Naming conventions that differ from C#/Java:

- No `I` prefix on interfaces (`Reader` not `IReader`)
- No `Obj`, `Manager`, `Helper` suffixes
- Short names in narrow scopes (`s`, `ss`, `i`) are idiomatic, not lazy
- Acronyms are all-caps: `URL`, `ID`, not `Url`, `Id`
- Constructor functions: `NewTypeName(...)` returning `*TypeName`

- [Effective Go — Names](https://go.dev/doc/effective_go#names)
- [Go Blog: Package Names](https://go.dev/blog/package-names)
- [Go Code Review Comments — naming](https://go.dev/wiki/CodeReviewComments#variable-names)

---

## 9. Time: Timers, Tickers, and Drift

> Relevant to: Plan item #9 — `elapsedTime` accumulation

`time.Ticker` fires approximately every interval — the operative word is approximately. Summing tick counts introduces drift. The correct pattern for elapsed time is:

```go
start := time.Now()
// later:
elapsed := time.Since(start)
```

| Type            | Use case                           |
| --------------- | ---------------------------------- |
| `time.Timer`    | fire once after a duration         |
| `time.Ticker`   | fire repeatedly at an interval     |
| `time.Since(t)` | elapsed time since a fixed point   |
| `time.Until(t)` | remaining time until a fixed point |

- [pkg.go.dev — time package](https://pkg.go.dev/time)
- [Go Blog: Timer and Ticker patterns](https://go.dev/doc/effective_go#goroutines) — see the timeout and ticker examples

---

## 10. Error Handling

> Relevant to: Plan item #11 — sequential error checks in `main.go`

Go has no exceptions. Functions return `(value, error)`. The idiomatic pattern is to check immediately after each fallible call, not batch them:

```go
// idiomatic
duration, err := time.ParseDuration(input + "s")
if err != nil {
    // handle and return
}

amount, err := strconv.Atoi(amountInput)
if err != nil {
    // handle and return
}
```

Reusing `err` with `:=` is valid as long as at least one variable on the left is new.

- [A Tour of Go — Errors](https://go.dev/tour/methods/19)
- [Effective Go — Errors](https://go.dev/doc/effective_go#errors)
- [Go Blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)

---

## Reading Order Recommendation

1. [Effective Go](https://go.dev/doc/effective_go) — skim fully, read _Names_, _Methods_, _Concurrency_ sections carefully
2. [A Tour of Go — Concurrency module](https://go.dev/tour/concurrency/1) — interactive, covers goroutines, channels, select, mutex
3. [Go Blog: Share Memory by Communicating](https://go.dev/blog/codelab-share) — sets the mental model for section 7
4. [Go Blog: Pipelines and Cancellation](https://go.dev/blog/pipelines) — directly maps to the goroutine ownership problem in this app
5. [Go Blog: Context](https://go.dev/blog/context) — read after pipelines, it is the modern solution to the same problem
