package main

import (
	"strings"
	"testing"
)

func TestNewGapBuffer(t *testing.T) {
	t.Run("valid cursor", func(t *testing.T) {
		gb, err := NewGapBuffer("hello", 3)
		if err != nil {
			t.Fatal(err)
		}
		if gb.String() != "hello" {
			t.Fatalf("got %q, want %q", gb.String(), "hello")
		}
		if gb.CursorPos() != 3 {
			t.Fatalf("cursor %d, want 3", gb.CursorPos())
		}
	})

	t.Run("cursor at start", func(t *testing.T) {
		gb, err := NewGapBuffer("hello", 0)
		if err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 0 {
			t.Fatalf("cursor %d, want 0", gb.CursorPos())
		}
	})

	t.Run("cursor at end", func(t *testing.T) {
		gb, err := NewGapBuffer("hello", 5)
		if err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 5 {
			t.Fatalf("cursor %d, want 5", gb.CursorPos())
		}
	})

	t.Run("out of range negative", func(t *testing.T) {
		_, err := NewGapBuffer("hello", -1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("out of range positive", func(t *testing.T) {
		_, err := NewGapBuffer("hello", 6)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		gb, err := NewGapBuffer("", 0)
		if err != nil {
			t.Fatal(err)
		}
		if gb.String() != "" {
			t.Fatalf("got %q, want empty", gb.String())
		}
		if gb.Len() != 0 {
			t.Fatalf("len %d, want 0", gb.Len())
		}
	})

	t.Run("unicode", func(t *testing.T) {
		gb, err := NewGapBuffer("héllo 世界", 3)
		if err != nil {
			t.Fatal(err)
		}
		if gb.String() != "héllo 世界" {
			t.Fatalf("got %q", gb.String())
		}
		if gb.Len() != 8 {
			t.Fatalf("len %d, want 8", gb.Len())
		}
	})
}

func TestInsert(t *testing.T) {
	t.Run("single rune at end", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 5)
		gb.Insert('!')
		if gb.String() != "hello!" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("at start", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 0)
		gb.Insert('>')
		if gb.String() != ">hello" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("at middle", func(t *testing.T) {
		gb, _ := NewGapBuffer("helo", 3)
		gb.Insert('l')
		if gb.String() != "hello" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("triggers grow", func(t *testing.T) {
		gb, _ := NewGapBuffer("", 0)
		long := strings.Repeat("a", 100)
		for _, r := range long {
			gb.Insert(r)
		}
		if gb.String() != long {
			t.Fatalf("got len %d, want 100", len(gb.String()))
		}
	})
}

func TestInsertString(t *testing.T) {
	t.Run("multi char", func(t *testing.T) {
		gb, _ := NewGapBuffer("hd", 1)
		gb.InsertString("ello worl")
		if gb.String() != "hello world" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("unicode", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello ", 6)
		gb.InsertString("世界")
		if gb.String() != "hello 世界" {
			t.Fatalf("got %q", gb.String())
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("at cursor", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 5)
		gb.Delete()
		if gb.String() != "hell" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("at start no-op", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 0)
		gb.Delete()
		if gb.String() != "hello" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("multiple deletes", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 5)
		gb.Delete()
		gb.Delete()
		gb.Delete()
		if gb.String() != "he" {
			t.Fatalf("got %q", gb.String())
		}
	})
}

func TestDeleteForward(t *testing.T) {
	t.Run("at end no-op", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 5)
		gb.DeleteForward()
		if gb.String() != "hello" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("middle", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 2)
		gb.DeleteForward()
		if gb.String() != "helo" {
			t.Fatalf("got %q", gb.String())
		}
	})
}

func TestMoveCursor(t *testing.T) {
	t.Run("left", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 3)
		if err := gb.MoveCursor(1); err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 1 {
			t.Fatalf("cursor %d, want 1", gb.CursorPos())
		}
		if gb.String() != "hello" {
			t.Fatalf("got %q", gb.String())
		}
	})

	t.Run("right", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 1)
		if err := gb.MoveCursor(4); err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 4 {
			t.Fatalf("cursor %d, want 4", gb.CursorPos())
		}
	})

	t.Run("out of range", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 3)
		if err := gb.MoveCursor(10); err == nil {
			t.Fatal("expected error")
		}
		if err := gb.MoveCursor(-1); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("to 0", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 3)
		if err := gb.MoveCursor(0); err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 0 {
			t.Fatalf("cursor %d", gb.CursorPos())
		}
	})

	t.Run("to end", func(t *testing.T) {
		gb, _ := NewGapBuffer("hello", 0)
		if err := gb.MoveCursor(5); err != nil {
			t.Fatal(err)
		}
		if gb.CursorPos() != 5 {
			t.Fatalf("cursor %d", gb.CursorPos())
		}
	})
}

func TestMoveCursorOriginal(t *testing.T) {
	gb, _ := NewGapBuffer("hello", 2)
	gb.Insert('X')
	// offset is now 1, cursor is at 3
	// MoveCursorOriginal(4) → MoveCursor(4+1=5)
	if err := gb.MoveCursorOriginal(4); err != nil {
		t.Fatal(err)
	}
	if gb.CursorPos() != 5 {
		t.Fatalf("cursor %d, want 5", gb.CursorPos())
	}
}

func TestOriginalCursorPos(t *testing.T) {
	gb, _ := NewGapBuffer("hello", 2)
	if gb.OriginalCursorPos() != 2 {
		t.Fatalf("got %d, want 2", gb.OriginalCursorPos())
	}
	gb.Insert('X')
	gb.Insert('Y')
	// cursor now at 4, offset 2 → original 2
	if gb.OriginalCursorPos() != 2 {
		t.Fatalf("got %d, want 2", gb.OriginalCursorPos())
	}
	gb.Delete()
	// cursor now at 3, offset 1 → original 2
	if gb.OriginalCursorPos() != 2 {
		t.Fatalf("got %d, want 2", gb.OriginalCursorPos())
	}
}

func TestString(t *testing.T) {
	gb, _ := NewGapBuffer("hello", 2)
	gb.Insert('X')
	gb.MoveCursor(0)
	gb.Insert('>')
	if gb.String() != ">heXllo" {
		t.Fatalf("got %q", gb.String())
	}
}

func TestLen(t *testing.T) {
	gb, _ := NewGapBuffer("hello", 5)
	if gb.Len() != 5 {
		t.Fatalf("got %d", gb.Len())
	}
	gb.Insert('!')
	if gb.Len() != 6 {
		t.Fatalf("got %d", gb.Len())
	}
	gb.Delete()
	gb.Delete()
	if gb.Len() != 4 {
		t.Fatalf("got %d", gb.Len())
	}
}

func TestEditingWorkflow(t *testing.T) {
	// Simulate: start with empty, type "hello", backspace twice, type "p me"
	gb, _ := NewGapBuffer("", 0)
	gb.InsertString("hello")
	if gb.String() != "hello" {
		t.Fatalf("got %q", gb.String())
	}
	gb.Delete()
	gb.Delete()
	if gb.String() != "hel" {
		t.Fatalf("got %q", gb.String())
	}
	gb.InsertString("p me")
	if gb.String() != "help me" {
		t.Fatalf("got %q", gb.String())
	}

	// Move to start, insert "Please "
	gb.MoveCursor(0)
	gb.InsertString("Please ")
	if gb.String() != "Please help me" {
		t.Fatalf("got %q", gb.String())
	}

	// Move to end, add "!"
	gb.MoveCursor(gb.Len())
	gb.Insert('!')
	if gb.String() != "Please help me!" {
		t.Fatalf("got %q", gb.String())
	}
}
