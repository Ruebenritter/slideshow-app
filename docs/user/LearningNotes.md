# Learning Notes Reviewing Go Documentation and Guidelines

Roughly follows "../agent/LEARNING.md" structure. Main source is [Effective Go](https://go.dev/doc/effective_go).

## Go vs OOP / Go Naming section

**Getters:** For a private field `owner` Go suggest a Getter `Owner` without _Get_ as `owner := obj.Owner()` reads fine, but wants a setter `SetOwner` and says it reads fine. To me that doesn't seem consistent as i would either mark both consistently via keywords `GetVar/SetVar` or allow them to be understood from context/use:

> - var owner = obj.Owner // Getter
> - obj.Owner = value // Setter

C# auto properties seem much more intuitive to me. If `SetVar` is necessary, i will probably use `GetVar` for consistency or check whether a public field is appropriate.

**Interfaces:** one method interfaces end with -er: Reader, Writer etc.; ISlidable → Slider?

**if/switch:** accept init statement like 'for' → use to set up a local variable; early return idiomatic instead of if/else or use switch for if/else

**Short names:** single letter variables idiomatic in narrow scopes. Will probably refactor code closer to Csharp convention as reddit conversation has already shown that a `s.Images` citation is just not as readable as `slideshow.Images`.
Single letters are fine for loops, iterators or standardised letters for constants etc. Even if Go convention is more lenient.

## Pointers

Not much to it. Equivalent to C pointers or `ref` in C#.
We work on the original instead of a copy.

## Goroutine

Goroutine is comparable to C# async void. Slideshow window should be goroutine from my understanding. issue is the doubled go routine that creates a goroutine for slides as well but never closes them properly → leak

I believe i misused them as some form of event driven system. Goroutines are an important feature of Go so i can understand why i would use them to get my functionality but still unsure why i nested two of them. Even without knowing anything about goroutines, i should have detected the issue causing a leak as improper implementation of a recursive function. Or the attempt at a recursive function led to this weird hack.

## Channels

I created a stop signal when there was an explicit path for closing channels. A _send_ also does not reach all subscribers/recievers but only one.

`Select` can subscribe to multiple channels but will pick the first unblocked one and drop the others. I wasn't aware of that which is one reason why i implement the cancellation wrong.

`context package` provides functionality what stopChan was intended to do.

> Why use chan bool instead of chan struct{}?
> Question already highlighted in an inline comment in the legacy code. Can't remember the exact thought process from 2 years ago. Most likely a bool stop felt more familiar when trying to create a stop signal and probably lacked a good understanding of event types in C# itself (Action, Delegate etc).
> I knew what i wanted to do but used familiar tools to get it to work and hesitated to apply the idiomatic solution. (Hammer → Screw problem).

## Shared state

not looking into it yet as i'm not sure this will even be an issue if propper composition/architecture is applied.
Only one `object` needs to know the index to get the correct image. But this is still thinking within OOP. Still need to find out how to think in Go scripting terms.

## Go Rhythm

- Determine what processes run concurrently → separate into goroutines
- Determine state ownership → if multiple goroutines want to access the same field either use mutex to protect access or rethink goroutine split
- similar to DI, consumer defines need → interface communicates that need, any fitting provider can drive the consumer then (decoupling, testability)

# Consequence

**Agent refactor plan lists fixes by the severity of identified issues. However, most issues stem from designing within familiar OOP contrainsts working against the idiomatic Go flow. Functions are implemented with the first tool i found or understood instead of learning the intended solutions for the underlying problems i wanted to solve (like cancelling tasks).**

**My first refactor step will be structuring the project into better defined packages. Main() should then only assemble the app instead of containing the setup logic directly. This structure should help better identify ownership and needs to avoid similar mistakes in the future. It should also help navigating the code when fixing the critical bugs.**
