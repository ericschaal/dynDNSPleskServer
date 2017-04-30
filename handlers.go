package main

import (
	"encoding/json"
	"net/http"
	"log"
	"bytes"
	"io/ioutil"
	"encoding/xml"
	"strings"
)

type packet_struct struct {
	Token string
	Ip string
}

type data struct {
	Chihost *host `xml:" host,omitempty" json:"host,omitempty"`
	Chiopt *opt `xml:" opt,omitempty" json:"opt,omitempty"`
	Chisite_id *site_id `xml:" site-id,omitempty" json:"site-id,omitempty"`
	Chitype *dns_type `xml:" type,omitempty" json:"type,omitempty"`
	Chivalue *value `xml:" value,omitempty" json:"value,omitempty"`
}

type dns struct {
	Chiget_rec *get_rec `xml:" get_rec,omitempty" json:"get_rec,omitempty"`
}

type get_rec struct {
	Chiresult []*result `xml:" result,omitempty" json:"result,omitempty"`
}

type host struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type id struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type opt struct {
}

type packet struct {
	Chidns *dns `xml:" dns,omitempty" json:"dns,omitempty"`
}

type result struct {
	Chidata *data `xml:" data,omitempty" json:"data,omitempty"`
	Chiid *id `xml:" id,omitempty" json:"id,omitempty"`
	Chistatus *status `xml:" status,omitempty" json:"status,omitempty"`
}

type root struct {
	Chipacket *packet `xml:" packet,omitempty" json:"packet,omitempty"`
}

type site_id struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type status struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type dns_type struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type value struct {
	Text string `xml:",chardata" json:",omitempty"`
}



func Index(w http.ResponseWriter, r *http.Request) {
	var packet packet_struct
	if (r.Body == nil) {
		http.Error(w, "Please send a request body", 400)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&packet)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if (packet.Token != config.Token) {
		http.Error(w, "Please send a request body", 400)
		return
	}

	log.Println("Received ip address: " + packet.Ip)

	updatePlesk(packet.Ip)

}

func updatePlesk(receivedIP string) {

	var recordedIP string
	var recordedID string
	var packet packet
	var body string

	client := &http.Client{}

	// Getting dns entry list
	body = "<packet> <dns> <get_rec> <filter><site-id>" + config.PleskTargetSiteId + "</site-id></filter> </get_rec> </dns> </packet>"
	req, _ := http.NewRequest("POST", config.PleskServer + ":" +  config.PleskPort + "/enterprise/control/agent.php", bytes.NewBuffer([]byte(body)))
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("HTTP_AUTH_LOGIN",config.PleskLogin)
	req.Header.Add("HTTP_AUTH_PASSWD", config.PleskPassword)
	res, _ := client.Do(req)


	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errlog.Println("Error serializing client request. Dropping")
		return
	}

	xml.Unmarshal(resBody, &packet)

	for _,result := range packet.Chidns.Chiget_rec.Chiresult {
		if result.Chidata.Chihost.Text == config.PleskTargetDNSHost {
			recordedIP = result.Chidata.Chivalue.Text
			recordedID = result.Chiid.Text
		}
	}

	if recordedIP == "" || recordedID == "" {
		errlog.Println("No entry in DNS corresponding to home.sigmamu.me")
		return
	}



	if strings.TrimRight(recordedIP, "\n") != strings.TrimRight(receivedIP, "\n") {

		log.Println("Updating record...")

		body = "<packet> <dns> <del_rec> <filter> <id>" + recordedID + "</id> </filter> </del_rec> </dns> </packet>"
		req, _ := http.NewRequest("POST", config.PleskServer + ":" +  config.PleskPort + "/enterprise/control/agent.php", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("HTTP_AUTH_LOGIN",config.PleskLogin)
		req.Header.Add("HTTP_AUTH_PASSWD", config.PleskPassword)
		_, err := client.Do(req)

		if err != nil {
			errlog.Println("Error deleting old record. Abording")
			return
		}

		body = "<packet> <dns> <add_rec> <site-id>" + config.PleskTargetSiteId + "</site-id> <type>A</type> <host>home</host> <value>" + receivedIP + "</value> </add_rec> </dns> </packet>"
		req, _ = http.NewRequest("POST", config.PleskServer + ":" +  config.PleskPort + "/enterprise/control/agent.php", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("HTTP_AUTH_LOGIN",config.PleskLogin)
		req.Header.Add("HTTP_AUTH_PASSWD", config.PleskPassword)
		_, err = client.Do(req)

		if err != nil {
			errlog.Println("Error adding new record. Abording")
			return
		}

		stdlog.Println("Record updated with success.")

	}


}



