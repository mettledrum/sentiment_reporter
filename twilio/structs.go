package twilio

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type score float64

type addOns struct {
	Results struct {
		IBMWatsonSentiment struct {
			Result struct {
				DocSentiment struct {
					Type  string `json:"type"`
					Score score  `json:"score"`
				} `json:"docSentiment"`
			} `json:"result"`
		} `json:"ibm_watson_sentiment"`
	} `json:"results"`
}

type info struct {
	From    string  `json:"from"`
	Type    string  `json:"type"`
	Score   float64 `json:"score"`
	Content string  `json:"content"`
}

// IBM docSentiment.score is returned as string :(
// convert it to float64
func (s *score) UnmarshalJSON(d []byte) error {
	var str string
	err := json.Unmarshal(d, &str)
	if err != nil {
		return err
	}

	fl, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	*s = score(fl)

	return nil
}

func GetInfo(v url.Values) (info, error) {
	ao := addOns{}
	b := []byte(v.Get("AddOns"))
	err := json.Unmarshal(b, &ao)
	if err != nil {
		return info{}, err
	}

	ds := ao.Results.IBMWatsonSentiment.Result.DocSentiment
	return info{
		Content: v.Get("Body"),
		From:    v.Get("From"),
		Score:   float64(ds.Score),
		Type:    ds.Type,
	}, nil
}
