package furigana

import (
	"strings"

	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/nihongosensei/nihongosensei-backend/internal/japanese"
)

type Furigana struct {
	Text      string `json:"text"`                // the actual character(s), e.g. "読ん"
	Reading   string `json:"reading,omitempty"`   // the reading (in hiragana), e.g. "よん"
	BasicForm string `json:"basicForm,omitempty"` // basic form, e.g. "読む"
}

func TextToFurigana(tn *tokenizer.Tokenizer, text string) ([]any, error) {
	tokens := tn.Analyze(text, tokenizer.Normal)

	var furigana []any
	var carry strings.Builder

	for _, t := range tokens {
		if t.ID == tokenizer.BosEosID {
			continue
		}
		if japanese.ContainsKanji(t.Surface) {
			if carry.Len() > 0 {
				furigana = append(furigana, carry.String())
				carry.Reset()
			}
			furigana = append(furigana, tokenToFurigana(t))
		} else {
			carry.WriteString(t.Surface)
		}
	}

	if carry.Len() > 0 {
		furigana = append(furigana, carry.String())
	}

	return furigana, nil
}

func tokenToFurigana(t tokenizer.Token) Furigana {
	var reading string
	maybeReading, readingExists := t.Reading()
	if readingExists {
		reading = japanese.KatakanaToHiragana(maybeReading)
	}

	baseForm, _ := t.BaseForm()

	return Furigana{
		Text:      t.Surface,
		Reading:   reading,
		BasicForm: baseForm,
	}
}
