package japanese

import "unicode"

// FindKanjis returns a list of kanjis from the given string
func FindKanjis(s string) []string {
	var kanjis []string
	for _, char := range s {
		if IsKanji(char) {
			kanjis = append(kanjis, string(char))
		}
	}

	return kanjis
}

// ContainsKanji checks if the given string contains at least one kanji.
func ContainsKanji(s string) bool {
	for _, char := range s {
		if IsKanji(char) {
			return true
		}
	}

	return false
}

// ContainsJapanese checks if the given string contains at least one hiragana,
// katakana or kanji.
func ContainsJapanese(s string) bool {
	for _, char := range s {
		if unicode.Is(unicode.Hiragana, char) || unicode.Is(unicode.Katakana, char) || IsKanji(char) {
			return true
		}
	}

	return false
}

// IsKanji checks if the given character is a kanji
func IsKanji(char rune) bool {
	return unicode.Is(unicode.Han, char)
}

// KatakanaToHiragana converts all katakana of the given string to hiragana.
func KatakanaToHiragana(s string) string {
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
