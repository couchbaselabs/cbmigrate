package couchbase

import (
	"fmt"
	"strings"
)

// GenerateCouchbaseArrayIndex generates a Couchbase array index expression
// from a given input pattern like "schedule[].special_flights[].flight".
func GenerateCouchbaseArrayIndex(input string) string {
	parts := strings.Split(input, "[]")
	// Removing the last empty element if input ends with "[]"
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimPrefix(parts[i], ".")
	}
	return buildExpression(parts, "")
}

// buildExpression constructs the Couchbase array index expression recursively.
func buildExpression(parts []string, parent string) string {
	if len(parts) == 1 {
		return fmt.Sprintf("%s.%s", parent, formatFieldReference(parts[0]))
	}
	item := "`" + parts[0] + "Item" + "`"
	items := formatFieldReference(parts[0])
	if parent != "" {
		items = parent + "." + items
	}
	inner := buildExpression(parts[1:], item)
	field := inner

	if parent == "" {
		return fmt.Sprintf("DISTINCT ARRAY %s FOR %s IN %s END", field, item, items)
	}
	return fmt.Sprintf("(DISTINCT ARRAY %s FOR %s IN %s END)", field, item, items)
}

func formatFieldReference(field string) string {
	return "`" + strings.ReplaceAll(field, ".", "`.`") + "`"
}
