package twilio

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Info struct {
	From    string  `json:"from"`
	Type    string  `json:"type"`
	Score   float64 `json:"score"`
	Content string  `json:"content"`
}

// {"status":"successful","message":null,"code":null,"results":{"marchex_sentiment":{"request_sid":"asdf","status":"successful","message":null,"code":null,"result":{"result":0.6560785174369812}}}}

// GetInfo parses values from Twilio API
func GetInfo(v url.Values) (Info, error) {
	type resp struct {
		Results struct {
			MarchexSentiment struct {
				Result struct {
					Result float64 `json:"result"`
				} `json:"result"`
			} `json:"marchex_sentiment"`
		} `json:"results"`
	}

	b := []byte(v.Get("AddOns"))
	fmt.Printf("\n%+v\n", string(b))
	r := resp{}
	err := json.Unmarshal(b, &r)
	if err != nil {
		return Info{}, err
	}

	s := r.Results.MarchexSentiment.Result.Result
	return Info{
		Content: v.Get("Body"),
		From:    v.Get("From"),
		Score:   s,
		Type:    getType(s),
	}, nil
}

func getType(s float64) string {
	if s < 0.5 {
		return "negative"
	}
	return "positive"
}
