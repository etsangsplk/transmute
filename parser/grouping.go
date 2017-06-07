// The grouping code handles "query groups". These are expressions in search strategies that look like:
//
// or/3-5
// and/6,7
// 9 and 10
// 4 and (5 or 6)
// etc.
//
// Generally, the type of grouping (postfix or infix) can be inferred ahead of time. This means that there needs to be
// two types of parsing methods. Additionally, only "and", "not" and "or" operators need to be taken into account.
// The prefix groups are parsed and transformed in-place, however the infix groups are parsed into a prefix tree and
// then transformed.
package parser

import (
	"strconv"
	"strings"
	"log"
	"unicode"
)

// QueryGroup is an intermediate structure, not part of the transmute intermediate representation, but instead used to
// construct the intermediate representation. It forms a common representation of query groupings.
type QueryGroup struct {
	// Line number of the query group. Ids of Children are 0.
	Id             int
	// Operator (one of: and, or, not)
	Type           string
	// The line numbers a keyword falls on (numbering start at 1)
	KeywordNumbers []int
	// Nested groups. for example (4 and (5 or 6)). The inner expression (5 or 6) is a child.
	Children       []QueryGroup
}

// transformPrefixGroupToQueryGroup transforms a prefix syntax tree into a query group. The new QueryGroup is built by
// recursively navigating the syntax tree.
func transformPrefixGroupToQueryGroup(prefix []string, queryGroup QueryGroup) ([]string, QueryGroup) {
	if len(prefix) == 0 {
		return prefix, queryGroup
	}

	token := prefix[0]
	if token == "and" || token == "or" || token == "not" {
		queryGroup.Type = token
	} else if token == "(" {
		var subGroup QueryGroup
		prefix, subGroup = transformPrefixGroupToQueryGroup(prefix[1:], QueryGroup{})
		queryGroup.Children = append(queryGroup.Children, subGroup)
	} else if token == ")" {
		return prefix, queryGroup
	} else {
		keywordNum, err := strconv.Atoi(token)
		if err != nil {
			log.Panicln(err)
		}
		queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, keywordNum)
	}
	return transformPrefixGroupToQueryGroup(prefix[1:], queryGroup)
}

// convertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func convertInfixToPrefix(infix []string) []string {
	// The stack contains some intermediate values
	stack := []string{}
	// The result contains the actual expression
	result := []string{}

	precedence := map[string]int{
		"and": 0,
		"or": 1,
		"not": 0,
	}

	// The algorithm is slightly modified to also store the brackets in the result
	for i := len(infix) - 1; i >= 0; i-- {
		token := infix[i]
		if token == ")" {
			stack = append(stack, token)
			result = append(result, token)
		} else if token == "(" {
			for len(stack) > 0 {
				var t string
				t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
				if t == ")" {
					result = append(result, "(")
					break
				}
				result = append(result, t)
			}
		} else if _, ok := precedence[token]; !ok {
			result = append(result, token)
		} else {
			for len(stack) > 0 && precedence[stack[len(stack) - 1]] > precedence[token] {
				var t string
				t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
				result = append(result, t)
			}
			stack = append(stack, token)
		}

	}

	for len(stack) > 0 {
		var t string
		t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
		result = append(result, t)
	}

	// The algorithm actually produces a postfix expression so it must be reversed
	// We can do this in-place with go!
	for i := len(result) / 2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}

	return result
}

// parseInfixGrouping translates an infix grouping into a prefix grouping, and then transforms it into a QueryGroup in
// two separate steps.
func parseInfixGrouping(group string, startsAfter rune) QueryGroup {
	group += "\n"
	group = strings.ToLower(group)

	stack := []string{}
	inGroup := startsAfter == rune(0)
	keyword := ""

	for _, char := range group {
		// Ignore the first few characters of the line
		if !inGroup && char == startsAfter {
			inGroup = true
			continue
		} else if !inGroup {
			continue
		}

		if unicode.IsSpace(char) && len(keyword) > 0 {
			stack = append(stack, strings.TrimSpace(keyword))
			keyword = ""
			continue
		} else if char == '(' {
			stack = append(stack, "(")
			keyword = ""
			continue
		} else if char == ')' {
			if len(keyword) > 0 {
				stack = append(stack, strings.TrimSpace(keyword))
				keyword = ""
			}
			stack = append(stack, ")")
			continue
		} else if !unicode.IsSpace(char) {
			keyword += string(char)
		}
	}

	prefix := convertInfixToPrefix(stack)
	_, queryGroup := transformPrefixGroupToQueryGroup(prefix, QueryGroup{})
	return queryGroup
}

// parseGrouping parses and constructs a QueryGroup in-place. Since the grouping is post-fix, no additional
// transformation is necessary.
func parsePrefixGrouping(group string, startsAfter rune) QueryGroup {
	group += "\n"

	var nums []string
	var num string

	var sep string

	var operator string

	queryGroup := QueryGroup{}

	inGroup := false
	if startsAfter == rune(0) {
		inGroup = true
	}

	for _, char := range group {
		// Set the separator
		if char == '-' {
			sep = "-"
		} else if char == ',' {
			sep = ","
		}

		// Ignore the first few characters of the line
		if !inGroup && char == startsAfter {
			inGroup = true
			continue
		} else if !inGroup {
			continue
		}

		// Extract the numbers
		if unicode.IsNumber(char) {
			num += string(char)
		} else if len(num) > 0 {
			nums = append(nums, num)
			num = ""

			// Now, unfortunately there is an infix operator INSIDE the postfix expression.
			// This is parsed as follows:
			if len(nums) == 2 {
				if sep == "-" {
					queryGroup.Type = operator

					lhs, err := strconv.Atoi(nums[0])
					if err != nil {
						panic(err)
					}

					rhs, err := strconv.Atoi(nums[1])
					if err != nil {
						panic(err)
					}

					for i := lhs; i <= rhs; i++ {
						queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, i)
					}

				}
				nums = []string{}

			} else if len(nums) == 1 {
				if sep == "," {
					lhs, err := strconv.Atoi(nums[0])
					if err != nil {
						panic(err)
					}
					queryGroup.Type = operator
					queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, lhs)
					nums = []string{}
				}
			}

		}

		// Extract the groups
		if operator != "or" && operator != "and" && operator != "not" {
			operator += strings.ToLower(string(char))
		}

	}
	return queryGroup
}
