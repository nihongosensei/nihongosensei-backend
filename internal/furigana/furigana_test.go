package furigana

import (
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"testing"
)

func TestTextToFurigana(t *testing.T) {
	expected := Furigana{
		Text:      "読ん",
		Reading:   "よん",
		BasicForm: "読む",
	}

	tn, _ := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	result, _ := TextToFurigana(tn, "読んで")

	if result[0] != expected {
		t.Errorf("Expected %+v, but got %+v", expected, result[0])
	}
	if result[1] != "で" {
		t.Errorf("Expected %+v, but got %+v", "で", result[0])
	}
}
