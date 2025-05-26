package rpc_proxy_mongodb

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func processDateValues(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 2 && v["|t_"] == "datetime" {
			if timeStr, ok := v["|v_"].(string); ok {
				if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
					return primitive.NewDateTimeFromTime(t)
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

// TODO: check
func serializeForPHPMongoDB(doc interface{}) (string, error) {
	bsonData, err := bson.Marshal(doc)

	if err != nil {
		return "", fmt.Errorf("ошибка BSON маршалинга: %w", err)
	}

	var raw bson.Raw = bsonData

	jsonData, err := bson.MarshalExtJSON(raw, true, true)

	if err != nil {
		return "", fmt.Errorf("ошибка конвертации в Extended JSON: %w", err)
	}

	return string(jsonData), nil
}
