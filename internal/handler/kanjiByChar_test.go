package handler

import (
	"testing"
)

func TestContainsExactlyOneKanji(t *testing.T) {
	t.Run("WithAdditionalCharacters", func(t *testing.T) {
		if ContainsExactlyOneKanji("読んで") == true {
			t.Errorf("Expected false for \"読んで\"")
		}
	})

	t.Run("WithWhitespace", func(t *testing.T) {
		if ContainsExactlyOneKanji("読 ") == true {
			t.Errorf("Expected false for \"読 \"")
		}
	})

	t.Run("Success", func(t *testing.T) {
		if ContainsExactlyOneKanji("読") == false {
			t.Errorf("Expected true for \"読\"")
		}
	})

	t.Run("WithMultipleKanjis", func(t *testing.T) {
		if ContainsExactlyOneKanji("読読") == true {
			t.Errorf("Expected false for \"読読\"")
		}
	})
}
