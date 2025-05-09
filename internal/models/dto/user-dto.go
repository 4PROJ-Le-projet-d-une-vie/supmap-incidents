package dto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supmap-users/internal/config"
)

type RoleDTO struct {
	Name string `json:"name"`
}

type PartialUserDTO struct {
	ID     int64    `json:"id"`
	Handle string   `json:"handle"`
	Role   *RoleDTO `json:"role"`
}

func UserIdToDTO(userId int64) (*PartialUserDTO, error) {
	res, err := http.Get(fmt.Sprintf("%s/internal/users/%d", config.UsersBaseUrl, userId))
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
