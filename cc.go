package main

import (
	"bufio"
	"fmt"
	"strings"
)

type Locale struct {
	Language string
	Tokens   []Token
}

func (l *Locale) SymbolCount() int {
	var count int
	for _, t := range l.Tokens {
		count += len(t.Value)
	}
	return count
}

type Token struct {
	Key   string
	Value string
}

func ParseLocaleFile(input string) (*Locale, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	locale := &Locale{}
	inTokensBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || line == "{" || line == "}" {
			continue
		}

		if strings.HasPrefix(line, "\"language\"") {
			parts := splitQuotedLine(line)
			if len(parts) == 2 {
				locale.Language = parts[1]
			}
			continue
		}

		if line == "\"tokens\"" {
			inTokensBlock = true
			continue
		}

		if inTokensBlock {
			if line == "}" {
				inTokensBlock = false
				continue
			}

			parts := splitQuotedLine(line)
			if len(parts) == 2 {
				locale.Tokens = append(locale.Tokens, Token{
					Key:   parts[0],
					Value: parts[1],
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return locale, nil
}

func splitQuotedLine(line string) []string {
	var parts []string
	var buf strings.Builder
	inQuote := false
	escaped := false

	for _, r := range line {
		switch {
		case escaped:
			buf.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			if inQuote {
				parts = append(parts, buf.String())
				buf.Reset()
			}
			inQuote = !inQuote
		default:
			if inQuote {
				buf.WriteRune(r)
			}
		}
	}

	return parts
}

func WriteLocaleFile(locale *Locale) string {
	var buf strings.Builder
	buf.WriteString("\"lang\"\n{\n")
	buf.WriteString(fmt.Sprintf("\t\"language\" \"%s\"\n", locale.Language))
	buf.WriteString("\t\"tokens\"\n\t{\n")

	for _, token := range locale.Tokens {
		escapedValue := strings.ReplaceAll(token.Value, `"`, `\"`)
		buf.WriteString(fmt.Sprintf("\t\t\"%s\"\t\t\"%s\"\n", token.Key, escapedValue))
	}

	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String()
}
