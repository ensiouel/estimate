package dto

import (
	"encoding/json"
	"estimate/pkg/apperror"
	"github.com/goware/urlx"
	"time"
)

type Duration struct {
	time.Duration
}

func (duration Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(duration.String())
}

type GetWebsiteAccessTimeRequest struct {
	URL string `query:"url"`
}

func (request GetWebsiteAccessTimeRequest) Validate() error {
	_, err := urlx.Parse(request.URL)
	if err != nil {
		return apperror.BadRequest.WithMessage("invalid url")
	}

	return nil
}

type GetWebsiteAccessTimeResponse struct {
	AccessTime  Duration  `json:"access_time"`
	LastCheckAt time.Time `json:"last_check_at"`
}

type GetWebsiteWithMinAccessTimeResponse struct {
	URL         string    `json:"url"`
	AccessTime  Duration  `json:"access_time"`
	LastCheckAt time.Time `json:"last_check_at"`
}

type GetWebsiteWithMaxAccessTimeResponse struct {
	URL         string    `json:"url"`
	AccessTime  Duration  `json:"access_time"`
	LastCheckAt time.Time `json:"last_check_at"`
}
