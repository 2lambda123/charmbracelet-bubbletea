package tea

import (
	"time"
)

// Batch performs a bunch of commands concurrently with no ordering guarantees
// about the results. Use a Batch to return several commands.
//
// Example:
//
//	    func (m model) Init() Cmd {
//		       return tea.Batch(someCommand, someOtherCommand)
//	    }
func Batch(cmds ...Cmd) Cmd {
	var validCmds []Cmd //nolint:prealloc
	for _, c := range cmds {
		if c == nil {
			continue
		}
		validCmds = append(validCmds, c)
	}
	if len(validCmds) == 0 {
		return nil
	}
	return func() Msg {
		return BatchMsg(validCmds)
	}
}

// BatchMsg is a message used to perform a bunch of commands concurrently with
// no ordering guarantees. You can send a BatchMsg with Batch.
type BatchMsg []Cmd

// Sequence runs the given commands one at a time, in order. Contrast this with
// Batch, which runs commands concurrently.
func Sequence(cmds ...Cmd) Cmd {
	return func() Msg {
		return sequenceMsg(cmds)
	}
}

// sequenceMsg is used internally to run the given commands in order.
type sequenceMsg []Cmd

// Wrap allows a parent component to accurately wrap the results of any Cmd,
// regardless of whether it's a direct Cmd or a collective one like Batch or
// Sequence.  Note that the "wrapping" arg could also be an interface{}, or a
// generic type, or even a full `func(Cmd) Msg` (for the most flexibility), but
// using an int seems to satisfy the 80:20 rule of the simplest solution for the
// most-common case: enabling routing for sub-components.
func Wrap(cmd Cmd, id int) Cmd {
	if cmd == nil {
		return cmd
	}

	// Both Batch and Sequence are interesting in that they actually create a
	// command whose entire purpose is to "return a slice of commands as a new
	// message". Because of this, we have let the command execute, and *then*
	// detect that it's a Batch or Sequence, then unroll and re-roll with
	// individually-wrapped commands.
	return func() Msg {
		msg := cmd()

		var cmds []Cmd

		switch msg := msg.(type) {
		// Sadly, we can't lump BatchMsg and sequenceMsg together even
		// though they are both []Cmd... the compiler will only allow
		// casting for one of the known types at a time; in which case, even
		// an implicit cast is good enough!  (With both types listed in the
		// 'case', it assumes it's just a generic Msg.)
		case BatchMsg:
			cmds = msg
		case sequenceMsg:
			cmds = msg

		default:
			// for all other messages, we "simply" wrap the result
			return WrappedMsg{Id: id, Msg: msg}
		}

		// BatchMsg and sequenceMsg are specially-handled by the event loop;
		// we need to propagate the wrapping into the individual commands.
		wrapped := make([]Cmd, 0, len(cmds))
		for _, c := range cmds {
			if c == nil {
				continue
			}
			wrapped = append(wrapped, Wrap(c, id))
		}
		if _, ok := msg.(BatchMsg); ok {
			return BatchMsg(wrapped)
		}
		return sequenceMsg(wrapped)
	}
}

// WrappedMsg represents a wrapper around the Msg returned from a Cmd. See Wrap
// for further details.
type WrappedMsg struct {
	Id  int
	Msg Msg
}

// Every is a command that ticks in sync with the system clock. So, if you
// wanted to tick with the system clock every second, minute or hour you
// could use this. It's also handy for having different things tick in sync.
//
// Because we're ticking with the system clock the tick will likely not run for
// the entire specified duration. For example, if we're ticking for one minute
// and the clock is at 12:34:20 then the next tick will happen at 12:35:00, 40
// seconds later.
//
// To produce the command, pass a duration and a function which returns
// a message containing the time at which the tick occurred.
//
//	type TickMsg time.Time
//
//	cmd := Every(time.Second, func(t time.Time) Msg {
//	   return TickMsg(t)
//	})
//
// Beginners' note: Every sends a single message and won't automatically
// dispatch messages at an interval. To do that, you'll want to return another
// Every command after receiving your tick message. For example:
//
//	type TickMsg time.Time
//
//	// Send a message every second.
//	func tickEvery() Cmd {
//	    return Every(time.Second, func(t time.Time) Msg {
//	        return TickMsg(t)
//	    })
//	}
//
//	func (m model) Init() Cmd {
//	    // Start ticking.
//	    return tickEvery()
//	}
//
//	func (m model) Update(msg Msg) (Model, Cmd) {
//	    switch msg.(type) {
//	    case TickMsg:
//	        // Return your Every command again to loop.
//	        return m, tickEvery()
//	    }
//	    return m, nil
//	}
//
// Every is analogous to Tick in the Elm Architecture.
func Every(duration time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		n := time.Now()
		d := n.Truncate(duration).Add(duration).Sub(n)
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}

// Tick produces a command at an interval independent of the system clock at
// the given duration. That is, the timer begins precisely when invoked,
// and runs for its entire duration.
//
// To produce the command, pass a duration and a function which returns
// a message containing the time at which the tick occurred.
//
//	type TickMsg time.Time
//
//	cmd := Tick(time.Second, func(t time.Time) Msg {
//	   return TickMsg(t)
//	})
//
// Beginners' note: Tick sends a single message and won't automatically
// dispatch messages at an interval. To do that, you'll want to return another
// Tick command after receiving your tick message. For example:
//
//	type TickMsg time.Time
//
//	func doTick() Cmd {
//	    return Tick(time.Second, func(t time.Time) Msg {
//	        return TickMsg(t)
//	    })
//	}
//
//	func (m model) Init() Cmd {
//	    // Start ticking.
//	    return doTick()
//	}
//
//	func (m model) Update(msg Msg) (Model, Cmd) {
//	    switch msg.(type) {
//	    case TickMsg:
//	        // Return your Tick command again to loop.
//	        return m, doTick()
//	    }
//	    return m, nil
//	}
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}

// Sequentially produces a command that sequentially executes the given
// commands.
// The Msg returned is the first non-nil message returned by a Cmd.
//
//	func saveStateCmd() Msg {
//	   if err := save(); err != nil {
//	       return errMsg{err}
//	   }
//	   return nil
//	}
//
//	cmd := Sequentially(saveStateCmd, Quit)
//
// Deprecated: use Sequence instead.
func Sequentially(cmds ...Cmd) Cmd {
	return func() Msg {
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			if msg := cmd(); msg != nil {
				return msg
			}
		}
		return nil
	}
}

// setWindowTitleMsg is an internal message used to set the window title.
type setWindowTitleMsg string

// SetWindowTitle produces a command that sets the terminal title.
//
// For example:
//
//	func (m model) Init() Cmd {
//	    // Set title.
//	    return tea.SetWindowTitle("My App")
//	}
func SetWindowTitle(title string) Cmd {
	return func() Msg {
		return setWindowTitleMsg(title)
	}
}
