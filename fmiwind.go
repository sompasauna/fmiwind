package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sompasauna/fmiwind/version"
)

type Point struct {
	Pos string `xml:"pos"`
}

type BsWfsElement struct {
	Location       Point   `xml:"Location>Point"`
	Time           string  `xml:"Time"`
	ParameterName  string  `xml:"ParameterName"`
	ParameterValue float64 `xml:"ParameterValue"`
}

type FeatureCollection struct {
	XMLName xml.Name       `xml:"FeatureCollection"`
	Members []BsWfsElement `xml:"member>BsWfsElement"`
}

var (
	influxURL    = flag.String("influx-url", "http://localhost:8086", "InfluxDB server URL")
	influxToken  = flag.String("influx-token", "", "InfluxDB authentication token")
	influxOrg    = flag.String("influx-org", "", "InfluxDB organization")
	influxBucket = flag.String("influx-bucket", "", "InfluxDB bucket")
	place        = flag.String("place", "kaisaniemi,helsinki", "Place to query")
	versionF     = flag.Bool("v", false, "Print version and exit")
)

func main() {
	flag.Parse()
	if *versionF {
		fmt.Printf("fmiwind %s\n", version.Version)
		os.Exit(0)
	}

	url := "http://opendata.fmi.fi/wfs?request=getFeature&storedquery_id=fmi::observations::weather::simple&place=" + *place + "&parameters=winddirection,windspeedms,pressure,humidity"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var fc FeatureCollection
	err = xml.Unmarshal(body, &fc)
	if err != nil {
		panic(err)
	}

	sort.Slice(fc.Members, func(i, j int) bool {
		return fc.Members[i].Time > fc.Members[j].Time
	})

	params := make(map[string]any)

	for _, member := range fc.Members {
		if _, ok := params[member.ParameterName]; !ok {
			params[member.ParameterName] = member.ParameterValue
		}
		if len(params) == 4 {
			break
		}
	}

	if len(params) == 0 {
		println("no data")
		os.Exit(1)
	}

	if *influxURL == "" || *influxToken == "" || *influxOrg == "" || *influxBucket == "" {
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(params); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	client := influxdb.NewClient(*influxURL, *influxToken)
	defer client.Close()
	writeAPI := client.WriteAPI(*influxOrg, *influxBucket)
	p := influxdb.NewPoint(
		"weather",
		map[string]string{"location": "kaisaniemi"},
		params,
		time.Now(),
	)
	writeAPI.WritePoint(p)
	writeAPI.Flush()
}
