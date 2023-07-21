package tea

// Renderer is the interface for Bubble Tea renderers.
type Renderer interface {
	// Start the renderer.
	start()

	// Stop the renderer, but render the final frame in the buffer, if any.
	stop()

	// Stop the renderer without doing any final rendering.
	kill()

	// Write a frame to the renderer. The renderer can write this data to
	// output at its discretion.
	Write(string)

	// Request a full re-render. Note that this will not trigger a render
	// immediately. Rather, this method causes the next render to be a full
	// repaint. Because of this, it's safe to call this method multiple times
	// in succession.
	repaint()

	// Clears the terminal.
	clearScreen()

	// Whether or not the alternate screen buffer is enabled.
	altScreen() bool
	// Enable the alternate screen buffer.
	enterAltScreen()
	// Disable the alternate screen buffer.
	exitAltScreen()

	// Show the cursor.
	showCursor()
	// Hide the cursor.
	hideCursor()

	// enableMouseCellMotion enables mouse click, release, wheel and motion
	// events if a mouse button is pressed (i.e., drag events).
	enableMouseCellMotion()

	// DisableMouseCellMotion disables Mouse Cell Motion tracking.
	disableMouseCellMotion()

	// EnableMouseAllMotion enables mouse click, release, wheel and motion
	// events, regardless of whether a mouse button is pressed. Many modern
	// terminals support this, but not all.
	enableMouseAllMotion()

	// DisableMouseAllMotion disables All Motion mouse tracking.
	disableMouseAllMotion()

	handleMessages(Msg)
}

// repaintMsg forces a full repaint.
type repaintMsg struct{}
