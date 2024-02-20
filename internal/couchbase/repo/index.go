package repo

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/index"
	"reflect"
	"strconv"
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

// ConvertMongoToCouchbase Convert MongoDB partial filter expression to Couchbase WHERE clause
func ConvertMongoToCouchbase(expression map[string]interface{}, fieldPath index.IndexFieldPath) (string, error) {
	exp, err := processExpression(expression, fieldPath)
	if err != nil {
		return "", err
	}
	if exp != "" {
		exp = "WHERE " + exp
	}
	return exp, nil
}

// Recursively process the MongoDB expression
func processExpression(expression map[string]interface{}, fieldPath index.IndexFieldPath) (string, error) {
	var conditions []string

	for key, value := range expression {
		switch key {
		case "$and", "$or", "$not":
			// Directly handle logical operators, translating them to N1QL syntax
			logicalCondition, err := processLogicalOperator(key, value, fieldPath)
			if err != nil {
				return "", err
			}
			if logicalCondition != "" {
				conditions = append(conditions, logicalCondition)
			}
		default:
			// Process comparison operators
			fieldCondition, err := ProcessField(key, value, fieldPath)
			if err != nil {
				return "", err
			}
			if fieldCondition != "" {
				conditions = append(conditions, fieldCondition)
			}
		}
	}

	return strings.Join(conditions, " AND "), nil
}

// Handle logical operators by processing each contained expression
func processLogicalOperator(operator string, value interface{}, fieldPath index.IndexFieldPath) (string, error) {
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
			condition, err := processExpression(expr.(map[string]interface{}), fieldPath)
			if err != nil {
				return "", err
			}
			if condition != "" {
				conditions = append(conditions, condition)
			}
		}
		return fmt.Sprintf("(%s)", strings.Join(conditions, fmt.Sprintf(" %s ", opSymbol))), nil
	case map[string]interface{}:
		if operator == "$not" {
			condition, err := processExpression(val, fieldPath)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("NOT (%s)", condition), nil
		}
	}

	return "", nil
}

// ProcessField Process individual field conditions
func ProcessField(field string, value interface{}, fieldPath index.IndexFieldPath) (string, error) {
	field = fieldPath.Get(field)
	switch v := value.(type) {
	case map[string]interface{}:
		return convertOperator(field, v)
	default:
		if strings.Index(field, "[]") > -1 {
			conditionSuffix := fmt.Sprintf("%s %#v", "=", value)
			arrFieldExpression := GenerateArrayFilterExpression(field, false)
			return fmt.Sprintf(arrFieldExpression, conditionSuffix), nil
		}
		// Handle direct equality as a special case
		return fmt.Sprintf("%s = %#v", formatFieldReference(field), value), nil
	}
}

// Convert MongoDB comparison operators to their Couchbase equivalents
func convertOperator(field string, operators map[string]interface{}) (string, error) {
	var conditions []string
	isTypeOperand := false
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
				couchbaseOp = "IS NOT"
			}
			if ok && !v {
				couchbaseOp = "IS"
			}
			val = "NULL"
		case "$type":
			couchbaseOp = "="
			t, err := getCBType(val)
			if err != nil {
				return "", err
			}
			val = t
			if strings.Index(t, ",") > -1 {
				couchbaseOp = "IN"
				val = "(" + t + ")"
			}
			isTypeOperand = true
		default:
			return "", fmt.Errorf("operand %s cannot be parsed to couchbase operand", op)
		}
		condition := ""
		if strings.Index(field, "[]") > -1 {
			conditionSuffix := fmt.Sprintf("%s %#v", couchbaseOp, val)
			arrFieldExpression := GenerateArrayFilterExpression(field, isTypeOperand)
			condition = fmt.Sprintf(arrFieldExpression, conditionSuffix)
		} else {
			if isTypeOperand {
				field = "type(" + formatFieldReference(field) + ")"
			}
			condition = fmt.Sprintf("%s %s %#v", field, couchbaseOp, val)
		}
		conditions = append(conditions, condition)
	}
	return strings.Join(conditions, " AND "), nil
}

