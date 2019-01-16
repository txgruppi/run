package text

func parse(data []byte) []*Token {
	length := len(data)
	tokens := []*Token{}
	inToken := false
	var token *Token
	var start int

	for index := 0; index < length-1; {
		switch {
		case isSpace(data[index]):
			index++

		case data[index] == '{' && data[index+1] == '{':
			if inToken {
				return nil
			}
			start = index
			index += 2
			inToken = true
			token = &Token{
				Keys: []string{},
			}

		case data[index] == '}' && data[index+1] == '}':
			if !inToken {
				return nil
			}
			index += 2
			inToken = false
			token.Raw = string(data[start:index])
			tokens = append(tokens, token)

		case data[index] == '|' && inToken:
			index += 1

		case inToken:
			var id []byte
			id, index = consumeIdentifier(data, index, length)
			token.Keys = append(token.Keys, string(id))

		case !inToken:
			index += 1
		}
	}

	return tokens
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

func isInRange(b, min, max byte) bool {
	return b >= min && b <= max
}

func isLower(b byte) bool {
	return isInRange(b, 'a', 'z')
}

func isUpper(b byte) bool {
	return isInRange(b, 'A', 'Z')
}

func isIdentifier(b byte) bool {
	return isLower(b) || isUpper(b) || b == '.' || b == '-' || b == '_'
}

func consumeIdentifier(data []byte, i, l int) ([]byte, int) {
	id := []byte{}
	for j := i; j < l; j++ {
		if isIdentifier(data[j]) {
			id = append(id, data[j])
			continue
		}
		return id, j
	}
	return nil, l
}
