package discussion

import "testing"

func TestThreadState(t *testing.T) {
	open := Thread{
		Status: ThreadStatusOpen,
	}

	if !open.CanComment() {
		t.Error("open thread should accept comments")
	}

	if !open.CanResolve() {
		t.Error("open thread should be resolvable")
	}

	if open.CanReopen() {
		t.Error("open thread should not be reopened")
	}

	resolved := Thread{
		Status: ThreadStatusResolved,
	}

	if resolved.CanComment() {
		t.Error("resolved thread should not accept comments")
	}

	if resolved.CanResolve() {
		t.Error("resolved thread should not be resolved again")
	}

	if !resolved.CanReopen() {
		t.Error("resolved thread should be reopenable")
	}
}

func TestCommentValidation(t *testing.T) {
	comment := Comment{
		ID:       "COM00001",
		ThreadID: "THR00001",
		AuthorID: "USR00001",
		Body:     "Please clarify the reason for this change.",
	}

	if err := comment.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
