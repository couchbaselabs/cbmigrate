package repo

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"strings"
)

// GenerateCouchbaseArrayIndex generates a Couchbase array index expression
// from a given input pattern like "schedule[].special_flights[].flight".
func GenerateCouchbaseArrayIndex(input string) string {
	parts := strings.Split(input, "[]")
	// Removing the last empty element if input ends with "[]"
	//if parts[len(parts)-1] == "" {
	//	parts = parts[:len(parts)-1]
	//}
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimPrefix(parts[i], ".")
	}
	return buildExpression(parts, "", 0)
}

// buildExpression constructs the Couchbase array index expression recursively.
func buildExpression(parts []string, parent string, l int) string {
	l++
	if len(parts) == 1 {
		if parts[0] == "" {
			return fmt.Sprintf("%s", parent)
		}
		tparts := strings.Split(parts[0], ",")
		if len(tparts) > 1 {
			var exp strings.Builder
			exp.WriteString("FLATTEN_KEYS(")
			for i, part := range tparts {
				if i != 0 {
					exp.WriteString(",")
				}
				exp.WriteString(fmt.Sprintf("%s.%s", parent,
					formatFieldReferenceWithAddtionalOptions(strings.TrimPrefix(part, "."))))
			}
			exp.WriteString(")")
			return exp.String()
		}
		return fmt.Sprintf("%s.%s", parent, formatFieldReferenceWithAddtionalOptions(parts[0]))
	}
	item := fmt.Sprintf("`l%dItem`", l)
	items := formatFieldReference(parts[0])
	if parent != "" {
		items = parent + "." + items
	}
	inner := buildExpression(parts[1:], item, l)
	field := inner

	if parent == "" {
		return fmt.Sprintf("DISTINCT ARRAY %s FOR %s IN %s END", field, item, items)
	}
	return fmt.Sprintf("(DISTINCT ARRAY %s FOR %s IN %s END)", field, item, items)
}

func formatFieldReference(field string) string {
	return "`" + strings.ReplaceAll(field, ".", "`.`") + "`"
}

func formatFieldReferenceWithAddtionalOptions(field string) string {
	// for the give array group "k2[].n1k1[].n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC INCLUDE MISSING,.n2k2 ASC INCLUDE MISSING"
	// the last element after splitting has .n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC INCLUDE MISSING,.n2k2 ASC INCLUDE MISSING
	// this handle the quotes to not include DESC INCLUDE MISSING content
	splitString := strings.Split(field, " ")
	for i, str := range splitString {
		if str == "" {
			continue
		}
		splitString[i] = "`" + strings.ReplaceAll(str, ".", "`.`") + "`"
		break
	}
	return strings.Join(splitString, " ")
}

// Convert MongoDB partial filter expression to Couchbase WHERE clause
func ConvertMongoToCouchbase(expression map[string]interface{}, fieldPath common.IndexFieldPath) string {
	return processExpression(expression, fieldPath)
}

// Recursively process the MongoDB expression
func processExpression(expression map[string]interface{}, fieldPath common.IndexFieldPath) string {
	var conditions []string

	for key, value := range expression {
		switch key {
		case "$and", "$or", "$not":
			// Directly handle logical operators, translating them to N1QL syntax
			logicalCondition := processLogicalOperator(key, value, fieldPath)
			if logicalCondition != "" {
				conditions = append(conditions, logicalCondition)
			}
		default:
			// Process comparison operators
			fieldCondition := processField(key, value, fieldPath)
			if fieldCondition != "" {
				conditions = append(conditions, fieldCondition)
			}
		}
	}

	return strings.Join(conditions, " AND ")
}

// Handle logical operators by processing each contained expression
func processLogicalOperator(operator string, value interface{}, fieldPath common.IndexFieldPath) string {
	var conditions []string
	var opSymbol string

	switch operator {
	case "$and":
		opSymbol = "AND"
	case "$or":
		opSymbol = "OR"
	case "$not":
		opSymbol = "NOT"
	}

	switch val := value.(type) {
	case []interface{}:
		for _, expr := range val {
			condition := processExpression(expr.(map[string]interface{}), fieldPath)
			if condition != "" {
				conditions = append(conditions, condition)
			}
		}
		return fmt.Sprintf("(%s)", strings.Join(conditions, fmt.Sprintf(" %s ", opSymbol)))
	case map[string]interface{}:
		if operator == "$not" {
			condition := processExpression(val, fieldPath)
			return fmt.Sprintf("NOT (%s)", condition)
		}
	}

	return ""
}

// Process individual field conditions
func processField(field string, value interface{}, fieldPath common.IndexFieldPath) string {
	field = fieldPath.Get(field)
	switch v := value.(type) {
	case map[string]interface{}:
		return convertOperator(field, v)
	default:
		if strings.Index(field, "[]") > -1 {
			conditionSuffix := fmt.Sprintf("%s %v", "=", value)
			arrFieldExpression := GenerateArrayFilterExpression(field)
			return fmt.Sprintf(arrFieldExpression, conditionSuffix)
		}
		// Handle direct equality as a special case
		return fmt.Sprintf("`%s` = %v", field, value)
	}
}

// Convert MongoDB comparison operators to their Couchbase equivalents
func convertOperator(field string, operators map[string]interface{}) string {
	var conditions []string
	for op, val := range operators {
		couchbaseOp := ""
		switch op {
		case "$gt":
			couchbaseOp = ">"
		case "$gte":
			couchbaseOp = ">="
		case "$lt":
			couchbaseOp = "<"
		case "$lte":
			couchbaseOp = "<="
		case "$eq":
			couchbaseOp = "="
		case "$ne":
			couchbaseOp = "!="
		case "$exists":
			v, ok := val.(bool)
			if ok && v {
				couchbaseOp = "IS NOT NULL"
			}
			if ok && !v {
				couchbaseOp = "IS NULL"
			}
		}
		condition := ""
		if strings.Index(field, "[]") > -1 {
			conditionSuffix := fmt.Sprintf("%s %v", couchbaseOp, val)
			arrFieldExpression := GenerateArrayFilterExpression(field)
			condition = fmt.Sprintf(arrFieldExpression, conditionSuffix)
		}
		condition = fmt.Sprintf("`%s` %s %v", field, couchbaseOp, val)
		conditions = append(conditions, condition)
	}
	return strings.Join(conditions, " AND ")
}

func GenerateArrayFilterExpression(input string) string {
	parts := strings.Split(input, "[]")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimPrefix(parts[i], ".")
	}
	return createArrayFilterExpression(parts, "", 0)
}

func createArrayFilterExpression(parts []string, parent string, l int) string {
	l++
	if len(parts) == 1 {
		if parts[0] == "" {
			return fmt.Sprintf("%s %s", parent, "%s")
		}
		return fmt.Sprintf("%s.%s %s", parent, formatFieldReferenceWithAddtionalOptions(parts[0]), "%s")
	}
	item := fmt.Sprintf("`l%dItem`", l)
	items := formatFieldReference(parts[0])
	if parent != "" {
		items = parent + "." + items
	}
	inner := createArrayFilterExpression(parts[1:], item, l)

	return fmt.Sprintf("ANY %s IN %s SATISFIES (%s) END", item, items, inner)
}
