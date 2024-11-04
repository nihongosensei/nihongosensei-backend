package japanese

import (
	"strings"
	"testing"
)

func TestKatakanaToHiragana(t *testing.T) {
	expected := "Konnichiwa, わたしは学生です。"
	result := KatakanaToHiragana("Konnichiwa, ワタシは学生です。")

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestFindKanjis(t *testing.T) {
	text := "すべての人間は、生まれながらにして自由であり"

	kanjis := FindKanjis(text)

	if len(kanjis) != 5 {
		t.Errorf("Expected 5 kanjis, but got %+v", len(kanjis))
	}

	if strings.Join(kanjis, "") != "人間生自由" {
		t.Errorf("Expected 人間生自由 as kanjis, but got %+v", strings.Join(kanjis, ""))
	}
}

func TestContainsKanjiWithOneKanji(t *testing.T) {
	if !ContainsKanji("学") {
		t.Error("Must return true for exact one kanji")
	}
}

func TestContainsKanjiWithMultipleKanji(t *testing.T) {
	if ContainsKanji("学学") {
		t.Error("Must return false for multiple kanjis")
	}
}
