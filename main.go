package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/imroc/req/v3"
)

type Stroom struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"page_title"`
}

type Adres struct {
	BagID      string `json:"bagid"`
	Postcode   string `json:"postcode"`
	Huisnummer string `json:"huisnummer"`
}

type Moment struct {
	StroomID int    `json:"afvalstroom_id"`
	Datum    string `json:"ophaaldatum"`
}

func main() {
	postcode := "CHANGE_THIS_OTHERWISE_NO_WORK"
	huisnummer := "CHANGE_THIS_OTHERWISE_NO_WORK"

	client := req.C()
	adresResp, aerr := client.R().
		Get(fmt.Sprintf("https://afvalkalender.alphenaandenrijn.nl/adressen/%s:%s", postcode, huisnummer))
	if aerr != nil {
		panic(aerr)
	}

	var adressen []Adres
	json.Unmarshal([]byte(adresResp.String()), &adressen)

	adres := adressen[0]

	if adres.BagID == "" {
		panic("Can't find BagID")
	}

	stromenResp, serr := client.R().
		Get(fmt.Sprintf("https://afvalkalender.alphenaandenrijn.nl/rest/adressen/%s/afvalstromen", adres.BagID))
	if serr != nil {
		panic(serr)
	}

	momentResp, merr := client.R().
		Get(fmt.Sprintf("https://afvalkalender.alphenaandenrijn.nl/rest/adressen/%s/kalender/%d", adres.BagID, time.Now().Year()))
	if merr != nil {
		panic(merr)
	}

	var stromen []Stroom
	json.Unmarshal([]byte(stromenResp.String()), &stromen)

	var momenten []Moment
	json.Unmarshal([]byte(momentResp.String()), &momenten)

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	cal.SetName("Afvalkalender")

	for _, moment := range momenten {
		day, _ := time.Parse("2006-01-02", moment.Datum)

		for _, stroom := range stromen {
			if moment.StroomID != stroom.ID {
				continue
			}

			event := cal.AddEvent(day.Format("2006-01-02"))
			event.SetCreatedTime(time.Now())
			event.SetAllDayStartAt(day)
			event.SetAllDayStartAt(day)
			event.SetSummary(stroom.Title)
			event.SetDescription(stroom.Title)

			alarm := event.AddAlarm()
			alarm.SetTrigger("-PT120M")
			alarm.SetAction(ics.ActionDisplay)
			alarm.SetProperty(ics.ComponentPropertyDescription, stroom.Description)
		}
	}

	os.WriteFile("calendar.ics", []byte(cal.Serialize()), 0666)
}
