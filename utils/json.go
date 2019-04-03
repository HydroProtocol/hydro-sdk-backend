package utils

import (
	"encoding/json"
	"log"
)

func ToJsonString(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		log.Println("json err:", err)
		return ""
	}

	return string(b)
}
