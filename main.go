package main

// 参考にした
// https://zenn.dev/empenguin/articles/9ce4b7dd4edb66
// https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
// https://help.raritan.com/json-rpc/pdu/v3.4.0/pdumodel.html

import (
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/ybbus/jsonrpc/v2"
)

type (
	ResultChild struct {
		Json struct {
			Result struct {
				Value struct {
					Valid bool    `json:"valid"`
					Value float64 `json:"value"`
				} `json:"_ret_"`
			} `json:"result"`
			Id int `json:"id"`
		} `json:"json"`
	}

	Result struct {
		Responses []*ResultChild `json:"responses"`
	}

	BulkParamRpc struct {
		Ver    string  `json:"jsonrpc"`
		Method string  `json:"method"`
		Params *string `json:"params"`
		Id     int     `json:"id"`
	}
	BulkParamChild struct {
		Rid  string        `json:"rid"`
		Json *BulkParamRpc `json:"json"`
	}
	BulkParam struct {
		Requests []*BulkParamChild `json:"requests"`
		Id       int               `json:"id"`
	}

	CallResult struct {
		Caption string
		Value   float64
	}
)

func doCall(endpoint, username, password string) []*CallResult {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	rpcClient := jsonrpc.NewClientWithOpts("https://"+endpoint+"/bulk",
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
			},
		})

	ids := map[int]string{1: "ApparentPower", 2: "ActivePower"}

	apparentPower := BulkParamChild{
		Rid:  "/tfwopaque/sensors.NumericSensor:4.0.3/I0ApparentPower",
		Json: &BulkParamRpc{Ver: "2.0", Method: "getReading", Params: nil, Id: 1},
	}

	activePower := BulkParamChild{
		Rid:  "/tfwopaque/sensors.NumericSensor:4.0.3/I0ActivePower",
		Json: &BulkParamRpc{Ver: "2.0", Method: "getReading", Params: nil, Id: 2},
	}

	response, err := rpcClient.Call("performBulk", &BulkParam{Id: 3, Requests: []*BulkParamChild{&apparentPower, &activePower}})
	if err != nil {
		log.Fatalln(err)
	}

	var res Result
	err = response.GetObject(&res)
	if err != nil {
		log.Fatalln(err)
	}
	// log.Printf("%#v", res)

	var ret []*CallResult

	for i := range res.Responses {
		ret = append(ret, &CallResult{
			Caption: ids[res.Responses[i].Json.Id],
			Value:   res.Responses[i].Json.Result.Value.Value,
		})
		// log.Printf("%s", ids[res.Responses[i].Json.Id])
		// log.Printf("%#v", res.Responses[i].Json.Result.Value.Value)
	}
	return ret
}

func main() {
	log.Printf("%#v", doCall("192.0.2.1", "***********", "************"))
}
