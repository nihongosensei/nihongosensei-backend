package furigana

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"unicode"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

type TokenizeHandler struct {
	Tokenizer *tokenizer.Tokenizer
}

type PostBody struct {
	Text string `json:"text"`
}

type Response struct {
	Text      string `json:"text"`                // the actual character(s), e.g. "読ん"
	Reading   string `json:"reading,omitempty"`   // the reading (in hiragana), e.g. "よん"
	BasicForm string `json:"basicForm,omitempty"` // basic form, e.g. "読む"
}

func (h *TokenizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input PostBody
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if input.Text == "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
		return
	}

	furigana, err := parseJapaneseTextToFurigana(h.Tokenizer, input.Text)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating furigana: %v\n", err)
		return
	}

	resp, err := json.Marshal(furigana)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func parseJapaneseTextToFurigana(tn *tokenizer.Tokenizer, text string) ([]interface{}, error) {
	tokens := tn.Analyze(text, tokenizer.Normal)

	var furigana []interface{}
	var carry bytes.Buffer

	for _, t := range tokens {
		if t.ID == tokenizer.BosEosID {
			continue
		}
		if containsKanji(t.Surface) {
			if carry.Len() > 0 {
				furigana = append(furigana, carry.String())
				carry.Reset()
			}
			furigana = append(furigana, transformTokenToFurigana(t))
		} else {
			carry.WriteString(t.Surface)
		}
	}

	if carry.Len() > 0 {
		furigana = append(furigana, carry.String())
	}

	return furigana, nil
}

func transformTokenToFurigana(t tokenizer.Token) Response {
	var reading string
	maybeReading, readingExists := t.Reading()
	if readingExists {
		reading = katakanaToHiragana(maybeReading)
	}

	baseForm, _ := t.BaseForm()

	return Response{
		Text:      t.Surface,
		Reading:   reading,
		BasicForm: baseForm,
	}
}

func containsKanji(s string) bool {
	for _, char := range s {
		if unicode.Is(unicode.Han, char) {
			return true
		}
	}
	return false
}

func katakanaToHiragana(s string) string {
	var result []rune

	for _, char := range s {
		// Check if the character is a Katakana character
		if char >= '\u30A1' && char <= '\u30F6' {
			// Convert to Hiragana by adjusting the Unicode value
			result = append(result, char-0x60)
		} else {
			// Keep the character as is if it's not Katakana
			result = append(result, char)
		}
	}

	return string(result)
}
