# ADR 1: Refactor Project Structure

## Status: Accepted

## Context

[Design the architecture, name the components, document the details.](https://go-proverbs.github.io/).

We want to start small and avoid deeply nested directories to keep import paths shorter. 'Src' is not a fitting directory as a Go project is just plain source code.

## Decision

- main.go should be nothing but the entry point of the app
- files should somewhat follow the Single Responsibility principle
  → slideshow.go should at least be separated into config (data) and player (business logic)
  → main window and slideshow window should be separated into files

First iteration structure:

- main.go → entry point, only assemble components
- main_window.go → fyne.App + startup window
- slideshow_window.go -> fyne slideshow window
- image_repository.go → load images + preprocess (validate, sort, shuffle)
- slideshow_player.go → timer based slideshow within goroutine

Following the dependency inversion principle our image DTO should be implemented as an interface abstraction. The slideshow_player communicates it's need with the interface.
DI is not required here and in principle you shouldn't write speculative code. The scope of the project does not expect a different provider/driver for our player than the direct call from the app.
I simply implement it this way as practise within this learning project.

## Consequences

The more atomic code separation into files grouping function should help navigate the codebase and make anti-patterns more easily visible.

Since the project is relatively small in scope the refactor will most likely take more time than it's benefits will save when fixing the more critical bugs.
