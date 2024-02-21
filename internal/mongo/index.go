package mongo

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	INCLUDE_MISSSING = " INCLUDE MISSING"
	ASC              = " ASC"
	DESC             = " DESC"
)

type Index struct {
	Name              string
	Keys              []Key
	PartialExpression map[string]interface{}
	Unique            bool
	Sparse            bool
	NotSupported      bool
}
type Key struct {
	Field string
	Order int
}

// IndexFieldPath is used to have array representation for a particular path
// example: k1.n1k1.n2k1 is path for field n2k1 in a document. n1k1 is an array, and it is represented as k1.n1k1[].n2k1.
type IndexFieldPath map[string]string

func (i IndexFieldPath) Get(key string) string {
	if i == nil {
		return key
	}
	v := i[key]
	if v == "" {
		return key
	}
	return v
}

func CreateIndexQuery(bucket, scope, collection string, index Index, fieldPath IndexFieldPath) (string, error) {
	var arrFields []Key
	isArrayFieldAtFistPos := false
	for i, key := range index.Keys {
		key.Field = fieldPath.Get(key.Field)
		if strings.Index(key.Field, "[]") > 0 {
			if i == 0 {
				isArrayFieldAtFistPos = true
			}
			arrFields = append(arrFields, key)
		}
	}
	arrayIndexExp, err := GroupAndCombine(arrFields, !index.Sparse && isArrayFieldAtFistPos)
	if err != nil {
		return "", err
	}
	arrIndex := true
	var fields []string

	for i, key := range index.Keys {
		includeMissing := false
		if i == 0 && !index.Sparse {
			includeMissing = true
		}
		switch {
		// I am grouping all the array notation fields into single flatten couchbase array index expression
		case strings.Index(fieldPath.Get(key.Field), "[]") > 0:
			if arrIndex {
				keyAttribs := ""
				if len(arrFields) == 0 {
					keyAttribs = getLeadKeyAttr(key.Order, includeMissing)
				}
				field := fmt.Sprintf("%s%s", GenerateCouchbaseArrayIndex(arrayIndexExp), keyAttribs)
				fields = append(fields, field)
				arrIndex = false
			}
		default:
			fields = append(fields, getField(key.Field, includeMissing, key.Order))
		}
	}
	reg := regexp.MustCompile(`[^A-Za-z0-9#_]`)
	// Replace characters that do not match the pattern with "_"
	name := reg.ReplaceAllString(index.Name, "_")
	partialFilter, err := ConvertMongoToCouchbase(index.PartialExpression, fieldPath)
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf(
		"create index `%s` on `%s`.`%s`.`%s` (%s) %s",
		name, bucket, scope, collection, strings.Join(fields, ","), partialFilter)
	return query, nil
}

func getField(field string, includeMissing bool, order int) string {
	return fmt.Sprintf("%s%s", formatFieldReference(field), getLeadKeyAttr(order, includeMissing))
}
func getLeadKeyAttr(order int, includeMissing bool) string {
	im := INCLUDE_MISSSING
	indexOrder := ASC
	if !includeMissing {
		im = ""
	}
	if order == -1 {
		indexOrder = DESC
	}
	return fmt.Sprintf("%s%s", indexOrder, im)
}

// GroupAndCombine array fields are combined because only one array field can be indexed in a compound index
func GroupAndCombine(keys []Key, includeMissing bool) (string, error) {
	// Assuming all keys have the same prefix for simplicity, as demonstrated in the combineStrings function
	var prefix string
	var combined []string
	keyLen := len(keys)
	for i, key := range keys {
		lastIndex := strings.LastIndex(key.Field, "[]")
		if lastIndex != -1 {
			tempPrefix := key.Field[:lastIndex+2]
			if prefix == "" {
				prefix = tempPrefix
			}
			if tempPrefix != prefix {
				return "", errors.New("multiple array reference")
			}

			suffix := key.Field[lastIndex+2:]

			keyAttr := ""
			if keyLen > 1 {
				keyAttr = getLeadKeyAttr(key.Order, includeMissing && i == 0)
			}
			combinedPart := fmt.Sprintf("%s%s", suffix, keyAttr)
			combined = append(combined, combinedPart)
		}
	}

	// Join the combined parts with commas and prepend the prefix
	result := prefix + strings.Join(combined, ",")
	return result, nil
}

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
func ConvertMongoToCouchbase(expression map[string]interface{}, fieldPath IndexFieldPath) (string, error) {
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
func processExpression(expression map[string]interface{}, fieldPath IndexFieldPath) (string, error) {
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
	l := len(conditions)
	switch {
	case l > 1:
		return "(" + strings.Join(conditions, " AND ") + ")", nil
	case l == 1:
		return conditions[0], nil
	}

	return "", nil
}

// Handle logical operators by processing each contained expression
func processLogicalOperator(operator string, value interface{}, fieldPath IndexFieldPath) (string, error) {
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
func ProcessField(field string, value interface{}, fieldPath IndexFieldPath) (string, error) {
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
