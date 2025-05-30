package rpc_proxy_mongodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"sparallel_server/pkg/foundation/errs"
	"time"
)

const dateFormat = time.RFC3339

const typeKey = "|t_"
const valueKey = "|v_"

const datetimeType = "datetime"
const idType = "id"

type WriteModelWrapper struct {
	Type  string          `json:"type"`
	Model json.RawMessage `json:"model"`
}

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

func unmarshalModels(data string) ([]mongo.WriteModel, error) {
	var wrappers []WriteModelWrapper

	if err := json.Unmarshal([]byte(data), &wrappers); err != nil {
		return nil, err
	}

	models := make([]mongo.WriteModel, 0, len(wrappers))

	for _, wrapper := range wrappers {
		var model mongo.WriteModel

		switch wrapper.Type {
		case "insertOne":
			var im struct {
				Document interface{} `json:"document"`
			}
			if err := json.Unmarshal(wrapper.Model, &im); err != nil {
				return nil, errors.New("insertOne [" + err.Error() + "]")
			}
			model = mongo.NewInsertOneModel().SetDocument(processDateValues(im.Document))
		case "updateOne":
			var um struct {
				Filter interface{} `json:"filter"`
				Update interface{} `json:"update"`
				Upsert *bool       `json:"upsert,omitempty"`
			}
			if err := json.Unmarshal(wrapper.Model, &um); err != nil {
				return nil, errors.New("updateOne [" + err.Error() + "]")
			}
			model = mongo.NewUpdateOneModel().
				SetFilter(processDateValues(um.Filter)).
				SetUpdate(processDateValues(um.Update))
			if um.Upsert != nil {
				model.(*mongo.UpdateOneModel).SetUpsert(*um.Upsert)
			}

		case "updateMany":
			var um struct {
				Filter interface{} `json:"filter"`
				Update interface{} `json:"update"`
				Upsert *bool       `json:"upsert,omitempty"`
			}
			if err := json.Unmarshal(wrapper.Model, &um); err != nil {
				return nil, errors.New("updateMany [" + err.Error() + "]")
			}
			model = mongo.NewUpdateManyModel().
				SetFilter(processDateValues(um.Filter)).
				SetUpdate(processDateValues(um.Update))
			if um.Upsert != nil {
				model.(*mongo.UpdateManyModel).SetUpsert(*um.Upsert)
			}

		case "deleteOne":
			var dm struct {
				Filter interface{} `json:"filter"`
			}
			if err := json.Unmarshal(wrapper.Model, &dm); err != nil {
				return nil, errors.New("deleteOne [" + err.Error() + "]")
			}
			model = mongo.NewDeleteOneModel().SetFilter(processDateValues(dm.Filter))

		case "deleteMany":
			var dm struct {
				Filter interface{} `json:"filter"`
			}
			if err := json.Unmarshal(wrapper.Model, &dm); err != nil {
				return nil, errors.New("deleteMany [" + err.Error() + "]")
			}
			model = mongo.NewDeleteManyModel().SetFilter(processDateValues(dm.Filter))

		case "replaceOne":
			var rm struct {
				Filter      interface{} `json:"filter"`
				Replacement interface{} `json:"replacement"`
				Upsert      *bool       `json:"upsert,omitempty"`
			}
			if err := json.Unmarshal(wrapper.Model, &rm); err != nil {
				return nil, errors.New("replaceOne [" + err.Error() + "]")
			}
			model = mongo.NewReplaceOneModel().
				SetFilter(processDateValues(rm.Filter)).
				SetReplacement(processDateValues(rm.Replacement))
			if rm.Upsert != nil {
				model.(*mongo.ReplaceOneModel).SetUpsert(*rm.Upsert)
			}

		default:
			return nil, fmt.Errorf("unknown type of model: %s", wrapper.Type)
		}

		models = append(models, model)
	}

	return models, nil

}
