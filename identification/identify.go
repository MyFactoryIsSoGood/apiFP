package identification

import (
	"apiFP/compare"
	"apiFP/db"
	"apiFP/template/utils"
	"encoding/json"
	"fmt"
	"runtime"
)

type IdentificationResponse struct {
	MatchesWith string  `json:"matches_with"`
	Features    int     `json:"features"`
	Likeness    float64 `json:"likeness"`
}

func Identify(from, to int, tmp utils.Template, responseCh chan<- IdentificationResponse) {
	for i := from; i < to; i++ {
		byteFP, name := db.Take(i)

		current := compare.ISO19794_BytesToTemplate(byteFP)

		percent := compare.Compare(tmp, current)
		if percent > 0.1 {
			fmt.Println(i, percent)
		}
		if percent > 0.5 {
			responseCh <- IdentificationResponse{
				MatchesWith: name,
				Features:    int(current.Fingerprints[0].MinutiaeCount),
				Likeness:    percent,
			}
		}
	}
}

func IdentifyInParallel(tmp utils.Template) []byte {
	results := make(chan IdentificationResponse, 100)
	var fpCount int
	db.DB.Table("iso_templates_strs").Count(&fpCount)
	for w := 1; w < runtime.NumCPU(); w++ {
		go Identify(fpCount/runtime.NumCPU()*w, fpCount/runtime.NumCPU()*(w+1), tmp, results)
	}

	var a IdentificationResponse
	a = <-results
	close(results)

	out, _ := json.Marshal(a)
	return out
}
