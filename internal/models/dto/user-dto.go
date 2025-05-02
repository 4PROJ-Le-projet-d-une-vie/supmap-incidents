package dto

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PartialUserDTO struct {
	ID             int64     `json:"id"`
	Handle         string    `json:"handle"`
}

func UserIdToDTO(userId int64) (*PartialUserDTO, error) {
	fmt.Println(userId)
	res, err := http.Get(fmt.Sprintf("http://localhost:8080/internal/user/%d", userId))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d returned for user id %d", res.StatusCode, userId)
	}

	var body PartialUserDTO
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
