package furigana

import (
	"testing"
)

func TestTransformTokenToFurigana(t *testing.T) {
	expected := Response{
		Text:      "読ん",
		Reading:   "よん",
		BasicForm: "読む",
	}

	result, _ := parseJapaneseTextToFurigana("読んで")

	if result[0] != expected {
		t.Errorf("Expected %+v, but got %+v", expected, result[0])
	}
	if result[1] != "で" {
		t.Errorf("Expected %+v, but got %+v", "で", result[0])
	}
}

func TestKatakanaToHiragana(t *testing.T) {
	expected := "よんで"
	result := katakanaToHiragana("ヨンデ")

	if result != expected {
		t.Errorf("Expected %+v, but got %+v", expected, result[0])
	}
}
