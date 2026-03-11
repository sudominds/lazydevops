package search

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

func NewTextInput(prompt string, placeholder string, charLimit int, width int) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt
	input.Placeholder = placeholder
	input.CharLimit = charLimit
	input.Width = width
	return input
}

func Matches(candidate string, query string) bool {
	normalizedQuery := NormalizeText(query)
	if normalizedQuery == "" {
		return true
	}

	normalizedCandidate := NormalizeText(candidate)
	if strings.Contains(normalizedCandidate, normalizedQuery) {
		return true
	}
	if strings.Contains(strings.ReplaceAll(normalizedCandidate, " ", ""), strings.ReplaceAll(normalizedQuery, " ", "")) {
		return true
	}

	queryTokens := strings.Fields(normalizedQuery)
	candidateTokens := strings.Fields(normalizedCandidate)
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return false
	}

	for _, queryToken := range queryTokens {
		matchedToken := false
		for _, candidateToken := range candidateTokens {
			if strings.Contains(candidateToken, queryToken) {
				matchedToken = true
				break
			}
		}
		if !matchedToken {
			return false
		}
	}

	return true
}

func NormalizeText(value string) string {
	loweredValue := strings.ToLower(strings.TrimSpace(value))
	if loweredValue == "" {
		return ""
	}

	normalizedBuilder := strings.Builder{}
	normalizedBuilder.Grow(len(loweredValue))
	previousWasSpace := false

	for _, character := range loweredValue {
		isSeparator := character == '_' || character == '-' || character == '.'
		isSpace := character == ' ' || character == '\t' || character == '\n' || character == '\r'
		if isSeparator || isSpace {
			if !previousWasSpace {
				normalizedBuilder.WriteRune(' ')
				previousWasSpace = true
			}
			continue
		}

		normalizedBuilder.WriteRune(character)
		previousWasSpace = false
	}

	return strings.TrimSpace(normalizedBuilder.String())
}

func MatchRuneIndexes(candidate string, query string) map[int]bool {
	normalizedQuery := NormalizeText(query)
	if normalizedQuery == "" {
		return nil
	}

	normalizedCandidate, candidateRuneMap := normalizeTextWithRuneMap(candidate)
	if normalizedCandidate == "" || len(candidateRuneMap) == 0 {
		return nil
	}

	matches := map[int]bool{}
	markNormalizedRange := func(start int, length int) {
		if length <= 0 {
			return
		}
		for index := start; index < start+length && index < len(candidateRuneMap); index++ {
			originalRuneIndex := candidateRuneMap[index]
			if originalRuneIndex >= 0 {
				matches[originalRuneIndex] = true
			}
		}
	}

	firstMatchIndex := strings.Index(normalizedCandidate, normalizedQuery)
	if firstMatchIndex >= 0 {
		markNormalizedRange(firstMatchIndex, len([]rune(normalizedQuery)))
		return matches
	}

	compactCandidate, compactRuneMap := compactNormalizedWithRuneMap(normalizedCandidate, candidateRuneMap)
	compactQuery := strings.ReplaceAll(normalizedQuery, " ", "")
	compactMatchIndex := strings.Index(compactCandidate, compactQuery)
	if compactQuery != "" && compactMatchIndex >= 0 {
		for index := compactMatchIndex; index < compactMatchIndex+len([]rune(compactQuery)) && index < len(compactRuneMap); index++ {
			originalRuneIndex := compactRuneMap[index]
			if originalRuneIndex >= 0 {
				matches[originalRuneIndex] = true
			}
		}
		if len(matches) > 0 {
			return matches
		}
	}

	queryTokens := strings.Fields(normalizedQuery)
	candidateTokens := splitNormalizedTokensWithOffsets(normalizedCandidate)
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return nil
	}

	for _, queryToken := range queryTokens {
		if queryToken == "" {
			continue
		}
		tokenMatched := false
		for _, token := range candidateTokens {
			matchIndex := runeSubsequenceIndex([]rune(token.Value), []rune(queryToken))
			if matchIndex < 0 {
				continue
			}
			start := token.Start + matchIndex
			markNormalizedRange(start, len([]rune(queryToken)))
			tokenMatched = true
			break
		}
		if !tokenMatched {
			return nil
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return matches
}

type normalizedToken struct {
	Value string
	Start int
}

func splitNormalizedTokensWithOffsets(value string) []normalizedToken {
	runes := []rune(value)
	tokens := make([]normalizedToken, 0, 8)
	start := -1
	for index, character := range runes {
		if character == ' ' {
			if start >= 0 {
				tokens = append(tokens, normalizedToken{Value: string(runes[start:index]), Start: start})
				start = -1
			}
			continue
		}
		if start < 0 {
			start = index
		}
	}
	if start >= 0 {
		tokens = append(tokens, normalizedToken{Value: string(runes[start:]), Start: start})
	}
	return tokens
}

func runeSubsequenceIndex(value []rune, target []rune) int {
	if len(target) == 0 || len(value) < len(target) {
		return -1
	}
	maxStart := len(value) - len(target)
	for start := 0; start <= maxStart; start++ {
		matched := true
		for offset := 0; offset < len(target); offset++ {
			if value[start+offset] != target[offset] {
				matched = false
				break
			}
		}
		if matched {
			return start
		}
	}
	return -1
}

func normalizeTextWithRuneMap(value string) (string, []int) {
	loweredValue := strings.ToLower(strings.TrimSpace(value))
	if loweredValue == "" {
		return "", nil
	}

	originalRunes := []rune(loweredValue)
	normalizedRunes := make([]rune, 0, len(originalRunes))
	runeMap := make([]int, 0, len(originalRunes))
	previousWasSpace := false
	for originalIndex, character := range originalRunes {
		isSeparator := character == '_' || character == '-' || character == '.'
		isSpace := character == ' ' || character == '\t' || character == '\n' || character == '\r'
		if isSeparator || isSpace {
			if !previousWasSpace {
				normalizedRunes = append(normalizedRunes, ' ')
				runeMap = append(runeMap, -1)
				previousWasSpace = true
			}
			continue
		}
		normalizedRunes = append(normalizedRunes, character)
		runeMap = append(runeMap, originalIndex)
		previousWasSpace = false
	}

	for len(normalizedRunes) > 0 && normalizedRunes[0] == ' ' {
		normalizedRunes = normalizedRunes[1:]
		runeMap = runeMap[1:]
	}
	for len(normalizedRunes) > 0 && normalizedRunes[len(normalizedRunes)-1] == ' ' {
		normalizedRunes = normalizedRunes[:len(normalizedRunes)-1]
		runeMap = runeMap[:len(runeMap)-1]
	}
	return string(normalizedRunes), runeMap
}

func compactNormalizedWithRuneMap(normalizedValue string, runeMap []int) (string, []int) {
	if normalizedValue == "" || len(runeMap) == 0 {
		return "", nil
	}
	normalizedRunes := []rune(normalizedValue)
	compactRunes := make([]rune, 0, len(normalizedRunes))
	compactMap := make([]int, 0, len(normalizedRunes))
	for index, character := range normalizedRunes {
		if character == ' ' {
			continue
		}
		compactRunes = append(compactRunes, character)
		if index < len(runeMap) {
			compactMap = append(compactMap, runeMap[index])
		} else {
			compactMap = append(compactMap, -1)
		}
	}
	return string(compactRunes), compactMap
}
