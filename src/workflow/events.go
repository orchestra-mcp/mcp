package workflow

import "sync"

// TransitionEvent is emitted whenever an issue changes state.
type TransitionEvent struct {
	Project string `json:"project"`
	EpicID  string `json:"epic_id,omitempty"`
	StoryID string `json:"story_id,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
	Type    string `json:"type"` // epic, story, task, bug, hotfix
	From    string `json:"from"`
	To      string `json:"to"`
	Time    string `json:"time"`
}

// TransitionListener receives workflow transition events.
type TransitionListener interface {
	OnTransition(event TransitionEvent)
}

// TransitionListenerFunc adapts a function to TransitionListener.
type TransitionListenerFunc func(TransitionEvent)

func (f TransitionListenerFunc) OnTransition(e TransitionEvent) { f(e) }

var (
	mu        sync.RWMutex
	listeners []TransitionListener
)

// RegisterListener adds a listener for workflow transitions.
func RegisterListener(l TransitionListener) {
	mu.Lock()
	defer mu.Unlock()
	listeners = append(listeners, l)
}

// Emit broadcasts a transition event to all registered listeners.
func Emit(e TransitionEvent) {
	mu.RLock()
	defer mu.RUnlock()
	for _, l := range listeners {
		l.OnTransition(e)
	}
}
