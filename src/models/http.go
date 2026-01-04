package models

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func listRows[T BaserowData](url string, c *BaserowClient, instance T) (BaserowQueryResponse[T], error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return BaserowQueryResponse[T]{}, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Token "+c.ApiKey)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return BaserowQueryResponse[T]{}, fmt.Errorf("Error making API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return BaserowQueryResponse[T]{}, fmt.Errorf("API request failed with status: %s and error reading body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)

		return BaserowQueryResponse[T]{}, fmt.Errorf("API request failed with status: %s and body: %s", resp.Status, bodyString)
	}

	data, err := instance.UnmarshalJSON(resp.Body)
	response, ok := data.(BaserowQueryResponse[T])
	if !ok {
		return BaserowQueryResponse[T]{}, fmt.Errorf("error asserting type to BaserowQueryResponse")
	}

	return response, nil
}

func ListRows[T BaserowData](c *BaserowClient) ([]T, error) {
	var instance T
	if reflect.TypeOf(instance).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("ListRows: type T must be a pointer type")
	}

	instance = reflect.New(reflect.TypeOf(instance).Elem()).Interface().(T)

	url := fmt.Sprintf("%s/api/database/rows/table/%s/?user_field_names=true", c.BaseURL, instance.GetTableID())

	var results []T
	var count int
	for {
		resp, err := listRows[T](url, c, instance)
		if err != nil {
			return nil, fmt.Errorf("Error listing rows: %v, url: %s", err, url)
		}

		results = append(results, resp.Results...)
		count = resp.Count

		if resp.Next == "" {
			break
		} else {
			url = resp.Next
		}
	}

	if len(results) != count {
		return nil, fmt.Errorf("Mismatch in expected count and results length. Expected: %d, Got: %d", count, len(results))
	}

	return results, nil
}
