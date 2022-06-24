package loggerInjector

import (
	"encoding/json"
	"strconv"
)

func ConvertToIntOrDefault(str string) int {
	if str != "" {
		number, err := strconv.Atoi(str)
		if err != nil {
			return 0
		}
		return number
	}
	return 0
}

func ConvertToBooleanOrDefault(str string) bool {
	if str != "" {
		obj, err := strconv.ParseBool(str)
		if err != nil {
			return false
		}
		return obj
	}
	return false
}

func MapToStruct(this map[string]interface{}, obj interface{}) error {
	data, err := json.Marshal(this)
	if err != nil {

		return err
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	return nil
}
