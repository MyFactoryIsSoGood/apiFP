package identification

import (
	"apiFP/compare"
	"apiFP/db"
	"apiFP/template/utils"
	"encoding/json"
)

type IdentificationResponse struct {
	MatchesWith string  `json:"matches_with"`
	Features    int     `json:"features"`
	Likeness    float64 `json:"likeness"`
}

// Identify is worker function for IdentifyInParallel. Takes templates from db and finds the best match
func Identify(from, to int, tmp utils.Template, responseCh chan<- IdentificationResponse) {
	maxPercent := 0.0
	maxName := ""
	maxFeatures := 0
	for i := from; i < to; i++ {
		byteFP, name := db.Take(i)

		current := compare.ISO19794_BytesToTemplate(byteFP)

		percent := compare.Compare(tmp, current)
		if percent > maxPercent {
			maxPercent = percent
			maxName = name
			maxFeatures = int(current.Fingerprints[0].MinutiaeCount)
		}
	}
	responseCh <- IdentificationResponse{
		MatchesWith: maxName,
		Features:    maxFeatures,
		Likeness:    maxPercent,
	}
}

// IdentifyInParallel is function for identification. It takes template and returns the best match
func IdentifyInParallel(tmp utils.Template) []byte {
	results := make(chan IdentificationResponse, 100)
	var fpCount int
	db.DB.Table("iso_templates_strs").Count(&fpCount)
	for w := 1; w < 10; w++ {
		go Identify(fpCount/10*w, fpCount/10*(w+1), tmp, results)
	}

	var a IdentificationResponse
	var counter = 0
	for b := range results {
		counter++
		if b.Likeness > a.Likeness {
			a = b
		}
		if counter == 9 {
			break
		}
	}
	close(results)

	if a.Likeness < 0.5 {
		return []byte("no matches")
	}
	out, _ := json.Marshal(a)
	return out
}