func getCBType(val interface{}) (string, error) {
	switch v := val.(type) {
	case int32:
		pv, ok := mongoCBPartialAlias[strconv.Itoa(int(v))]
		if !ok {
			return "", fmt.Errorf("type %d cannot be parsed to couchbase type", v)
		}
		return pv, nil
	case string:
		pv, ok := mongoCBPartialAlias[v]
		if !ok {
			return "", fmt.Errorf("type %s cannot be parsed to couchbase type", v)
		}
		return pv, nil
	case []interface{}:
		var types []string
		for _, iv := range v {
			pv, err := getCBType(iv)
			if err != nil {
				return "", err
			}
			types = append(types, pv)
		}
		return strings.Join(types, ","), nil
	default:
		return "", fmt.Errorf("invalid type %#v", reflect.TypeOf(val).String())
	}
	return "", nil
}

func GenerateArrayFilterExpression(input string, isTypeFilter bool) string {
	parts := strings.Split(input, "[]")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimPrefix(parts[i], ".")
	}
	return createArrayFilterExpression(parts, "", 0, isTypeFilter)
}

func createArrayFilterExpression(parts []string, parent string, l int, isTypeFilter bool) string {
	l++
	if len(parts) == 1 {
		if parts[0] == "" {
			filter := parent
			if isTypeFilter {
				filter = "type(" + filter + ")"
			}
			return fmt.Sprintf("%s %s", filter, "%s")
		}
		filter := formatFieldReference(parts[0])
		if isTypeFilter {
			filter = "type(" + filter + ")"
		}
		return fmt.Sprintf("%s.%s %s", parent, filter, "%s")
	}
	item := fmt.Sprintf("`l%dItem`", l)
	items := formatFieldReference(parts[0])
	if parent != "" {
		items = parent + "." + items
	}
	inner := createArrayFilterExpression(parts[1:], item, l, isTypeFilter)

	return fmt.Sprintf("ANY %s IN %s SATISFIES (%s) END", item, items, inner)
}

var mongoCBPartialAlias = make(map[string]string)

type mongoCBPartialAliasList struct {
	mongoTypeNumber int
	mongoTypeString string
	couchbase       string
}

func init() {
	var list []mongoCBPartialAliasList
	list = append(
		list,
		mongoCBPartialAliasList{
			mongoTypeNumber: 1,
			mongoTypeString: "double",
			couchbase:       "number",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 16,
			mongoTypeString: "int",
			couchbase:       "number",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 18,
			mongoTypeString: "long",
			couchbase:       "number",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 19,
			mongoTypeString: "decimal",
			couchbase:       "number",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 2,
			mongoTypeString: "string",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 3,
			mongoTypeString: "object",
			couchbase:       "object",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 4,
			mongoTypeString: "array",
			couchbase:       "array",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 5,
			mongoTypeString: "binData",
			couchbase:       "binary",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 6,
			mongoTypeString: "undefined",
			couchbase:       "null",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 7,
			mongoTypeString: "objectId",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 8,
			mongoTypeString: "bool",
			couchbase:       "bool",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 9,
			mongoTypeString: "date",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 10,
			mongoTypeString: "null",
			couchbase:       "null",
		},
		//mongoCBPartialAliasList{
		//	mongoTypeNumber: 11,
		//	mongoTypeString: "regex",
		//	couchbase:       "string",
		//},
		//mongoCBPartialAliasList{
		//	mongoTypeNumber: 12,
		//	mongoTypeString: "dbPointer",
		//	couchbase:       "string",
		//},
		mongoCBPartialAliasList{
			mongoTypeNumber: 13,
			mongoTypeString: "javascript",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 14,
			mongoTypeString: "symbol",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 15,
			mongoTypeString: "javascriptWithScope",
			couchbase:       "string",
		},
		mongoCBPartialAliasList{
			mongoTypeNumber: 17,
			mongoTypeString: "timestamp",
			couchbase:       "number",
		},
	)
	for _, v := range list {
		mongoCBPartialAlias[v.mongoTypeString] = v.couchbase
		mongoCBPartialAlias[strconv.Itoa(v.mongoTypeNumber)] = v.couchbase
	}
}
