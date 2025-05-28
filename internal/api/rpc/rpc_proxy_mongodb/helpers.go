package rpc_proxy_mongodb

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sparallel_server/pkg/foundation/errs"
	"time"
)

const dateFormat = time.RFC3339

const typeKey = "|t_"
const valueKey = "|v_"

const datetimeType = "datetime"
const idType = "id"

func processDateValues(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 2 && v[typeKey] == datetimeType {
			if timeStr, ok := v[valueKey].(string); ok {
				if t, err := time.Parse(dateFormat, timeStr); err == nil {
					return primitive.NewDateTimeFromTime(t)
				}
			}
			return v
		}

		if len(v) == 2 && v[typeKey] == idType {
			if idStr, ok := v[valueKey].(string); ok {
				if objectID, err := primitive.ObjectIDFromHex(idStr); err == nil {
					return objectID
				}
			}
			return v
		}

		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = processDateValues(value)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = processDateValues(value)
		}
		return result

	default:
		return v
	}
}

func unmarshalJson(data string) (interface{}, error) {
	var document interface{}

	err := json.Unmarshal([]byte(data), &document)

	if err != nil {
		return nil, err
	}

	return processDateValues(document), nil
}

func serializeForPHPMongoDB(doc interface{}) (string, error) {
	bsonData, err := bson.Marshal(doc)

	if err != nil {
		return "", errs.Err(fmt.Errorf("error BSON marshaling: %w", err))
	}

	var raw bson.Raw = bsonData

	jsonData, err := bson.MarshalExtJSON(raw, true, true)

	if err != nil {
		return "", errs.Err(fmt.Errorf("converting error in Extended Json: %w", err))
	}

	return string(jsonData), nil
}
