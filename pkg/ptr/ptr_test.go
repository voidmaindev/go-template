package ptr

import (
	"testing"
)

func TestTo(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		s := "hello"
		p := To(s)
		if *p != s {
			t.Errorf("To() = %s, want %s", *p, s)
		}
	})

	t.Run("int", func(t *testing.T) {
		i := 42
		p := To(i)
		if *p != i {
			t.Errorf("To() = %d, want %d", *p, i)
		}
	})

	t.Run("int64", func(t *testing.T) {
		i := int64(100)
		p := To(i)
		if *p != i {
			t.Errorf("To() = %d, want %d", *p, i)
		}
	})

	t.Run("bool", func(t *testing.T) {
		b := true
		p := To(b)
		if *p != b {
			t.Errorf("To() = %v, want %v", *p, b)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type MyStruct struct {
			Name string
			Age  int
		}
		s := MyStruct{Name: "test", Age: 30}
		p := To(s)
		if *p != s {
			t.Errorf("To() = %v, want %v", *p, s)
		}
	})
}

func TestDeref(t *testing.T) {
	t.Run("string nil", func(t *testing.T) {
		var s *string
		got := Deref(s)
		if got != "" {
			t.Errorf("Deref() = %s, want empty string", got)
		}
	})

	t.Run("string value", func(t *testing.T) {
		s := "hello"
		got := Deref(&s)
		if got != s {
			t.Errorf("Deref() = %s, want %s", got, s)
		}
	})

	t.Run("int nil", func(t *testing.T) {
		var i *int
		got := Deref(i)
		if got != 0 {
			t.Errorf("Deref() = %d, want 0", got)
		}
	})

	t.Run("int value", func(t *testing.T) {
		i := 42
		got := Deref(&i)
		if got != i {
			t.Errorf("Deref() = %d, want %d", got, i)
		}
	})

	t.Run("bool nil", func(t *testing.T) {
		var b *bool
		got := Deref(b)
		if got != false {
			t.Errorf("Deref() = %v, want false", got)
		}
	})
}

func TestDerefOr(t *testing.T) {
	t.Run("string nil uses default", func(t *testing.T) {
		var s *string
		got := DerefOr(s, "default")
		if got != "default" {
			t.Errorf("DerefOr() = %s, want default", got)
		}
	})

	t.Run("string value ignores default", func(t *testing.T) {
		s := "hello"
		got := DerefOr(&s, "default")
		if got != s {
			t.Errorf("DerefOr() = %s, want %s", got, s)
		}
	})

	t.Run("int nil uses default", func(t *testing.T) {
		var i *int
		got := DerefOr(i, 10)
		if got != 10 {
			t.Errorf("DerefOr() = %d, want 10", got)
		}
	})

	t.Run("int value ignores default", func(t *testing.T) {
		i := 42
		got := DerefOr(&i, 10)
		if got != i {
			t.Errorf("DerefOr() = %d, want %d", got, i)
		}
	})

	t.Run("int zero value ignores default", func(t *testing.T) {
		i := 0
		got := DerefOr(&i, 10)
		if got != 0 {
			t.Errorf("DerefOr() = %d, want 0", got)
		}
	})

	t.Run("negative value ignores default", func(t *testing.T) {
		i := -1
		got := DerefOr(&i, 10)
		if got != -1 {
			t.Errorf("DerefOr() = %d, want -1", got)
		}
	})
}
