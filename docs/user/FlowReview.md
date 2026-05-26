# App flow review before refactor

## Assumption from quick scan

// Startup

1. app entry point is main.go.
2. main sets up UI widgets in the startup window and sets default or init values for input fields as well as presets like the duration options. // preset values are offered as buttons
3. main then binds input elements to handler (startButton) which validates input.
4. widgets are packed in a layout container, window sized and the app runs.
   // Action phase
5. We now expect user input for:
   - reference images directory
   - wait before next slide duration
   - total slides count
     All input fields start with empty invalid values! → QoL change after cricital bug fixes could be valid default values for quick start

6. User needs to confirm input with pressing start → QoL change post refactor: input validation with key input delay possible? user sees problems before needing to confirm
7. Input validation on startButton press with early return if clauses // check whether Go keywords like goto/loop labels allow for a simpler solution (lowest priority)
8. We also try to get images from the selected directory and validate for existance and count, then shuffle for random slide order
9. NewSlideshow calls Go constructor in slideshow.go.
10. Ctr initializes slideshow parameters, exposing the Images array/list, SlideDuration, CurrentIndex and StopChan
    // slideshow is setup as a "self-sufficient" controller; something irks me when reviewing the script: something about how the data and logic are setup in this "class" feels wrong.
    // if redone with my current MVVM/programming experience i would separate slideshow data (images (contains count), duration) from viewModel business logic (channels) and data (progress related variables).
    // i also can't tell (yet) why SlideDuration and Images are public if we set them via cstr parameters (init) and handle the slideshow operation ourselves
    // Why do we need a separate Image Setter if every slideshow is a new "object"?
11. we then call showSlideShow pointing to our new slideshow "object" // i can fully understand why i would name it an object if it comes from a constructor not bothering too much with the "finesse" of pointers/references/objects definition
12. We build a new Window and setup UI widgets again. // i think now i see why certain things are public. UI is built in main taking data from the slideshow even though we owned that data originally
    // i would think i was trying to build decoupled, clean code where showSlideshow just needs a valid slideshow "object" and is independent of main since these functions build separate windows
    // with MVVM, MVC in mind i probably understood these as my viewmodels but it seems some viewmodel logic also bled into slideshow.go; it certainly looks like a grown project than a properly designed one
13. Again we bind our input elements (only buttons this time for slideshow control)
14. package, resize and show
15. Then we start slideshow with a goroutine. // Why?
16. We call start on our slideshow "object" passing control of the slideshow function but bind our desired user actions (skip, pause, etc), progress callback and image output to channels
    // Again this setup irks me as i feel like responsibilities are not clearly divided/separated.
    // A goroutine seems generally fine as it is Go's way to handle concurrency, not necessarily parallelism. Not sure if there are Go alternatives but it seems to be the correct approach
    // to an event driven system combining the sequential timer based slideshow with (async) user input to override and pause the process. Again slideshow acts like a viewmodel when it kinda isn't as bindings point to showSlideshow/main.go. While i already see issues (or room for improvement) with my coding experience now, i still need to research the correct Go approach as MVVM lens might also be misleading.
17. a signal from StopChan seems to be ignored/dropped → Why?
18. Start() resets the timer (and ticker). we bind (select) to channels to listen to events. Again StopChan is ignored/dropped.
    Timer triggers NextSlide while ticker updates the progress bar. // i initially wondered why there are two timer components. they supply different needs. not sure if this is a hack or necessary from design
19. NextSlide(index) emits the next image and calls Start( ) // immediately clear this is a endless loop risk and improper recursive approach (recursion would call itself with a clear end condition)
    // funcs are setup up in a OOP style so methods that would be implemented in the Slideshow class in C#/Java (ignored the problematic separation of concerns) are grouped together in a .go file.
    // but all these methods being comparable to static methods that only interact with the received args (a pointer to itself basically) already "scream" this is not the intended pattern.
    // GPT-4 as external chat definitely did not have the insight into the project as current embedded assitants but i must have been very strict with the questions/prompts passed to the LLM that i was able
    // to remain in my OOP structure this much. I know i was controlled/careful with AI help to make sure i would do it myself in a way that i would expect to offer learning experience.
    // I understand how this well meant approach created so much friction and frustration that i was happy once the app worked and did not review/refactor the result afterwards with the "lessons learned".

## Important things i missed

### 1. Point 19 is worse than a recursion risk — it's an active goroutine leak (critical bug)

`NextSlide` calls `Start()`, which spawns a **new goroutine** and returns immediately. The old goroutine from the previous `Start()` call is never stopped before the new one launches. That old goroutine keeps running, its timer will eventually fire, it calls `NextSlide` again, which calls `Start()` again — spawning yet another goroutine. The goroutine count grows with every slide advance. This is not recursion (the call stack does not grow), it is unbounded goroutine accumulation. This is the most critical bug.

// You are right. I missed the second go keyword as i focused on the channels immediately. That is an obvious error which leaks forgotten goroutines.
// I assume since we always point towards the same slideshow we dont get artifacts inside the visual slideshow as long as the time drift doesnt take effect, since we simply duplicate the "bridge"
// to the slideshow. Any coroutine pausing will have the same effect for all siblings. Slideshow having control over the slideshow state might have been a hack to combat the goroutine duplicate issues
// or simply be an unlucky error masking a bigger one.

### 2. Point 17 — the StopChan has two consumers but only one signal (critical bug)

There are two goroutines consuming `StopChan`:

- the inner goroutine inside `Start()` (`case <-s.StopChan: return`)
- the outer `for/select` loop in `showSlideshow` (`case <-slideshowObj.StopChan: return`)

`Stop()` only sends one value (`s.StopChan <- true`). Whichever goroutine happens to receive it wins — the other is never unblocked and leaks. Because of the goroutine accumulation in point 1, there can be many goroutines all listening on `StopChan`, making stop behaviour increasingly unpredictable as the slideshow runs longer.

// Yes, i happened to read this scanning the Go documentation beforehand that while Select allows us to subscribe to multiple channels only the first unblocked one is handled.

### 3. ~~Setup window is never closed when the slideshow starts~~ — intentional design, not a bug

When `showSlideshow` is called from `main()`, the original setup window stays open alongside the slideshow window. There is no `w.Hide()` or `w.Close()` call on the transition. The user ends up with two open windows.

// That was kind of intentional as testing and actual use led to opening several slideshows with different parameters. Restarting the app for a new one would be too bothersome. Hiding it during the slideshow
// would have been an elegant solution, but obviously not a critical feature for the app's function.

Confirmed intentional. Hiding the setup window during an active slideshow would be a minor QoL improvement, but the current behaviour is not a bug.

### 4. ~~Pause button label logic is inverted~~ — not a bug, correct "next action" convention

In `showSlideshow` the condition reads:

```go
if slideshowObj.IsPaused() {
    slideshowObj.Pause()
    pauseButton.SetText("Pause")   // should be "Resume"
} else {
    slideshowObj.Pause()
    pauseButton.SetText("Resume")  // should be "Pause"
}
```

When the slideshow is already paused, the button sets its own text to "Pause" instead of "Resume", and vice versa. The label assignments are swapped relative to the resulting state.

// Not a bug as that is probably intentional. I'm constantly frustrated with buttons in different apps since there seems to be no dominating norm what button text describes.
// Sometimes button text communicates the current state, sometimes it's function or the next state it'll activate.
// I probably just went with the version that i understood as i felt there is no convention, just personal preference.
