package rpc_proxy_mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func processDateValues(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339, strValue); err == nil {
					v[key] = primitive.NewDateTimeFromTime(t)
				}
			} else if mapValue, ok := value.(map[string]interface{}); ok {
				v[key] = processDateValues(mapValue)
			} else if sliceValue, ok := value.([]interface{}); ok {
				v[key] = processDateValues(sliceValue)
			}
		}
		return v
	case []interface{}:
		for i, value := range v {
			if strValue, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339, strValue); err == nil {
					v[i] = primitive.NewDateTimeFromTime(t)
				}
			} else if mapValue, ok := value.(map[string]interface{}); ok {
				v[i] = processDateValues(mapValue)
			} else if sliceValue, ok := value.([]interface{}); ok {
				v[i] = processDateValues(sliceValue)
			}
		}
		return v
	default:
		return data
	}
}
